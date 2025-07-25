#!/bin/bash

################################################################################
# Branch Helper Script
# 帮助开发者遵循GitFlow分支管理策略的自动化脚本
################################################################################

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
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

# 检查Git仓库状态
check_git_status() {
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        print_error "当前目录不是Git仓库"
        exit 1
    fi
    
    if [[ -n $(git status --porcelain) ]]; then
        print_warning "工作目录有未提交的更改"
        git status --short
        read -p "是否继续？(y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
}

# 更新分支
update_branch() {
    local branch=$1
    print_info "更新分支 $branch"
    git checkout $branch
    git pull origin $branch
}

# 创建Feature分支
create_feature_branch() {
    local feature_name=$1
    
    if [[ -z "$feature_name" ]]; then
        read -p "请输入功能名称: " feature_name
    fi
    
    if [[ -z "$feature_name" ]]; then
        print_error "功能名称不能为空"
        exit 1
    fi
    
    # 验证分支名称格式
    if [[ ! "$feature_name" =~ ^[a-z0-9-]+$ ]]; then
        print_error "功能名称只能包含小写字母、数字和连字符"
        exit 1
    fi
    
    local branch_name="feature/$feature_name"
    
    # 检查分支是否已存在
    if git show-ref --verify --quiet refs/heads/$branch_name; then
        print_error "分支 $branch_name 已存在"
        exit 1
    fi
    
    print_info "从develop分支创建功能分支: $branch_name"
    update_branch "develop"
    git checkout -b $branch_name
    git push -u origin $branch_name
    
    print_success "功能分支 $branch_name 创建成功"
    print_info "现在可以开始开发功能了！"
}

# 创建Bugfix分支
create_bugfix_branch() {
    local bug_name=$1
    
    if [[ -z "$bug_name" ]]; then
        read -p "请输入bug描述: " bug_name
    fi
    
    if [[ -z "$bug_name" ]]; then
        print_error "bug描述不能为空"
        exit 1
    fi
    
    # 验证分支名称格式
    if [[ ! "$bug_name" =~ ^[a-z0-9-]+$ ]]; then
        print_error "bug描述只能包含小写字母、数字和连字符"
        exit 1
    fi
    
    local branch_name="bugfix/$bug_name"
    
    # 检查分支是否已存在
    if git show-ref --verify --quiet refs/heads/$branch_name; then
        print_error "分支 $branch_name 已存在"
        exit 1
    fi
    
    print_info "从develop分支创建bug修复分支: $branch_name"
    update_branch "develop"
    git checkout -b $branch_name
    git push -u origin $branch_name
    
    print_success "Bug修复分支 $branch_name 创建成功"
}

