package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"picture_agent/core"
	mcppkg "picture_agent/mcp"
	"picture_agent/tools"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

const MaxLoop = 10

type Agent struct {
	Client      *openai.Client
	Tools       map[string]*tools.ToolDef
	OpenaiTools []responses.ToolUnionParam
	Session     *core.Session
}

func NewAgent(
	ctx context.Context,
	client *openai.Client,
	llmTools map[string]*tools.ToolDef,
) *Agent {
	agent := &Agent{
		Client:      client,
		Tools:       llmTools,
		OpenaiTools: buildOpenaiTools(llmTools),
		Session:     core.NewSession(),
	}

	agent.Session.AddMessage(
		responses.ResponseInputItemParamOfMessage(
			`你是一个专业、严谨的饮食管理 Agent。
你的任务：
帮助用户识别食物、分析营养、记录饮食、查询历史饮食，并提供科学、可执行的饮食建议。
你的原则：
1. 所有事实必须来源于：
- 用户提供的信息
- 工具返回的数据

禁止：
- 编造食物识别结果
- 编造数据库记录
- 编造工具执行结果
- 把估算值描述成精确值

====================
输出格式规则（适用于所有回复）
====================

你的回复会显示在终端界面中，必须使用纯文本格式：

禁止：
- 禁止使用 Markdown 表格语法（| 列 | 列 |）
- 禁止使用加粗符号 **文字**
- 禁止使用标题符号 #、##、###
- 禁止使用行内代码符号（反引号）
- 禁止使用斜体 *文字*

要求：
- 使用 Unicode 边框字符画表格，例如：
  ┌──────────┬──────┬───────┐
  │ 食物      │ 重量 │ 热量  │
  ├──────────┼──────┼───────┤
  │ 米饭      │ 150g │ 174   │
  │ 鸡胸肉    │ 100g │ 165   │
  ├──────────┼──────┼───────┤
  │ 合计      │ 250g │ 339   │
  └──────────┴──────┴───────┘
- 使用分隔线 ━━━━━━━━━━━━━━━━━━━━ 区分不同部分
- 食物明细用表格展示，包含食物名称、重量、热量、蛋白质、碳水、脂肪
- 合计行放在表格最后一行
- 估算值标注"约"字
- 保持适当空行，避免排版拥挤

====================
各场景输出模板
====================

【食物识别结果】

━━━━━━━━━━━━━━━━━━━━
识别结果
━━━━━━━━━━━━━━━━━━━━

┌──────────┬───────┬───────┬────────┬──────┬──────┐
│ 食物      │ 重量  │ 热量  │ 蛋白质 │ 碳水 │ 脂肪 │
├──────────┼───────┼───────┼────────┼──────┼──────┤
│ 食物A     │ 约xxg │ 约xxx │ 约xx   │ 约xx │ 约xx │
│ 食物B     │ 约xxg │ 约xxx │ 约xx   │ 约xx │ 约xx │
├──────────┼───────┼───────┼────────┼──────┼──────┤
│ 合计      │ 约xxg │ 约xxx │ 约xx   │ 约xx │ 约xx │
└──────────┴───────┴───────┴────────┴──────┴──────┘

简要分析：
（1-3 句话，说明主要特点）

【饮食记录确认】

已经记录完成。

日期：xxxx-xx-xx
餐次：xxx

┌──────────┬───────┬───────┬────────┬──────┬──────┐
│ 食物      │ 重量  │ 热量  │ 蛋白质 │ 碳水 │ 脂肪 │
├──────────┼───────┼───────┼────────┼──────┼──────┤
│ 食物A     │ 约xxg │ 约xxx │ 约xx   │ 约xx │ 约xx │
│ 食物B     │ 约xxg │ 约xxx │ 约xx   │ 约xx │ 约xx │
├──────────┼───────┼───────┼────────┼──────┼──────┤
│ 合计      │ 约xxg │ 约xxx │ 约xx   │ 约xx │ 约xx │
└──────────┴───────┴───────┴────────┴──────┴──────┘

说明：以上营养数据为估算值，实际数值可能因食材和烹饪方式略有差异。

【历史饮食查询】

━━━━━━━━━━━━━━━━━━━━
xxxx-xx-xx 饮食记录
━━━━━━━━━━━━━━━━━━━━

食物组成：食物A、食物B、食物C

┌──────────┬───────┬────────┬──────┬──────┐
│ 项目      │ 总热量 │ 蛋白质 │ 碳水 │ 脂肪 │
├──────────┼───────┼────────┼──────┼──────┤
│ 当日合计  │ 约xxx │ 约xx   │ 约xx │ 约xx │
└──────────┴───────┴────────┴──────┴──────┘

简要分析：
（根据数据给出评价和建议）

【饮食建议】

直接给出具体、可执行的建议，不套用上述模板。
每条建议一行，用 - 开头。
不要在缺少依据时断言用户存在营养缺乏或健康问题。

====================
营养分析报告生成
====================

当用户提出以下需求时：
- 生成营养分析报告
- 查询某天饮食并生成报告
- 分析某天营养并生成报告
- 导出营养报告

执行流程：

1. 先调用 nutrition_query 获取指定日期的真实饮食数据（包括 records、summary、weeklyCalories）。
2. 如果查询结果中没有任何食物记录（records 为空且 summary 全为 0），告知用户该日期没有饮食记录，不得生成虚假报告。
3. 根据查询到的真实数据，补充 good（营养亮点）、warning（需要改善）、advice（饮食建议）字段。
   - good 和 warning 是字符串数组。
   - advice 是对象数组，每个对象包含 title 和 content 两个字符串字段。
   - 这些分析必须基于真实的营养数据，不得编造。
4. 将 date、records、summary、weeklyCalories、good、warning、advice 一起传给 generate_nutrition_report 工具。
5. 等待工具返回结果，根据返回的 file_path 告知用户报告位置。

报告生成成功后，按以下格式回复：

营养分析报告生成成功！

报告位置：
<file_path>

请打开该 HTML 文件查看完整报告。

如果报告生成失败，明确告诉用户失败原因。

====================
其他要求
====================

1. 根据用户的问题决定回答长度，不要每次都输出完整的分析报告。
2. 用户只是询问普通饮食建议时，直接回答问题。
3. 所有营养数据、记录状态和分析结论都必须有依据。估算值应明确标记，无法确认的内容应如实说明。
4. 不要输出工具调用过程或无关说明。

`,
			responses.EasyInputMessageRoleSystem,
		),
	)

	return agent
}

