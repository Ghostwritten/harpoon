#!/bin/bash

# 代码审查自动化脚本
# 用于执行全面的代码质量检查和分析

set -e

echo "🔍 开始代码审查..."

# 创建报告目录
REPORT_DIR=".kiro/specs/harpoon-code-review/review-results/reports"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
REPORT_FILE="$REPORT_DIR/code_review_$TIMESTAMP.md"

mkdir -p "$REPORT_DIR"

echo "# 代码审查报告 - $TIMESTAMP" > "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

# 1. Go 代码格式检查
echo "📝 检查代码格式..."
echo "## 代码格式检查" >> "$REPORT_FILE"
if ! gofmt -l . | grep -q .; then
    echo "✅ 代码格式符合标准" >> "$REPORT_FILE"
else
    echo "❌ 发现格式问题:" >> "$REPORT_FILE"
    echo '```' >> "$REPORT_FILE"
    gofmt -l . >> "$REPORT_FILE"
    echo '```' >> "$REPORT_FILE"
fi
echo "" >> "$REPORT_FILE"

# 2. Go vet 检查
echo "🔧 运行 go vet..."
echo "## Go Vet 检查" >> "$REPORT_FILE"
if go vet ./...; then
    echo "✅ Go vet 检查通过" >> "$REPORT_FILE"
else
    echo "❌ Go vet 发现问题" >> "$REPORT_FILE"
fi
echo "" >> "$REPORT_FILE"

# 3. 运行 golangci-lint (如果安装了)
echo "🧹 运行 golangci-lint..."
echo "## GolangCI-Lint 检查" >> "$REPORT_FILE"
if command -v golangci-lint &> /dev/null; then
    if golangci-lint run --out-format=github-actions; then
        echo "✅ GolangCI-Lint 检查通过" >> "$REPORT_FILE"
    else
        echo "❌ GolangCI-Lint 发现问题" >> "$REPORT_FILE"
    fi
else
    echo "⚠️ GolangCI-Lint 未安装，跳过检查" >> "$REPORT_FILE"
    echo "安装命令: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$(go env GOPATH)/bin v1.54.2"
fi
echo "" >> "$REPORT_FILE"

# 4. 测试覆盖率
echo "🧪 检查测试覆盖率..."
echo "## 测试覆盖率" >> "$REPORT_FILE"
if go test -coverprofile=coverage.out ./... 2>/dev/null; then
    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
    echo "📊 总体测试覆盖率: $COVERAGE" >> "$REPORT_FILE"
    
    # 生成 HTML 覆盖率报告
    go tool cover -html=coverage.out -o coverage.html
    echo "📄 详细覆盖率报告已生成: coverage.html" >> "$REPORT_FILE"
else
    echo "❌ 测试执行失败或无测试文件" >> "$REPORT_FILE"
fi
echo "" >> "$REPORT_FILE"

# 5. 依赖检查
echo "📦 检查依赖..."
echo "## 依赖分析" >> "$REPORT_FILE"
echo "### 直接依赖:" >> "$REPORT_FILE"
echo '```' >> "$REPORT_FILE"
go list -m -f '{{.Path}} {{.Version}}' all | grep -v "github.com/harpoon/hpn" | head -20 >> "$REPORT_FILE"
echo '```' >> "$REPORT_FILE"

# 检查过时的依赖
echo "### 依赖更新检查:" >> "$REPORT_FILE"
if command -v go-mod-outdated &> /dev/null; then
    go list -u -m -json all | go-mod-outdated -update -direct >> "$REPORT_FILE" 2>/dev/null || echo "所有依赖都是最新的" >> "$REPORT_FILE"
else
    echo "⚠️ go-mod-outdated 未安装，无法检查过时依赖" >> "$REPORT_FILE"
fi
echo "" >> "$REPORT_FILE"

# 6. 安全检查
echo "🔒 安全检查..."
echo "## 安全分析" >> "$REPORT_FILE"
if command -v gosec &> /dev/null; then
    if gosec -fmt json -out gosec-report.json ./... 2>/dev/null; then
        ISSUES=$(jq '.Issues | length' gosec-report.json 2>/dev/null || echo "0")
        if [ "$ISSUES" -eq 0 ]; then
            echo "✅ 未发现安全问题" >> "$REPORT_FILE"
        else
            echo "❌ 发现 $ISSUES 个潜在安全问题" >> "$REPORT_FILE"
            echo "详细报告: gosec-report.json" >> "$REPORT_FILE"
        fi
    else
        echo "❌ 安全检查执行失败" >> "$REPORT_FILE"
    fi
else
    echo "⚠️ gosec 未安装，跳过安全检查" >> "$REPORT_FILE"
    echo "安装命令: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"
fi
echo "" >> "$REPORT_FILE"

# 7. 代码复杂度分析
echo "📈 代码复杂度分析..."
echo "## 代码复杂度" >> "$REPORT_FILE"
if command -v gocyclo &> /dev/null; then
    echo "### 高复杂度函数 (>10):" >> "$REPORT_FILE"
    echo '```' >> "$REPORT_FILE"
    gocyclo -over 10 . >> "$REPORT_FILE" 2>/dev/null || echo "未发现高复杂度函数" >> "$REPORT_FILE"
    echo '```' >> "$REPORT_FILE"
else
    echo "⚠️ gocyclo 未安装，跳过复杂度分析" >> "$REPORT_FILE"
    echo "安装命令: go install github.com/fzipp/gocyclo/cmd/gocyclo@latest"
fi
echo "" >> "$REPORT_FILE"

# 8. 项目统计
echo "📊 项目统计..."
echo "## 项目统计" >> "$REPORT_FILE"
echo "### 代码行数:" >> "$REPORT_FILE"
echo '```' >> "$REPORT_FILE"
find . -name "*.go" -not -path "./vendor/*" | xargs wc -l | tail -1 >> "$REPORT_FILE"
echo '```' >> "$REPORT_FILE"

echo "### 文件统计:" >> "$REPORT_FILE"
echo '```' >> "$REPORT_FILE"
echo "Go 文件数量: $(find . -name "*.go" -not -path "./vendor/*" | wc -l)" >> "$REPORT_FILE"
echo "测试文件数量: $(find . -name "*_test.go" -not -path "./vendor/*" | wc -l)" >> "$REPORT_FILE"
echo '```' >> "$REPORT_FILE"

echo "" >> "$REPORT_FILE"
echo "---" >> "$REPORT_FILE"
echo "报告生成时间: $(date)" >> "$REPORT_FILE"
echo "审查者: 自动化脚本" >> "$REPORT_FILE"

echo "✅ 代码审查完成！"
echo "📄 报告已保存到: $REPORT_FILE"

# 清理临时文件
rm -f coverage.out gosec-report.json