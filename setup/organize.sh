#!/bin/bash

# Script to organize the project structure and create symbolic links
# Run this script after cloning the repository to set up the project structure

echo "🔧 Setting up project structure..."

# Get the project root directory
PROJECT_ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." &> /dev/null && pwd )"
cd "$PROJECT_ROOT"

# Create directories if they don't exist
mkdir -p core exec_bin meta_ops tests

# Move files to appropriate directories
# Only move if the files exist and are not already in their target directories

# Core files
if [ -f "lsfoobar.py" ] && [ ! -f "core/lsfoobar.py" ]; then
    mv lsfoobar.py core/
fi
if [ -f "main.py" ] && [ ! -f "core/main.py" ]; then
    mv main.py core/
fi

# Meta operations files
if [ -f "metadata.py" ] && [ ! -f "meta_ops/metadata.py" ]; then
    mv metadata.py meta_ops/
fi
if [ -f "cover_art.py" ] && [ ! -f "meta_ops/cover_art.py" ]; then
    mv cover_art.py meta_ops/
fi
if [ -f "downloader.py" ] && [ ! -f "meta_ops/downloader.py" ]; then
    mv downloader.py meta_ops/
fi

# Executable scripts
if [ -f "lsfoobar.sh" ] && [ ! -f "exec_bin/lsfoobar.sh" ] && [ ! -L "lsfoobar.sh" ]; then
    mv lsfoobar.sh exec_bin/
fi
if [ -f "lsfoobar.bat" ] && [ ! -f "exec_bin/lsfoobar.bat" ] && [ ! -L "lsfoobar.bat" ]; then
    mv lsfoobar.bat exec_bin/
fi
if [ -f "run.sh" ] && [ ! -f "exec_bin/run.sh" ] && [ ! -L "run.sh" ]; then
    mv run.sh exec_bin/
fi
if [ -f "run.bat" ] && [ ! -f "exec_bin/run.bat" ] && [ ! -L "run.bat" ]; then
    mv run.bat exec_bin/
fi

# Test files
if [ -f "test.py" ] && [ ! -f "tests/test.py" ]; then
    mv test.py tests/
fi

# Setup files
if [ -f "requirements.txt" ] && [ ! -f "setup/requirements.txt" ]; then
    mv requirements.txt setup/
fi
if [ -f "setup.py" ] && [ ! -f "setup/setup.py" ]; then
    mv setup.py setup/
fi
if [ -f "install.sh" ] && [ ! -f "setup/install.sh" ]; then
    mv install.sh setup/
fi
if [ -f "install.bat" ] && [ ! -f "setup/install.bat" ]; then
    mv install.bat setup/
fi

# Make scripts executable
chmod +x exec_bin/*.sh setup/*.sh
chmod +x core/*.py
chmod +x *.sh
chmod +x run.sh lsfoobar.sh

# Create symbolic links in the project root
ln -sf exec_bin/lsfoobar.sh lsfoobar.sh
ln -sf exec_bin/run.sh run.sh
ln -sf exec_bin/lsfoobar.bat lsfoobar.bat
ln -sf exec_bin/run.bat run.bat

echo "✅ Project structure set up successfully!"
echo "You can now run the application using:"
echo "./lsfoobar.sh or ./run.sh"