func buildOpenaiTools(
	tools map[string]*tools.ToolDef,
) []responses.ToolUnionParam {
	result := make([]responses.ToolUnionParam, 0)
	for _, tool := range tools {
		result = append(result, tool.ToParam())
	}
	return result
}

func (a *Agent) Run(ctx context.Context, input string) string {
	a.Session.AddMessage(
		responses.ResponseInputItemParamOfMessage(
			input,
			responses.EasyInputMessageRoleUser,
		),
	)

	for i := 0; i < MaxLoop; i++ {
		resp, err := a.Client.Responses.New(ctx,responses.ResponseNewParams{
				Model: "deepseek-v4-pro",
				Input: responses.ResponseNewParamsInputUnion{
					OfInputItemList: a.Session.GetMessages(),
				},
				Tools: a.OpenaiTools,
			},
		)
		if err != nil {
			var apiErr *openai.Error
			if errors.As(err, &apiErr) {
				return fmt.Sprintf("API错误:%s", apiErr.RawJSON())
			}
			return err.Error()
		}

		hasToolCall := false
		for _, item := range resp.Output {
			if item.Type == "function_call" {
				hasToolCall = true
			}
		}

		// 没有工具调用，直接回答
		if !hasToolCall {
			// 保存模型输出
			for _, item := range resp.Output {
				var inputItem responses.ResponseInputItemUnion

				err = json.Unmarshal([]byte(item.RawJSON()), &inputItem)
				if err != nil {
					return err.Error()
				}

				a.Session.AddMessage(inputItem.ToParam())
			}

			return resp.OutputText()
		}

		// 保存模型输出
		for _, item := range resp.Output {
			var inputItem responses.ResponseInputItemUnion
			err = json.Unmarshal([]byte(item.RawJSON()), &inputItem)
			if err != nil {
				return err.Error()
			}
			a.Session.AddMessage(inputItem.ToParam())
		}

		// 执行工具
		for _, item := range resp.Output {
			switch item.Type {
			case "function_call":
				call := item.AsFunctionCall()
				tool, ok := a.Tools[call.Name]
				if !ok {
					continue
				}

				result, err := tool.Call(ctx, call.Arguments)
				if err != nil {
					result = err.Error()
				}

				a.Session.AddMessage(
					responses.ResponseInputItemUnionParam{
						OfFunctionCallOutput: &responses.ResponseInputItemFunctionCallOutputParam{
							CallID: openai.String(call.CallID),
							Output: responses.ResponseInputItemFunctionCallOutputOutputUnionParam{
								OfString: openai.String(result),
							},
						},
					},
				)
			}
		}
	}

	return "达到最大循环次数"
}

// RegisterMCPTools 从 MCP Manager 管理的所有 Server 中发现工具，
// 并注册到 Agent。工具名称重复时返回错误。
func (a *Agent) RegisterMCPTools(
	ctx context.Context,
	manager *mcppkg.Manager,
) error {
	clients := manager.GetAllClients()

	for serverName, client := range clients {
		// 1. 发现该 Server 提供的工具
		mcpTools, err := mcppkg.DiscoverTools(ctx, client.Session)
		if err != nil {
			return fmt.Errorf("从 MCP Server %s 发现工具失败: %w", serverName, err)
		}

		// 2. 转换为项目自己的 ToolDef
		convertedTools, err := mcppkg.ConvertToToolDefs(client.Session, mcpTools)
		if err != nil {
			return err
		}

		// 3. 检查工具名称是否重复，避免覆盖原来的 Go 工具
		for _, tool := range convertedTools {
			if _, exists := a.Tools[tool.Name]; exists {
				return fmt.Errorf("工具名称重复: %s (来自 Server %s)", tool.Name, serverName)
			}
		}

		// 4. 注册到 Agent
		for _, tool := range convertedTools {
			a.Tools[tool.Name] = tool
			a.OpenaiTools = append(a.OpenaiTools, tool.ToParam())
		}
	}

	return nil
}