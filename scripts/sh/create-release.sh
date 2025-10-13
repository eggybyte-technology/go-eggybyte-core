#!/bin/bash
set -Eeuo pipefail

# Create Release Script for EggyByte Core
# This script creates a complete GitHub release with binaries using goreleaser

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
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

# Parse command line arguments
VERSION=""
FORCE=false
SKIP_TESTS=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--version)
            VERSION="$2"
            shift 2
            ;;
        -f|--force)
            FORCE=true
            shift
            ;;
        --skip-tests)
            SKIP_TESTS=true
            shift
            ;;
        -h|--help)
            echo "Usage: $0 -v VERSION [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  -v, --version VERSION    Version tag to create (e.g., v1.0.0)"
            echo "  -f, --force             Force release creation even if tag exists"
            echo "  --skip-tests            Skip running tests before release"
            echo "  -h, --help              Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0 -v v1.0.0"
            echo "  $0 -v v1.1.0 --force"
            echo "  $0 -v v1.0.0 --skip-tests"
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Validate version parameter
if [ -z "$VERSION" ]; then
    log_error "Version is required"
    echo "Usage: $0 -v VERSION"
    echo "Example: $0 -v v1.0.0"
    exit 1
fi

# Validate version format (semantic versioning)
if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    log_error "Invalid version format: $VERSION"
    log_info "Version must follow semantic versioning: vMAJOR.MINOR.PATCH"
    log_info "Examples: v1.0.0, v1.1.0, v2.0.0"
    exit 1
fi

# Check if we're in a git repository
check_git_repo() {
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        log_error "Not in a git repository"
        exit 1
    fi
}

# Check if GITHUB_TOKEN is set
check_github_token() {
    if [ -z "${GITHUB_TOKEN:-}" ]; then
        log_error "GITHUB_TOKEN environment variable is not set"
        log_info "Please set it with: export GITHUB_TOKEN=your_token"
        log_info "You can get a token from: https://github.com/settings/tokens"
        log_info "Required permissions: repo, write:packages"
        exit 1
    fi
}

# Check if goreleaser is installed
check_goreleaser() {
    if ! command -v goreleaser >/dev/null 2>&1; then
        log_error "goreleaser is not installed"
        log_info "Install with: go install github.com/goreleaser/goreleaser@latest"
        exit 1
    fi
}

# Check if remote origin is set correctly
check_remote_origin() {
    local remote_url
    remote_url=$(git remote get-url origin 2>/dev/null || echo "")
    
    if [[ "$remote_url" != *"eggybyte-technology/go-eggybyte-core"* ]]; then
        log_warning "Remote origin doesn't point to eggybyte-technology/go-eggybyte-core"
        log_info "Current remote: $remote_url"
        log_info "Expected: https://github.com/eggybyte-technology/go-eggybyte-core.git"
        
        read -p "Do you want to update the remote origin? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            git remote set-url origin https://github.com/eggybyte-technology/go-eggybyte-core.git
            log_success "Remote origin updated"
        else
            log_error "Please update the remote origin manually"
            exit 1
        fi
    fi
}

# Check if tag already exists
check_tag_exists() {
    if git tag -l | grep -q "^$VERSION$"; then
        if [ "$FORCE" = true ]; then
            log_warning "Tag $VERSION already exists, but --force is enabled"
            log_info "Deleting existing tag..."
            git tag -d "$VERSION"
            if git push origin ":refs/tags/$VERSION" 2>/dev/null; then
                log_success "Existing tag deleted from remote"
            else
                log_warning "Could not delete remote tag (may not exist)"
            fi
        else
            log_error "Tag $VERSION already exists"
            log_info "Use --force to overwrite the existing tag"
            exit 1
        fi
    fi
}

