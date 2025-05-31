#!/bin/bash

# Utility script to enhance multiple music files with Spotify metadata

# Get the directory where the script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Check if virtual environment exists and activate it
if [ -d "$PROJECT_ROOT/.venv" ]; then
    source "$PROJECT_ROOT/.venv/bin/activate"
    PYTHON="$PROJECT_ROOT/.venv/bin/python"
else
    # Try to use system Python if venv doesn't exist
    PYTHON=$(which python3 || which python)
    if [ -z "$PYTHON" ]; then
        echo "❌ Python not found. Please install Python or run the setup script first."
        exit 1
    fi
fi

# Function to show usage
show_usage() {
    echo "Usage: enhance_files.sh [OPTIONS] <path>"
    echo ""
    echo "Options:"
    echo "  -h, --help                 Show this help message"
    echo "  -r, --recursive            Recursively scan directories"
    echo "  -e, --extensions EXT       Comma-separated list of file extensions to process (default: .mp3,.opus)"
    echo "  -d, --dry-run              Show what would be done without making changes"
    echo ""
    echo "Examples:"
    echo "  enhance_files.sh ~/Music"
    echo "  enhance_files.sh -r ~/Music"
    echo "  enhance_files.sh -e .mp3,.flac,.opus ~/Music"
    echo "  enhance_files.sh -d ~/Music/Artist"
    echo ""
}

# Check if arguments were provided
if [ $# -eq 0 ]; then
    show_usage
    exit 1
fi

# Check for help flag
if [[ "$1" == "-h" || "$1" == "--help" ]]; then
    show_usage
    exit 0
fi

# Set PYTHONPATH to include the project root
export PYTHONPATH="$PROJECT_ROOT:$PYTHONPATH"

# Run the Python script with all arguments
"$PYTHON" "$SCRIPT_DIR/enhance_files.py" "$@"