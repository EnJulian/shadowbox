# Use Python 3.11 slim image as base
FROM python:3.11-slim

# Set environment variables
ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    DEBIAN_FRONTEND=noninteractive

# Install system dependencies
RUN apt-get update && apt-get install -y \
    ffmpeg \
    aria2 \
    curl \
    wget \
    git \
    build-essential \
    pkg-config \
    libffi-dev \
    libssl-dev \
    && rm -rf /var/lib/apt/lists/*

# Install yt-dlp (latest version) - this is critical for downloading
RUN pip install --no-cache-dir --upgrade yt-dlp

# Create app directory
WORKDIR /app

# Create a non-root user
RUN useradd --create-home --shell /bin/bash appuser && \
    chown -R appuser:appuser /app

# Copy requirements first for better caching
COPY setup/requirements.txt /app/setup/
RUN pip install --no-cache-dir --upgrade pip && \
    pip install --no-cache-dir -r setup/requirements.txt

# Copy the entire project
COPY . /app/

# Create necessary directories
RUN mkdir -p /app/downloads /app/music && \
    chown -R appuser:appuser /app

# Make scripts executable
RUN find /app -name "*.sh" -type f -exec chmod +x {} \; && \
    find /app -name "*.py" -type f -exec chmod +x {} \;

# Switch to non-root user
USER appuser

# Set Python path
ENV PYTHONPATH="/app:$PYTHONPATH"

# Create volume mount points
VOLUME ["/app/downloads", "/app/music"]

# Expose port for potential web interface (if added later)
EXPOSE 8000

# Default command - interactive mode
CMD ["python", "/app/core/shbox.py"]