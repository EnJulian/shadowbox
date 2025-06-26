#!/bin/bash

# Shadowbox Docker Runner Script
# This script provides easy commands to run Shadowbox with Docker

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo -e "${BLUE}$1${NC}"
}

# Function to check if Docker is running
check_docker() {
    if ! docker info >/dev/null 2>&1; then
        print_error "Docker is not running. Please start Docker and try again."
        exit 1
    fi
}

# Function to create necessary directories
create_directories() {
    print_status "Creating necessary directories..."
    mkdir -p downloads music config
    print_status "Directories created: downloads/, music/, config/"
}

# Function to setup environment file
setup_env() {
    if [ ! -f .env ]; then
        print_status "Creating .env file from template..."
        cp .env.example .env
        print_warning "Please edit .env file with your API credentials for enhanced features"
        print_status "You can run: nano .env"
    else
        print_status ".env file already exists"
    fi
}

# Function to build the Docker image
build_image() {
    print_status "Building Shadowbox Docker image..."
    docker-compose build
    print_status "Docker image built successfully"
}

# Function to rebuild the Docker image (no cache)
rebuild_image() {
    print_status "Rebuilding Shadowbox Docker image (no cache)..."
    docker-compose build --no-cache
    print_status "Docker image rebuilt successfully"
}

# Function to run interactive mode
run_interactive() {
    print_header "Starting Shadowbox in Interactive Mode"
    print_status "Use Ctrl+C to exit"
    print_status "Ensuring proper terminal interaction..."
    
    # Use docker-compose run instead of up for better interactive support
    docker-compose run --rm shadowbox python /app/core/shbox.py
}

# Function to run interactive mode with docker-compose up (alternative)
run_interactive_up() {
    print_header "Starting Shadowbox in Interactive Mode (docker-compose up)"
    print_status "Use Ctrl+C to exit"
    print_warning "If input doesn't work, try './docker-run.sh run' instead"
    docker-compose up
}

# Function to run CLI mode
run_cli() {
    if [ -z "$1" ]; then
        print_error "Please provide a search query"
        echo "Usage: $0 cli \"Artist - Song Title\""
        exit 1
    fi
    
    print_header "Running Shadowbox CLI Mode"
    print_status "Searching for: $1"
    docker-compose run --rm shadowbox python /app/core/main.py -q "$1"
}

# Function to run in background
run_background() {
    print_header "Starting Shadowbox in Background Mode"
    docker-compose up -d
    print_status "Shadowbox is running in background"
    print_status "Use '$0 logs' to view logs"
    print_status "Use '$0 stop' to stop the container"
}

# Function to show logs
show_logs() {
    print_header "Shadowbox Logs"
    docker-compose logs -f
}

# Function to stop containers
stop_containers() {
    print_status "Stopping Shadowbox containers..."
    docker-compose down
    print_status "Containers stopped"
}

# Function to clean up everything
cleanup() {
    print_warning "This will remove all containers, images, and volumes"
    read -p "Are you sure? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        print_status "Cleaning up..."
        docker-compose down -v --rmi all
        print_status "Cleanup complete"
    else
        print_status "Cleanup cancelled"
    fi
}

# Function to update yt-dlp
update_ytdlp() {
    print_status "Updating yt-dlp..."
    docker-compose exec shadowbox pip install --upgrade yt-dlp
    print_status "yt-dlp updated"
}

# Function to check dependencies
check_deps() {
    print_header "Checking Dependencies in Container"
    if docker-compose ps | grep -q "shadowbox.*Up"; then
        docker-compose exec shadowbox ./docker-deps-check.sh
    else
        print_status "Starting temporary container for dependency check..."
        docker-compose run --rm shadowbox ./docker-deps-check.sh
    fi
}

# Function to show help
show_help() {
    print_header "Shadowbox Docker Runner"
    echo
    echo "Usage: $0 [command] [options]"
    echo
    echo "Commands:"
    echo "  setup          - Initial setup (create directories, .env file)"
    echo "  build          - Build the Docker image"
    echo "  rebuild        - Rebuild the Docker image (no cache)"
    echo "  run            - Run in interactive mode (recommended)"
    echo "  run-up         - Run in interactive mode (docker-compose up)"
    echo "  cli \"query\"    - Run single download via CLI"
    echo "  background     - Run in background mode"
    echo "  logs           - Show container logs"
    echo "  stop           - Stop running containers"
    echo "  update         - Update yt-dlp in running container"
    echo "  check          - Check all dependencies in container"
    echo "  cleanup        - Remove all containers, images, and volumes"
    echo "  shell          - Open shell in container"
    echo "  help           - Show this help message"
    echo
    echo "Examples:"
    echo "  $0 setup                           # Initial setup"
    echo "  $0 run                             # Interactive mode"
    echo "  $0 cli \"Imagine Dragons Believer\"  # Download specific song"
    echo "  $0 background                      # Run in background"
    echo
}

# Function to open shell in container
open_shell() {
    print_header "Opening Shell in Shadowbox Container"
    if docker-compose ps | grep -q "shadowbox.*Up"; then
        docker-compose exec shadowbox bash
    else
        print_status "Starting temporary container..."
        docker-compose run --rm shadowbox bash
    fi
}

# Main script logic
main() {
    check_docker
    
    case "${1:-run}" in
        "setup")
            create_directories
            setup_env
            build_image
            print_status "Setup complete! Run '$0 run' to start Shadowbox"
            ;;
        "build")
            build_image
            ;;
        "rebuild")
            rebuild_image
            ;;
        "run")
            create_directories
            run_interactive
            ;;
        "run-up")
            create_directories
            run_interactive_up
            ;;
        "cli")
            create_directories
            run_cli "$2"
            ;;
        "background")
            create_directories
            run_background
            ;;
        "logs")
            show_logs
            ;;
        "stop")
            stop_containers
            ;;
        "update")
            update_ytdlp
            ;;
        "check")
            check_deps
            ;;
        "cleanup")
            cleanup
            ;;
        "shell")
            create_directories
            open_shell
            ;;
        "help"|"-h"|"--help")
            show_help
            ;;
        *)
            print_error "Unknown command: $1"
            show_help
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"