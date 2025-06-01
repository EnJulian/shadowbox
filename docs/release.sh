#!/bin/bash

# Wrapper script to run release.py from the docs folder
# This script changes to the project root, activates the virtual environment,
# and runs the release script with the provided arguments

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Navigate to project root (one level up from docs)
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "Navigating to project root: $PROJECT_ROOT"
cd "$PROJECT_ROOT"

# Activate virtual environment
echo "Activating virtual environment..."
source .venv/bin/activate

# Run release.py with all provided arguments
echo "Running release.py $@"
python3 release.py "$@"