# 创建Release分支
create_release_branch() {
    local version=$1
    
    if [[ -z "$version" ]]; then
        read -p "请输入版本号 (例如: v1.1.0): " version
    fi
    
    if [[ -z "$version" ]]; then
        print_error "版本号不能为空"
        exit 1
    fi
    
    # 验证版本号格式
    if [[ ! "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[a-z0-9]+(\.[0-9]+)?)?$ ]]; then
        print_error "版本号格式不正确，应该类似: v1.1.0 或 v1.1.0-beta.1"
        exit 1
    fi
    
    local branch_name="release/$version"
    
    # 检查分支是否已存在
    if git show-ref --verify --quiet refs/heads/$branch_name; then
        print_error "分支 $branch_name 已存在"
        exit 1
    fi
    
    # 检查标签是否已存在
    if git show-ref --verify --quiet refs/tags/$version; then
        print_error "标签 $version 已存在"
        exit 1
    fi
    
    print_info "从develop分支创建发布分支: $branch_name"
    update_branch "develop"
    git checkout -b $branch_name
    
    # 更新版本信息
    print_info "更新版本信息..."
    if [[ -f "internal/version/version.go" ]]; then
        sed -i.bak "s/Version.*=.*/Version = \"$version\"/" internal/version/version.go
        rm internal/version/version.go.bak
        git add internal/version/version.go
    fi
    
    # 更新README中的版本徽章
    if [[ -f "README.md" ]]; then
        sed -i.bak "s/version-v[0-9]\+\.[0-9]\+\.[0-9]\+/version-$version/" README.md
        rm README.md.bak
        git add README.md
    fi
    
    git commit -m "chore: prepare release $version"
    git push -u origin $branch_name
    
    print_success "发布分支 $branch_name 创建成功"
    print_info "请完成以下步骤："
    print_info "1. 更新 docs/changelog.md"
    print_info "2. 更新 docs/release-notes.md"
    print_info "3. 运行发布前测试"
    print_info "4. 创建PR到main分支"
}

# 创建Hotfix分支
create_hotfix_branch() {
    local hotfix_name=$1
    
    if [[ -z "$hotfix_name" ]]; then
        read -p "请输入hotfix描述: " hotfix_name
    fi
    
    if [[ -z "$hotfix_name" ]]; then
        print_error "hotfix描述不能为空"
        exit 1
    fi
    
    # 验证分支名称格式
    if [[ ! "$hotfix_name" =~ ^[a-z0-9-]+$ ]]; then
        print_error "hotfix描述只能包含小写字母、数字和连字符"
        exit 1
    fi
    
    local branch_name="hotfix/$hotfix_name"
    
    # 检查分支是否已存在
    if git show-ref --verify --quiet refs/heads/$branch_name; then
        print_error "分支 $branch_name 已存在"
        exit 1
    fi
    
    print_info "从main分支创建hotfix分支: $branch_name"
    update_branch "main"
    git checkout -b $branch_name
    git push -u origin $branch_name
    
    print_success "Hotfix分支 $branch_name 创建成功"
    print_warning "这是紧急修复分支，请尽快完成修复并合并"
}

# 完成Feature分支
finish_feature_branch() {
    local current_branch=$(git branch --show-current)
    
    if [[ ! "$current_branch" =~ ^feature/ ]]; then
        print_error "当前不在feature分支上"
        exit 1
    fi
    
    print_info "准备完成功能分支: $current_branch"
    
    # 检查是否有未提交的更改
    if [[ -n $(git status --porcelain) ]]; then
        print_error "有未提交的更改，请先提交"
        exit 1
    fi
    
    # 推送最新更改
    git push origin $current_branch
    
    print_info "请在GitHub上创建Pull Request:"
    print_info "从: $current_branch"
    print_info "到: develop"
    
    # 生成PR URL
    local repo_url=$(git config --get remote.origin.url | sed 's/\.git$//')
    if [[ "$repo_url" =~ github\.com ]]; then
        local pr_url="${repo_url}/compare/develop...$current_branch?expand=1"
        print_info "PR链接: $pr_url"
    fi
}

# 清理已合并的分支
cleanup_merged_branches() {
    print_info "清理已合并的分支..."
    
    # 更新远程分支信息
    git fetch --prune
    
    # 获取已合并到develop的分支
    local merged_branches=$(git branch --merged develop | grep -E "^\s*(feature|bugfix)/" | sed 's/^\s*//')
    
    if [[ -z "$merged_branches" ]]; then
        print_info "没有需要清理的分支"
        return
    fi
    
    print_info "发现以下已合并的分支:"
    echo "$merged_branches"
    
    read -p "是否删除这些分支？(y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        for branch in $merged_branches; do
            print_info "删除分支: $branch"
            git branch -d $branch
            
            # 尝试删除远程分支
            if git show-ref --verify --quiet refs/remotes/origin/$branch; then
                git push origin --delete $branch 2>/dev/null || print_warning "无法删除远程分支 $branch"
            fi
        done
        print_success "分支清理完成"
    fi
}

# 显示分支状态
show_branch_status() {
    print_info "分支状态概览"
    echo
    
    print_info "当前分支: $(git branch --show-current)"
    echo
    
    print_info "本地分支:"
    git branch -v
    echo
    
    print_info "远程分支:"
    git branch -r
    echo
    
    print_info "最近的标签:"
    git tag --sort=-version:refname | head -5
    echo
    
    print_info "未推送的提交:"
    local current_branch=$(git branch --show-current)
    local unpushed=$(git log origin/$current_branch..$current_branch --oneline 2>/dev/null || echo "无法检查（分支可能不存在于远程）")
    if [[ -n "$unpushed" ]]; then
        echo "$unpushed"
    else
        echo "无"
    fi
}

# 显示帮助信息
show_help() {
    echo "Branch Helper Script - GitFlow分支管理助手"
    echo
    echo "用法: $0 <command> [options]"
    echo
    echo "命令:"
    echo "  feature <name>     创建功能分支"
    echo "  bugfix <name>      创建bug修复分支"
    echo "  release <version>  创建发布分支"
    echo "  hotfix <name>      创建hotfix分支"
    echo "  finish             完成当前功能分支"
    echo "  cleanup            清理已合并的分支"
    echo "  status             显示分支状态"
    echo "  help               显示此帮助信息"
    echo
    echo "示例:"
    echo "  $0 feature add-new-runtime"
    echo "  $0 bugfix fix-config-parsing"
    echo "  $0 release v1.1.0"
    echo "  $0 hotfix critical-security-fix"
    echo
}

# 主函数
main() {
    case "${1:-}" in
        "feature")
            check_git_status
            create_feature_branch "$2"
            ;;
        "bugfix")
            check_git_status
            create_bugfix_branch "$2"
            ;;
        "release")
            check_git_status
            create_release_branch "$2"
            ;;
        "hotfix")
            check_git_status
            create_hotfix_branch "$2"
            ;;
        "finish")
            check_git_status
            finish_feature_branch
            ;;
        "cleanup")
            check_git_status
            cleanup_merged_branches
            ;;
        "status")
            show_branch_status
            ;;
        "help"|"--help"|"-h")
            show_help
            ;;
        "")
            print_error "请指定命令"
            show_help
            exit 1
            ;;
        *)
            print_error "未知命令: $1"
            show_help
            exit 1
            ;;
    esac
}

# 运行主函数
main "$@"