#!/bin/bash
# Build script for EggyByte Core applications

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
SERVICE_NAME=""
BUILD_DIR="./build"
GO_VERSION="1.25.1"
TARGET_OS="linux"
TARGET_ARCH="amd64"
CGO_ENABLED="0"
LDFLAGS="-s -w"

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to show usage
show_usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Build EggyByte Core applications

OPTIONS:
    -s, --service NAME     Service name (required)
    -d, --dir DIR          Build directory (default: ./build)
    -o, --os OS            Target OS (default: linux)
    -a, --arch ARCH        Target architecture (default: amd64)
    -v, --version VERSION  Go version (default: 1.25.1)
    -h, --help             Show this help message

EXAMPLES:
    $0 -s user-service
    $0 -s payment-service -d ./dist -o darwin -a arm64
    $0 --service auth-service --dir ./bin

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -s|--service)
            SERVICE_NAME="$2"
            shift 2
            ;;
        -d|--dir)
            BUILD_DIR="$2"
            shift 2
            ;;
        -o|--os)
            TARGET_OS="$2"
            shift 2
            ;;
        -a|--arch)
            TARGET_ARCH="$2"
            shift 2
            ;;
        -v|--version)
            GO_VERSION="$2"
            shift 2
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Validate required parameters
if [[ -z "$SERVICE_NAME" ]]; then
    print_error "Service name is required"
    show_usage
    exit 1
fi

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed. Please install Go $GO_VERSION or later"
    exit 1
fi

# Check Go version
CURRENT_GO_VERSION=$(go version | cut -d' ' -f3 | sed 's/go//')
if [[ "$CURRENT_GO_VERSION" < "$GO_VERSION" ]]; then
    print_warning "Go version $CURRENT_GO_VERSION is older than recommended $GO_VERSION"
fi

print_info "Building $SERVICE_NAME for $TARGET_OS/$TARGET_ARCH"

# Create build directory
mkdir -p "$BUILD_DIR"

# Set environment variables
export GOOS="$TARGET_OS"
export GOARCH="$TARGET_ARCH"
export CGO_ENABLED="$CGO_ENABLED"

# Build the binary
BINARY_NAME="$SERVICE_NAME"
if [[ "$TARGET_OS" == "windows" ]]; then
    BINARY_NAME="${SERVICE_NAME}.exe"
fi

OUTPUT_PATH="$BUILD_DIR/$BINARY_NAME"

print_info "Compiling binary..."
if go build \
    -ldflags="$LDFLAGS" \
    -o "$OUTPUT_PATH" \
    "./cmd/$SERVICE_NAME"; then
    print_success "Binary built successfully: $OUTPUT_PATH"
else
    print_error "Build failed"
    exit 1
fi

# Show binary information
if [[ -f "$OUTPUT_PATH" ]]; then
    BINARY_SIZE=$(du -h "$OUTPUT_PATH" | cut -f1)
    print_info "Binary size: $BINARY_SIZE"
    
    # Show file type
    if command -v file &> /dev/null; then
        FILE_INFO=$(file "$OUTPUT_PATH")
        print_info "Binary info: $FILE_INFO"
    fi
fi

print_success "Build completed successfully!"