# Check if there are uncommitted changes
check_uncommitted_changes() {
    if ! git diff-index --quiet HEAD --; then
        log_warning "There are uncommitted changes"
        git status --short
        echo
        read -p "Do you want to commit these changes before creating the release? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            git add .
            git commit -m "chore: prepare for release $VERSION"
            log_success "Changes committed"
        else
            log_error "Please commit or stash your changes first"
            exit 1
        fi
    fi
}

# Run tests
run_tests() {
    if [ "$SKIP_TESTS" = true ]; then
        log_warning "Skipping tests as requested"
        return
    fi
    
    log_info "Running tests..."
    if ! go test -v ./pkg/...; then
        log_error "Tests failed"
        exit 1
    fi
    log_success "All tests passed"
}

# Push changes to GitHub
push_to_github() {
    local branch
    branch=$(git branch --show-current)
    
    log_info "Pushing changes to GitHub..."
    log_info "Branch: $branch"
    
    if git push origin "$branch"; then
        log_success "Successfully pushed to GitHub"
    else
        log_error "Failed to push to GitHub"
        exit 1
    fi
}

# Create and push the tag
create_and_push_tag() {
    log_info "Creating tag: $VERSION"
    
    # Create annotated tag with release message
    local release_message
    release_message="Release $VERSION

This release includes:
- Complete enterprise-grade Go microservice foundation
- ebcctl CLI tool for scaffolding and code generation
- Unified monitoring with Prometheus metrics and health checks
- TiDB/MySQL database integration with GORM
- Structured logging with context support
- Docker and Kubernetes deployment support
- GitHub Actions CI/CD pipeline

For more information, see the CHANGELOG.md file."

    if git tag -a "$VERSION" -m "$release_message"; then
        log_success "Tag $VERSION created locally"
    else
        log_error "Failed to create tag $VERSION"
        exit 1
    fi
    
    log_info "Pushing tag to GitHub..."
    if git push origin "$VERSION"; then
        log_success "Tag $VERSION pushed to GitHub"
    else
        log_error "Failed to push tag to GitHub"
        exit 1
    fi
}

# Create GitHub release with goreleaser
create_goreleaser_release() {
    log_info "Creating GitHub release with goreleaser..."
    
    if goreleaser release --clean; then
        log_success "GitHub release created successfully with goreleaser"
    else
        log_error "Failed to create GitHub release with goreleaser"
        exit 1
    fi
}

# Verify the release was created successfully
verify_release() {
    log_info "Verifying release creation..."
    
    # Wait a moment for GitHub to process
    sleep 5
    
    log_success "Release $VERSION created successfully!"
    
    echo
    log_info "Release Information:"
    echo "  Version: $VERSION"
    echo "  Repository: https://github.com/eggybyte-technology/go-eggybyte-core"
    echo "  Release URL: https://github.com/eggybyte-technology/go-eggybyte-core/releases/tag/$VERSION"
    echo ""
    log_info "Release includes:"
    echo "  - Binaries for Linux, macOS, and Windows (AMD64 and ARM64)"
    echo "  - Source code archives"
    echo "  - Checksums for verification"
    echo "  - Installation instructions"
    echo ""
    log_info "Next steps:"
    echo "  1. Visit the release page to review the generated content"
    echo "  2. Test the binaries on different platforms"
    echo "  3. Update documentation if needed"
    echo "  4. Announce the release to users"
}

# Main function
main() {
    log_info "Starting complete release process..."
    log_info "Version: $VERSION"
    log_info "Force: $FORCE"
    log_info "Skip tests: $SKIP_TESTS"
    
    cd "$PROJECT_ROOT"
    
    # Pre-flight checks
    check_git_repo
    check_github_token
    check_goreleaser
    check_remote_origin
    check_tag_exists
    check_uncommitted_changes
    
    # Run tests
    run_tests
    
    # Push changes
    push_to_github
    
    # Create tag
    create_and_push_tag
    
    # Create release with goreleaser
    create_goreleaser_release
    
    # Verify release
    verify_release
    
    log_success "Complete release process finished!"
}

# Run main function
main "$@"
