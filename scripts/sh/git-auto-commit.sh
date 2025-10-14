#!/bin/bash
set -Eeuo pipefail

# Git Auto-Commit Script for EggyByte Project
# Automates git add, commit, and push operations with intelligent message generation
# Usage: ./scripts/git-auto-commit.sh [options] [commit_message] [description]

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
NC='\033[0m' # No Color

# Default values
COMMIT_MESSAGE=""
COMMIT_DESCRIPTION=""
INTERACTIVE_MODE=false
REVIEW_MODE=false
STATUS_MODE=false
SELECT_FILES=false
PUSH_CHANGES=true
DRY_RUN=false

# Function to print colored output
print_header() {
    local message="$1"
    local border_char="${2:-=}"
    local width="${3:-80}"
    
    echo -e "${MAGENTA}"
    printf "%*s\n" $width | tr ' ' "$border_char"
    printf "%*s\n" $(((${#message} + $width) / 2)) "$message"
    printf "%*s\n" $width | tr ' ' "$border_char"
    echo -e "${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_step() {
    echo -e "${CYAN}🔄 $1${NC}"
}

# Function to show usage
show_usage() {
    cat << EOF
Git Auto-Commit Script for EggyByte Project

USAGE:
    $0 [OPTIONS] [COMMIT_MESSAGE] [DESCRIPTION]

OPTIONS:
    -h, --help              Show this help message
    -i, --interactive        Interactive mode for complex commits
    -r, --review             Review changes before committing
    -s, --status             Show repository status only
    --select-files           Select specific files to commit
    --no-push                Don't push changes to remote
    --dry-run                Show what would be done without executing
    -v, --verbose            Verbose output

EXAMPLES:
    $0 "feat: add user authentication"
    $0 "fix: resolve database connection" "Fixed timeout issue in user service"
    $0 --interactive
    $0 --review
    $0 --status
    $0 --select-files

COMMIT MESSAGE FORMAT:
    <type>(<scope>): <description>
    
    Types: feat, fix, docs, style, refactor, test, chore, perf, ci
    Scope: Optional component or module name
    Description: Clear description of the change

EOF
}

# Function to check if we're in a git repository
check_git_repo() {
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        print_error "Not in a git repository"
        exit 1
    fi
}

# Function to get repository status
get_repo_status() {
    print_header "Repository Status"
    
    echo -e "${WHITE}Current Branch:${NC} $(git branch --show-current)"
    echo -e "${WHITE}Remote Status:${NC}"
    git status --porcelain
    echo
    
    if git diff --staged --quiet; then
        print_info "No staged changes"
    else
        print_info "Staged changes:"
        git diff --staged --name-only | sed 's/^/  /'
    fi
    
    if git diff --quiet; then
        print_info "No unstaged changes"
    else
        print_info "Unstaged changes:"
        git diff --name-only | sed 's/^/  /'
    fi
    
    if git diff --quiet HEAD; then
        print_info "Working directory clean"
    else
        print_warning "Working directory has changes"
    fi
}

# Function to review changes
review_changes() {
    print_header "Review Changes"
    
    if ! git diff --quiet; then
        print_info "Unstaged changes:"
        git diff --stat
        echo
        print_info "Detailed diff:"
        git diff
        echo
    fi
    
    if ! git diff --staged --quiet; then
        print_info "Staged changes:"
        git diff --staged --stat
        echo
        print_info "Detailed staged diff:"
        git diff --staged
        echo
    fi
}

# Function to select files interactively
select_files() {
    print_header "Select Files to Commit"
    
    local modified_files=($(git diff --name-only))
    local staged_files=($(git diff --staged --name-only))
    local all_files=($(printf '%s\n' "${modified_files[@]}" "${staged_files[@]}" | sort -u))
    
    if [ ${#all_files[@]} -eq 0 ]; then
        print_info "No modified files found"
        return 0
    fi
    
    echo "Available files:"
    for i in "${!all_files[@]}"; do
        echo "  $((i+1)). ${all_files[$i]}"
    done
    
    echo
    read -p "Enter file numbers to stage (comma-separated, or 'all'): " selection
    
    if [ "$selection" = "all" ]; then
        git add .
        print_success "All files staged"
    else
        IFS=',' read -ra INDICES <<< "$selection"
        for index in "${INDICES[@]}"; do
            index=$((index-1))
            if [ $index -ge 0 ] && [ $index -lt ${#all_files[@]} ]; then
                git add "${all_files[$index]}"
                print_success "Staged: ${all_files[$index]}"
            fi
        done
    fi
}

# Function to generate commit message from changes
generate_commit_message() {
    local staged_files=($(git diff --staged --name-only))
    local message=""
    
    if [ ${#staged_files[@]} -eq 0 ]; then
        echo "chore: update repository"
        return
    fi
    
    # Analyze file types and changes
    local has_go_files=false
    local has_docs=false
    local has_config=false
    local has_scripts=false
    
    for file in "${staged_files[@]}"; do
        case "$file" in
            *.go) has_go_files=true ;;
            *.md|*.txt|*.rst) has_docs=true ;;
            *.yml|*.yaml|*.json|*.toml|*.ini) has_config=true ;;
            *.sh|*.py|*.js) has_scripts=true ;;
        esac
    done
    
    # Determine commit type and message
    if [ "$has_go_files" = true ]; then
        if git diff --staged --name-only | grep -q "test"; then
            message="test: add or update tests"
        elif git diff --staged --name-only | grep -q "main.go\|cmd/"; then
            message="feat: implement new functionality"
        else
            message="refactor: improve code structure"
        fi
    elif [ "$has_docs" = true ]; then
        message="docs: update documentation"
    elif [ "$has_config" = true ]; then
        message="chore: update configuration"
    elif [ "$has_scripts" = true ]; then
        message="chore: update automation scripts"
    else
        message="chore: update project files"
    fi
    
    echo "$message"
}

# Function to commit changes
commit_changes() {
    local message="$1"
    local description="$2"
    
    print_header "Committing Changes"
    
    # Check if there are changes to commit
    if git diff --staged --quiet; then
        print_warning "No staged changes to commit"
        return 0
    fi
    
    # Show what will be committed
    print_info "Files to be committed:"
    git diff --staged --name-only | sed 's/^/  /'
    echo
    
    # Create commit message
    local full_message="$message"
    if [ -n "$description" ]; then
        full_message="$message

$description"
    fi
    
    if [ "$DRY_RUN" = true ]; then
        print_info "DRY RUN - Would commit with message:"
        echo "$full_message"
        return 0
    fi
    
    # Commit changes
    print_step "Committing changes..."
    if git commit -m "$full_message"; then
        print_success "Changes committed successfully"
    else
        print_error "Failed to commit changes"
        return 1
    fi
}

# Function to push changes
push_changes() {
    if [ "$PUSH_CHANGES" = false ]; then
        print_info "Skipping push (--no-push specified)"
        return 0
    fi
    
    print_header "Pushing Changes"
    
    local current_branch=$(git branch --show-current)
    local remote="origin"
    
    # Check if remote exists
    if ! git remote get-url "$remote" > /dev/null 2>&1; then
        print_warning "No remote '$remote' configured"
        return 0
    fi
    
    if [ "$DRY_RUN" = true ]; then
        print_info "DRY RUN - Would push to $remote/$current_branch"
        return 0
    fi
    
    print_step "Pushing to $remote/$current_branch..."
    if git push "$remote" "$current_branch"; then
        print_success "Changes pushed successfully"
    else
        print_error "Failed to push changes"
        return 1
    fi
}

# Function for interactive mode
interactive_mode() {
    print_header "Interactive Git Commit"
    
    # Show current status
    get_repo_status
    echo
    
    # Review changes
    if ! git diff --quiet || ! git diff --staged --quiet; then
        read -p "Review changes? (y/N): " review
        if [[ "$review" =~ ^[Yy]$ ]]; then
            review_changes
            echo
        fi
    fi
    
    # Select files
    if ! git diff --quiet; then
        read -p "Select specific files to stage? (y/N): " select
        if [[ "$select" =~ ^[Yy]$ ]]; then
            select_files
            echo
        else
            git add .
            print_success "All changes staged"
        fi
    fi
    
    # Generate or input commit message
    local auto_message=$(generate_commit_message)
    echo "Suggested commit message: $auto_message"
    read -p "Use suggested message? (Y/n): " use_suggested
    
    if [[ "$use_suggested" =~ ^[Nn]$ ]]; then
        read -p "Enter commit message: " COMMIT_MESSAGE
        read -p "Enter commit description (optional): " COMMIT_DESCRIPTION
    else
        COMMIT_MESSAGE="$auto_message"
        read -p "Enter additional description (optional): " COMMIT_DESCRIPTION
    fi
    
    # Confirm commit
    echo
    print_info "Commit message: $COMMIT_MESSAGE"
    if [ -n "$COMMIT_DESCRIPTION" ]; then
        print_info "Description: $COMMIT_DESCRIPTION"
    fi
    echo
    
    read -p "Proceed with commit? (Y/n): " proceed
    if [[ "$proceed" =~ ^[Nn]$ ]]; then
        print_info "Commit cancelled"
        return 0
    fi
}

# Function to parse command line arguments
parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_usage
                exit 0
                ;;
            -i|--interactive)
                INTERACTIVE_MODE=true
                shift
                ;;
            -r|--review)
                REVIEW_MODE=true
                shift
                ;;
            -s|--status)
                STATUS_MODE=true
                shift
                ;;
            --select-files)
                SELECT_FILES=true
                shift
                ;;
            --no-push)
                PUSH_CHANGES=false
                shift
                ;;
            --dry-run)
                DRY_RUN=true
                shift
                ;;
            -v|--verbose)
                set -x
                shift
                ;;
            -*)
                print_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
            *)
                if [ -z "$COMMIT_MESSAGE" ]; then
                    COMMIT_MESSAGE="$1"
                elif [ -z "$COMMIT_DESCRIPTION" ]; then
                    COMMIT_DESCRIPTION="$1"
                else
                    print_error "Too many arguments"
                    show_usage
                    exit 1
                fi
                shift
                ;;
        esac
    done
}

