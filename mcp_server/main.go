package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// projectRoot 获取项目根目录。
// 优先使用当前工作目录（go run 模式下工作目录为项目根目录），
// 如果当前目录下没有 web/index.html，则尝试可执行文件所在目录的上一级。
func projectRoot() string {
	// 1. 尝试当前工作目录
	wd, err := os.Getwd()
	if err == nil {
		if _, err := os.Stat(filepath.Join(wd, "web", "index.html")); err == nil {
			return wd
		}
	}

	// 2. 尝试可执行文件目录的上一级
	exe, err := os.Executable()
	if err == nil {
		candidate := filepath.Dir(filepath.Dir(exe))
		if _, err := os.Stat(filepath.Join(candidate, "web", "index.html")); err == nil {
			return candidate
		}
	}

	// 3. 回退到当前工作目录
	return wd
}

func main() {
	root := projectRoot()
	reportsDir := filepath.Join(root, "reports")

	// 确保 reports 目录存在
	_ = os.MkdirAll(reportsDir, 0755)

	impl := &sdkmcp.Implementation{
		Name:    "nutrition-report-server",
		Version: "1.0.0",
	}

	server := sdkmcp.NewServer(impl, nil)

	// 注册 generate_nutrition_report 工具
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        "generate_nutrition_report",
		Description: "根据营养数据生成独立的 HTML 营养分析报告。需要传入 date（日期）、records（食物明细数组）、summary（营养汇总）、weeklyCalories（近七天热量）、good（营养亮点）、warning（需要改善）、advice（饮食建议）。生成后返回报告文件路径。",
	}, func(ctx context.Context, req *sdkmcp.CallToolRequest, input ReportInput) (*sdkmcp.CallToolResult, any, error) {
		result, err := GenerateReport(input, root, reportsDir)
		if err != nil {
			return &sdkmcp.CallToolResult{
				Content: []sdkmcp.Content{
					&sdkmcp.TextContent{Text: fmt.Sprintf("报告生成失败: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}

		resultJSON, _ := json.Marshal(result)
		return &sdkmcp.CallToolResult{
			Content: []sdkmcp.Content{
				&sdkmcp.TextContent{Text: string(resultJSON)},
			},
		}, result, nil
	})

	// 使用 stdio transport 运行 Server
	if err := server.Run(context.Background(), &sdkmcp.StdioTransport{}); err != nil {
		fmt.Fprintf(os.Stderr, "MCP Server 运行失败: %v\n", err)
		os.Exit(1)
	}
}
