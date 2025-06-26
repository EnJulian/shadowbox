#!/bin/bash

# Docker Setup Validation Script
# This script tests the Docker setup to ensure everything works correctly

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${GREEN}[TEST]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_header() {
    echo -e "${BLUE}=== $1 ===${NC}"
}

# Function to check if Docker is available
check_docker() {
    print_header "Docker Environment Check"
    
    if ! command -v docker >/dev/null 2>&1; then
        print_error "Docker is not installed or not in PATH"
        exit 1
    fi
    
    if ! docker info >/dev/null 2>&1; then
        print_error "Docker is not running"
        exit 1
    fi
    
    if ! command -v docker-compose >/dev/null 2>&1; then
        print_error "Docker Compose is not installed or not in PATH"
        exit 1
    fi
    
    print_status "Docker and Docker Compose are available"
}

# Function to test file structure
check_files() {
    print_header "File Structure Check"
    
    local required_files=(
        "Dockerfile"
        "docker-compose.yml"
        ".env.example"
        ".dockerignore"
        "docker-run.sh"
        "docker-run.bat"
        "docker-deps-check.sh"
        "DOCKER.md"
    )
    
    for file in "${required_files[@]}"; do
        if [ -f "$file" ]; then
            print_status "Found: $file"
        else
            print_error "Missing: $file"
            exit 1
        fi
    done
    
    # Check if scripts are executable
    if [ -x "docker-run.sh" ]; then
        print_status "docker-run.sh is executable"
    else
        print_warning "docker-run.sh is not executable, fixing..."
        chmod +x docker-run.sh
    fi
    
    if [ -x "docker-deps-check.sh" ]; then
        print_status "docker-deps-check.sh is executable"
    else
        print_warning "docker-deps-check.sh is not executable, fixing..."
        chmod +x docker-deps-check.sh
    fi
}

# Function to test Docker build
test_build() {
    print_header "Docker Build Test"
    
    print_status "Building Docker image (this may take a few minutes)..."
    if docker-compose build >/dev/null 2>&1; then
        print_status "Docker image built successfully"
    else
        print_error "Docker build failed"
        echo "Run 'docker-compose build' manually to see detailed error messages"
        exit 1
    fi
}

# Function to test basic container functionality
test_container() {
    print_header "Container Functionality Test"
    
    print_status "Testing container startup..."
    if docker-compose run --rm shadowbox python --version >/dev/null 2>&1; then
        print_status "Container starts and Python is available"
    else
        print_error "Container startup failed"
        exit 1
    fi
    
    print_status "Testing dependency checker..."
    if docker-compose run --rm shadowbox ./docker-deps-check.sh >/dev/null 2>&1; then
        print_status "Dependency checker passed"
    else
        print_warning "Dependency checker reported issues (run './docker-run.sh check' for details)"
    fi
    
    print_status "Testing core imports..."
    if docker-compose run --rm shadowbox python -c "
import sys
sys.path.insert(0, '/app')
from meta_ops.downloader import is_url
from meta_ops.metadata import extract_metadata
print('Core imports successful')
" >/dev/null 2>&1; then
        print_status "Core module imports work correctly"
    else
        print_error "Core module imports failed"
        exit 1
    fi
}

# Function to test helper scripts
test_scripts() {
    print_header "Helper Scripts Test"
    
    # Test help command
    if ./docker-run.sh help >/dev/null 2>&1; then
        print_status "docker-run.sh help command works"
    else
        print_error "docker-run.sh help command failed"
        exit 1
    fi
    
    # Test that all commands are recognized
    local commands=("setup" "build" "rebuild" "run" "cli" "background" "logs" "stop" "update" "check" "cleanup" "shell")
    
    for cmd in "${commands[@]}"; do
        # Just check that the command is recognized (doesn't execute)
        if ./docker-run.sh help | grep -q "$cmd"; then
            print_status "Command '$cmd' is documented"
        else
            print_warning "Command '$cmd' may not be properly documented"
        fi
    done
}

# Function to test environment setup
test_environment() {
    print_header "Environment Setup Test"
    
    # Create test directories
    mkdir -p downloads music config
    print_status "Created test directories"
    
    # Test .env file creation
    if [ ! -f ".env" ]; then
        cp .env.example .env
        print_status "Created .env file from template"
    else
        print_status ".env file already exists"
    fi
    
    # Test volume mounts
    if docker-compose run --rm shadowbox ls /app/downloads >/dev/null 2>&1; then
        print_status "Volume mounts are working"
    else
        print_error "Volume mounts failed"
        exit 1
    fi
}

# Function to cleanup test artifacts
cleanup() {
    print_header "Cleanup"
    
    print_status "Stopping any running containers..."
    docker-compose down >/dev/null 2>&1 || true
    
    print_status "Test completed successfully!"
    echo
    echo "Your Docker setup is ready to use. You can now run:"
    echo "  ./docker-run.sh run          # Start interactive mode"
    echo "  ./docker-run.sh check        # Verify all dependencies"
    echo "  ./docker-run.sh help         # See all available commands"
}

# Main test execution
main() {
    echo "Starting Docker setup validation..."
    echo
    
    check_docker
    echo
    
    check_files
    echo
    
    test_build
    echo
    
    test_container
    echo
    
    test_scripts
    echo
    
    test_environment
    echo
    
    cleanup
}

# Handle script interruption
trap 'echo; print_error "Test interrupted"; docker-compose down >/dev/null 2>&1 || true; exit 1' INT TERM

# Run main function
main "$@"