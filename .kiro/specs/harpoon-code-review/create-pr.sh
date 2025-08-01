#!/bin/bash

# 创建Pull Request脚本
# 用于从develop分支向main分支创建代码审查PR

set -e

echo "🚀 创建Harpoon代码审查Pull Request"
echo "=================================="

# 检查是否安装了GitHub CLI
if command -v gh &> /dev/null; then
    echo "✅ 检测到GitHub CLI，使用gh命令创建PR"
    
    # 使用GitHub CLI创建PR
    gh pr create \
        --base main \
        --head develop \
        --title "🔍 Harpoon项目全面代码审查结果" \
        --body-file .kiro/specs/harpoon-code-review/pr-description.md \
        --assignee @me \
        --label "code-review,documentation,quality-improvement" \
        --draft
    
    echo "✅ Pull Request已创建为草稿状态"
    echo "📝 请访问GitHub查看并在准备好后发布PR"
    
else
    echo "⚠️  GitHub CLI未安装，请手动创建Pull Request"
    echo ""
    echo "📋 PR创建信息:"
    echo "  - 源分支: develop"
    echo "  - 目标分支: main"
    echo "  - 标题: 🔍 Harpoon项目全面代码审查结果"
    echo "  - 描述文件: .kiro/specs/harpoon-code-review/pr-description.md"
    echo "  - 标签: code-review, documentation, quality-improvement"
    echo ""
    echo "🔗 GitHub PR创建链接:"
    echo "   https://github.com/Ghostwritten/harpoon/compare/main...develop"
    echo ""
    echo "📄 PR描述内容已保存在:"
    echo "   .kiro/specs/harpoon-code-review/pr-description.md"
    echo ""
    echo "📁 审查报告文件位于:"
    echo "   .kiro/specs/harpoon-code-review/review-results/"
    echo ""
    echo "🛠️  安装GitHub CLI (可选):"
    echo "   brew install gh  # macOS"
    echo "   # 或访问: https://cli.github.com/"
fi

echo ""
echo "📋 PR检查清单:"
echo "  □ PR标题和描述已填写"
echo "  □ 审查报告文件已包含"
echo "  □ 分配了审查者"
echo "  □ 添加了相关标签"
echo "  □ 设置为草稿状态（如需要）"
echo ""
echo "🎯 下一步行动:"
echo "  1. 在GitHub上查看创建的PR"
echo "  2. 邀请团队成员进行代码审查"
echo "  3. 根据审查反馈进行调整"
echo "  4. 准备好后将PR标记为ready for review"
echo ""
echo "✨ 代码审查PR创建完成！"