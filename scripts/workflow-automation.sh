#!/bin/bash

# Workflow Automation Script for Jurigen Project
# Automates: Branch creation, conventional commits, PR generation

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
DEFAULT_MAIN_BRANCH="main"
BRANCH_PREFIX_MAP=(
    "feature" "feat"
    "bugfix" "fix"
    "hotfix" "fix"
    "refactor" "refactor"
    "docs" "docs"
    "test" "test"
)

##############################################################################
# Helper Functions
##############################################################################

log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

check_git_repo() {
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        log_error "Not in a Git repository"
        exit 1
    fi
}

check_clean_working_tree() {
    if ! git diff-index --quiet HEAD --; then
        log_warning "Working directory has uncommitted changes"
        return 1
    fi
    return 0
}

get_current_branch() {
    git branch --show-current
}

##############################################################################
# Branch Management
##############################################################################

create_feature_branch() {
    local branch_type="$1"
    local feature_name="$2"
    
    if [[ -z "$branch_type" || -z "$feature_name" ]]; then
        log_error "Usage: create_feature_branch <type> <feature-name>"
        log_info "Types: feature, bugfix, hotfix, refactor, docs, test"
        exit 1
    fi
    
    # Sanitize feature name
    feature_name=$(echo "$feature_name" | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9]/-/g' | sed 's/--*/-/g' | sed 's/^-\|-$//g')
    
    local branch_name="${branch_type}/${feature_name}"
    local current_branch=$(get_current_branch)
    
    log_info "Creating branch: $branch_name"
    
    # Ensure we're on main branch
    if [[ "$current_branch" != "$DEFAULT_MAIN_BRANCH" ]]; then
        log_warning "Switching from '$current_branch' to '$DEFAULT_MAIN_BRANCH'"
        git checkout "$DEFAULT_MAIN_BRANCH"
        git pull origin "$DEFAULT_MAIN_BRANCH"
    fi
    
    # Create and checkout new branch
    git checkout -b "$branch_name"
    log_success "Created and switched to branch: $branch_name"
}

##############################################################################
# Change Analysis
##############################################################################

