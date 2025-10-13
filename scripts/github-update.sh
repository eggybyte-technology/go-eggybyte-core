#!/bin/bash
set -Eeuo pipefail

# GitHub Repository Update Script for EggyByte Core
# This script updates the GitHub repository with the latest changes

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

# Check if there are uncommitted changes
check_uncommitted_changes() {
    if ! git diff-index --quiet HEAD --; then
        log_warning "There are uncommitted changes"
        git status --short
        echo
        read -p "Do you want to commit these changes? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            git add .
            git commit -m "chore: update repository structure and dependencies"
            log_success "Changes committed"
        else
            log_error "Please commit or stash your changes first"
            exit 1
        fi
    fi
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

# Verify the push was successful
verify_push() {
    log_info "Verifying push..."
    
    # Wait a moment for GitHub to process
    sleep 2
    
    # Check if the latest commit is on GitHub
    local latest_commit
    latest_commit=$(git rev-parse HEAD)
    
    log_info "Latest commit: $latest_commit"
    log_success "GitHub repository updated successfully!"
    
    echo
    log_info "Next steps:"
    echo "  1. Visit: https://github.com/eggybyte-technology/go-eggybyte-core"
    echo "  2. Check the Actions tab for CI/CD status"
    echo "  3. Create a release tag if needed: make github-release VERSION=v1.0.0"
}

# Main function
main() {
    log_info "Starting GitHub repository update..."
    
    cd "$PROJECT_ROOT"
    
    check_git_repo
    check_github_token
    check_remote_origin
    check_uncommitted_changes
    push_to_github
    verify_push
    
    log_success "GitHub repository update completed!"
}

# Run main function
main "$@"