# Main function
main() {
    # Change to project root
    cd "$PROJECT_ROOT"
    
    # Check if we're in a git repository
    check_git_repo
    
    # Parse arguments
    parse_arguments "$@"
    
    # Handle special modes
    if [ "$STATUS_MODE" = true ]; then
        get_repo_status
        exit 0
    fi
    
    if [ "$REVIEW_MODE" = true ]; then
        review_changes
        exit 0
    fi
    
    if [ "$SELECT_FILES" = true ]; then
        select_files
        exit 0
    fi
    
    if [ "$INTERACTIVE_MODE" = true ]; then
        interactive_mode
    fi
    
    # Auto-generate commit message if not provided
    if [ -z "$COMMIT_MESSAGE" ]; then
        COMMIT_MESSAGE=$(generate_commit_message)
        print_info "Auto-generated commit message: $COMMIT_MESSAGE"
    fi
    
    # Stage all changes if not in interactive mode
    if [ "$INTERACTIVE_MODE" = false ]; then
        if ! git diff --quiet; then
            git add .
            print_success "All changes staged"
        fi
    fi
    
    # Commit changes
    commit_changes "$COMMIT_MESSAGE" "$COMMIT_DESCRIPTION"
    
    # Push changes
    push_changes
    
    print_success "Git operations completed successfully!"
}

# Run main function with all arguments
main "$@"
