# Shadowbox Tests

This directory contains all test files for the Shadowbox project.

## Test Files

- **`test.py`** - Main test suite for core functionality (cover art, metadata)
- **`test_date.py`** - Tests for date extraction functionality
- **`test_spotify_search.py`** - Tests for Spotify metadata search functionality
- **`test_enhanced_metadata.py`** - Tests for enhanced metadata fields
- **`run_all_tests.py`** - Test runner that executes all tests

## Running Tests

### Run All Tests
```bash
python3 tests/run_all_tests.py
```

### Run Individual Tests
```bash
python3 tests/test_date.py
python3 tests/test_spotify_search.py
python3 tests/test_enhanced_metadata.py
python3 tests/test.py
```

## Dependencies

Some tests require additional dependencies:
- **Spotify tests**: Require `spotipy` package and Spotify API credentials
- **Metadata tests**: Require `mutagen` package
- **Cover art tests**: Require `requests` package

Set environment variables for Spotify functionality:
```bash
export SPOTIFY_CLIENT_ID="your_client_id"
export SPOTIFY_CLIENT_SECRET="your_client_secret"
```

## Test Structure

All tests use a consistent import pattern:
```python
import os
import sys

# Add the project root to the Python path
project_root = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
sys.path.insert(0, project_root)

# Import modules from the project
from meta_ops.module_name import function_name
```

This ensures tests can be run from any location and will correctly import the project modules.