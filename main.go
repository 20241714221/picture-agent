package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	database "picture_agent/databases"
	mcppkg "picture_agent/mcp"
	"picture_agent/tools"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

// projectRoot 获取项目根目录
func projectRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}

func getLlmTools(
	client *openai.Client,
	db *sql.DB,
) map[string]*tools.ToolDef {
	llmTools := make(map[string]*tools.ToolDef)
	foodTool := tools.NewFoodRecognitionTool(client)
	timeTool := tools.NewCurrentTimeTool()
	dailyRecordTool := tools.NewDailyRecordTool(db)
	nutritionQueryTool := tools.NewNutritionQueryTool(db)
	llmTools[foodTool.Name] = foodTool
	llmTools[timeTool.Name] = timeTool
	llmTools[dailyRecordTool.Name] = dailyRecordTool
	llmTools[nutritionQueryTool.Name] = nutritionQueryTool
	return llmTools
}

func main() {
	ctx := context.Background()
	root := projectRoot()

	// 普通文本模型
	client := openai.NewClient(
		option.WithAPIKey("sk-635d0cdd88c148bfa7fa8fc38edb4804"),
		option.WithBaseURL("https://api.deepseek.com"),
		option.WithRequestTimeout(30*time.Second),
	)

	// 视觉模型
	visionClient := openai.NewClient(
		option.WithAPIKey("sk-ws-H.PPMXHDE.IaZy.MEUCIHTnukXoGhK-9vVX8a5yrh6vHU7cBMI9P1Lw9SiJIUDkAiEAh3KEyjX4YXIphWI7-mPZt4Wodw8H47ZNrQvYhRbksAE"),
		option.WithBaseURL("https://dashscope.aliyuncs.com/compatible-mode/v1"),
		option.WithRequestTimeout(100*time.Second),
	)

	// 数据库
	db, err := database.NewMySQLDB()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// 注册原有 Go 工具
	llmTools := getLlmTools(&visionClient, db)

	// 创建 Agent
	agent := NewAgent(ctx, &client, llmTools)

	// 加载 MCP 配置并启动 MCP Server
	mcpConfigPath := filepath.Join(root, "config", "mcp.json")
	mcpCfg, err := mcppkg.LoadConfig(mcpConfigPath)
	if err != nil {
		fmt.Printf("警告: 加载 MCP 配置失败: %v\n", err)
		fmt.Println("将以无 MCP 模式继续运行...")
	} else {
		manager := mcppkg.NewManager(root)

		if err := manager.StartServers(ctx, mcpCfg); err != nil {
			fmt.Printf("警告: 启动 MCP Server 失败: %v\n", err)
			fmt.Println("将以无 MCP 模式继续运行...")
		} else {
			// 程序退出时关闭 MCP 连接
			defer manager.CloseAll()

			// 发现并注册 MCP 工具
			if err := agent.RegisterMCPTools(ctx, manager); err != nil {
				fmt.Printf("警告: 注册 MCP 工具失败: %v\n", err)
			} else {
				fmt.Println("MCP 工具注册成功")
			}
		}
	}

	// 启动 Bubble Tea
	p := tea.NewProgram(NewTUI(agent))
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
