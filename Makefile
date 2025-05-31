.PHONY: help build clean install test release

help:
	@echo "Shadowbox Build Commands:"
	@echo "  build     - Build executable for current platform"
	@echo "  clean     - Clean build artifacts"
	@echo "  install   - Install dependencies"
	@echo "  test      - Run tests"
	@echo "  release   - Create a new release (interactive)"
	@echo "  help      - Show this help message"

build:
	python build.py

clean:
	rm -rf build/ dist/ *.spec __pycache__/ *.pyc
	find . -name "*.pyc" -delete
	find . -name "__pycache__" -delete

install:
	pip install -r setup/requirements.txt
	pip install pyinstaller

test:
	python -m pytest tests/ -v

release:
	python release.py

# Platform-specific builds (for development)
build-linux:
	pyinstaller shadowbox.spec

build-windows:
	pyinstaller shadowbox.spec

build-macos:
	pyinstaller shadowbox.spec