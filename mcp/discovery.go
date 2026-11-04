package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"picture_agent/tools"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/openai/openai-go/v3"
)

// DiscoverTools 调用 MCP Server 的 ListTools，发现其提供的所有工具
func DiscoverTools(ctx context.Context, session *sdkmcp.ClientSession) ([]*sdkmcp.Tool, error) {
	if session == nil {
		return nil, fmt.Errorf("MCP Session 不能为空")
	}

	result, err := session.ListTools(ctx, &sdkmcp.ListToolsParams{})
	if err != nil {
		return nil, fmt.Errorf("发现 MCP 工具失败: %w", err)
	}

	return result.Tools, nil
}

// ConvertToToolDefs 将 MCP 工具列表转换为项目自定义的 ToolDef 列表。
// 每个 ToolDef 的 Func 会调用对应 MCP Server 的 CallTool。
func ConvertToToolDefs(session *sdkmcp.ClientSession, mcpTools []*sdkmcp.Tool) ([]*tools.ToolDef, error) {
	if session == nil {
		return nil, fmt.Errorf("MCP Session 不能为空")
	}

	result := make([]*tools.ToolDef, 0, len(mcpTools))

	for _, mcpTool := range mcpTools {
		// 1. 将 MCP 的 InputSchema 转成 OpenAI FunctionParameters
		schemaJSON, err := json.Marshal(mcpTool.InputSchema)
		if err != nil {
			return nil, fmt.Errorf("转换工具 %s 的参数结构失败: %w", mcpTool.Name, err)
		}

		var params openai.FunctionParameters
		if err := json.Unmarshal(schemaJSON, &params); err != nil {
			return nil, fmt.Errorf("解析工具 %s 的参数结构失败: %w", mcpTool.Name, err)
		}

		// 2. 保存工具名称，避免闭包变量捕获问题
		toolName := mcpTool.Name

		// 3. 创建项目自己的 ToolDef
		tool := &tools.ToolDef{
			Name:        toolName,
			Description: mcpTool.Description,
			Params:      params,

			Func: func(ctx context.Context, args map[string]any) (string, error) {
				callResult, err := session.CallTool(
					ctx,
					&sdkmcp.CallToolParams{
						Name:      toolName,
						Arguments: args,
					},
				)
				if err != nil {
					return "", fmt.Errorf("调用 MCP 工具 %s 失败: %w", toolName, err)
				}

				// 将 MCP 返回结果序列化为 JSON 字符串，交给 Agent
				resultJSON, err := json.Marshal(callResult)
				if err != nil {
					return "", fmt.Errorf("序列化 MCP 工具 %s 的结果失败: %w", toolName, err)
				}

				return string(resultJSON), nil
			},
		}

		result = append(result, tool)
	}

	return result, nil
}
