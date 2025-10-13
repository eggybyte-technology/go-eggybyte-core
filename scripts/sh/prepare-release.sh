#!/bin/bash
set -Eeuo pipefail

# Prepare Release Script for EggyByte Core
# This script prepares the project for a new release

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if we're in a git repository
check_git_repo() {
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        log_error "Not in a git repository"
        exit 1
    fi
}

# Check if working directory is clean
check_clean_working_dir() {
    if ! git diff-index --quiet HEAD --; then
        log_error "Working directory is not clean. Please commit or stash changes."
        exit 1
    fi
}

# Run tests
run_tests() {
    log_info "Running tests..."
    cd "$PROJECT_ROOT"
    
    if ! go test -v -race=false ./pkg/...; then
        log_error "Tests failed"
        exit 1
    fi
    
    log_success "All tests passed"
}

# Run linting
run_linting() {
    log_info "Running linting..."
    cd "$PROJECT_ROOT"
    
    if command -v golangci-lint >/dev/null 2>&1; then
        if ! golangci-lint run; then
            log_error "Linting failed"
            exit 1
        fi
        log_success "Linting passed"
    else
        log_warning "golangci-lint not installed, skipping linting"
    fi
}

# Build the project
build_project() {
    log_info "Building project..."
    cd "$PROJECT_ROOT"
    
    if ! go build -v ./...; then
        log_error "Build failed"
        exit 1
    fi
    
    log_success "Build successful"
}

# Update version information
update_version() {
    local version="$1"
    log_info "Updating version to $version..."
    
    # Update CHANGELOG.md
    if [ -f "CHANGELOG.md" ]; then
        # Move unreleased changes to version section
        sed -i.bak "s/## \[Unreleased\]/## \[$version\] - $(date +%Y-%m-%d)/" CHANGELOG.md
        rm -f CHANGELOG.md.bak
        log_success "Updated CHANGELOG.md"
    fi
    
    # Update README.md version badge
    if [ -f "README.md" ]; then
        sed -i.bak "s/v[0-9]\+\.[0-9]\+\.[0-9]\+/$version/g" README.md
        rm -f README.md.bak
        log_success "Updated README.md version badge"
    fi
}

# Create git tag
create_tag() {
    local version="$1"
    log_info "Creating git tag $version..."
    
    if git tag -l | grep -q "^$version$"; then
        log_warning "Tag $version already exists"
        return
    fi
    
    git tag -a "$version" -m "Release $version"
    log_success "Created tag $version"
}

# Main function
main() {
    local version="${1:-}"
    
    if [ -z "$version" ]; then
        log_error "Usage: $0 <version>"
        log_error "Example: $0 v1.0.0"
        exit 1
    fi
    
    # Validate version format
    if [[ ! "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        log_error "Invalid version format. Use semantic versioning (e.g., v1.0.0)"
        exit 1
    fi
    
    log_info "Preparing release $version..."
    
    cd "$PROJECT_ROOT"
    
    # Pre-flight checks
    check_git_repo
    check_clean_working_dir
    
    # Run quality checks
    run_tests
    run_linting
    build_project
    
    # Update version information
    update_version "$version"
    
    # Create git tag
    create_tag "$version"
    
    log_success "Release $version prepared successfully!"
    log_info "Next steps:"
    log_info "1. Review changes: git diff"
    log_info "2. Commit changes: git add . && git commit -m \"Prepare release $version\""
    log_info "3. Push tag: git push origin $version"
    log_info "4. Create release: make github-release VERSION=$version"
}

# Run main function with all arguments
main "$@"