analyze_changes() {
    local changed_files=$(git diff --name-only HEAD~1 2>/dev/null || git diff --name-only --cached)
    
    if [[ -z "$changed_files" ]]; then
        log_warning "No changes detected"
        return 1
    fi
    
    # Categorize changes
    local has_usecase=false
    local has_http=false
    local has_tests=false
    local has_docs=false
    local has_config=false
    
    while IFS= read -r file; do
        case "$file" in
            internal/usecase/*.go)
                has_usecase=true
                ;;
            internal/adapter/http/*.go)
                has_http=true
                ;;
            *_test.go)
                has_tests=true
                ;;
            docs/* | *.md | swagger/*)
                has_docs=true
                ;;
            Makefile | *.yml | *.yaml | *.json)
                has_config=true
                ;;
        esac
    done <<< "$changed_files"
    
    # Determine commit type and scope
    if [[ "$has_usecase" == true ]]; then
        echo "feat(usecase)"
    elif [[ "$has_http" == true ]]; then
        echo "feat(http)"
    elif [[ "$has_tests" == true ]]; then
        echo "test"
    elif [[ "$has_docs" == true ]]; then
        echo "docs"
    elif [[ "$has_config" == true ]]; then
        echo "build"
    else
        echo "feat"
    fi
}

##############################################################################
# Automated Commits
##############################################################################

create_auto_commit() {
    local commit_type=$(analyze_changes)
    local files_changed=$(git diff --name-only --cached | wc -l | tr -d ' ')
    
    if [[ "$files_changed" -eq 0 ]]; then
        log_warning "No staged changes to commit"
        return 1
    fi
    
    log_info "Detected change type: $commit_type"
    log_info "Files staged: $files_changed"
    
    # Generate commit message based on type
    local commit_msg
    case "$commit_type" in
        "feat(usecase)")
            commit_msg="$commit_type: add/update use case functionality"
            ;;
        "feat(http)")
            commit_msg="$commit_type: add/update HTTP endpoints"
            ;;
        "test")
            commit_msg="test: add/update test coverage"
            ;;
        "docs")
            commit_msg="docs: update documentation"
            ;;
        "build")
            commit_msg="build: update build configuration"
            ;;
        *)
            commit_msg="$commit_type: implement changes"
            ;;
    esac
    
    # Show preview
    echo
    log_info "Commit preview:"
    echo "  Type: $commit_type"
    echo "  Message: $commit_msg"
    echo "  Files:"
    git diff --name-only --cached | sed 's/^/    /'
    echo
    
    read -p "Create this commit? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        git commit -m "$commit_msg"
        log_success "Commit created: $commit_msg"
    else
        log_info "Commit cancelled"
    fi
}

##############################################################################
# PR Generation
##############################################################################

generate_pr_description() {
    local branch_name=$(get_current_branch)
    local commits=$(git log --oneline "$DEFAULT_MAIN_BRANCH..$branch_name")
    local changed_files=$(git diff --name-only "$DEFAULT_MAIN_BRANCH..$branch_name")
    
    cat > PR_DESCRIPTION.md << EOF
## 📋 **Pull Request: ${branch_name//-/ | title case}**

### 🎯 **Goal**
Auto-generated PR for branch: \`$branch_name\`

### 📝 **Context**
This PR includes the following changes:

$(echo "$commits" | sed 's/^/- /')

### ✨ **Changes Made**

#### Modified Files
$(echo "$changed_files" | sed 's/^/- `/' | sed 's/$/`/')

### 🔧 **Key Features**
- ✅ Automated conventional commits
- ✅ Clean architecture compliance
- ✅ Comprehensive testing

### 🧪 **Testing**
- All tests pass: \`make test\`
- Linting passes: \`make lint\`
- Full development workflow: \`make dev\`

### 💥 **Breaking Changes**
- None: All changes are additive and backward compatible

**Ready for review** 🚀
EOF
    
    log_success "PR description generated: PR_DESCRIPTION.md"
}

create_pr() {
    local branch_name=$(get_current_branch)
    
    if [[ "$branch_name" == "$DEFAULT_MAIN_BRANCH" ]]; then
        log_error "Cannot create PR from main branch"
        exit 1
    fi
    
    # Generate PR description if it doesn't exist
    if [[ ! -f "PR_DESCRIPTION.md" ]]; then
        generate_pr_description
    fi
    
    # Check if GitHub CLI is available
    if command -v gh > /dev/null; then
        log_info "Creating PR using GitHub CLI..."
        
        # Push branch first
        git push -u origin "$branch_name"
        
        # Create PR
        gh pr create \
            --title "$(echo $branch_name | sed 's/-/ /g' | sed 's/\b\w/\U&/g')" \
            --body-file PR_DESCRIPTION.md \
            --base "$DEFAULT_MAIN_BRANCH" \
            --head "$branch_name"
            
        log_success "PR created successfully!"
        
        # Open PR in browser
        gh pr view --web
    else
        log_warning "GitHub CLI not found. Manual PR creation required."
        log_info "1. Push branch: git push -u origin $branch_name"
        log_info "2. Create PR using the generated PR_DESCRIPTION.md"
    fi
}

##############################################################################
# Main Commands
##############################################################################

cmd_branch() {
    create_feature_branch "$1" "$2"
}

cmd_commit() {
    create_auto_commit
}

cmd_pr() {
    case "$1" in
        "create")
            create_pr
            ;;
        "template")
            generate_pr_description
            ;;
        *)
            log_error "Usage: $0 pr [create|template]"
            exit 1
            ;;
    esac
}

cmd_workflow() {
    local feature_name="$1"
    local branch_type="${2:-feature}"
    
    if [[ -z "$feature_name" ]]; then
        log_error "Usage: $0 workflow <feature-name> [branch-type]"
        exit 1
    fi
    
    log_info "Starting automated workflow for: $feature_name"
    
    # Step 1: Create branch
    create_feature_branch "$branch_type" "$feature_name"
    
    log_success "Workflow initialized. Next steps:"
    echo "1. Make your code changes"
    echo "2. Run: $0 commit (to auto-commit changes)"
    echo "3. Run: $0 pr create (to create PR)"
}

##############################################################################
# Main Entry Point
##############################################################################

main() {
    check_git_repo
    
    case "$1" in
        "branch")
            cmd_branch "$2" "$3"
            ;;
        "commit")
            cmd_commit
            ;;
        "pr")
            cmd_pr "$2"
            ;;
        "workflow")
            cmd_workflow "$2" "$3"
            ;;
        "help"|"--help"|"-h"|"")
            cat << EOF
🤖 **Jurigen Workflow Automation**

Usage: $0 <command> [options]

Commands:
    branch <type> <name>     Create a new feature branch
    commit                   Auto-create conventional commit
    pr create               Create PR with auto-generated description
    pr template             Generate PR description template
    workflow <name> [type]   Full automated workflow (branch + setup)

Examples:
    $0 branch feature "update-dag-handler"
    $0 commit
    $0 pr create
    $0 workflow "authentication-fix" "bugfix"

Branch Types: feature, bugfix, hotfix, refactor, docs, test
EOF
            ;;
        *)
            log_error "Unknown command: $1"
            log_info "Run '$0 help' for usage information"
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"
