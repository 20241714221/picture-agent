package tools

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"io"
	"net/http"
	"strings"

	"github.com/openai/openai-go/v3"
)

type FoodItem struct {
	Name         string  `json:"name"`
	Weight       float64 `json:"weight"`
	Calorie     float64 `json:"calorie"`
	Protein      float64 `json:"protein"`
	Carbohydrate float64 `json:"carbohydrate"`
	Fat          float64 `json:"fat"`
}
type FoodAnalysis struct {
	Foods []FoodItem `json:"foods"`
}

func NewFoodRecognitionTool(visionClient *openai.Client) *ToolDef {
	return &ToolDef{
		//AI 调用工具时，需要传：一个对象
		Name:        "food_recognition",
		Description: "分析食物图片，识别食物种类并估算热量",
		Params: openai.FunctionParameters{
			"type": "object",
			"properties": map[string]any{
				"image_path": map[string]any{
					"type":        "string",
					"description": "图片路径",
				},
			},
			"required": []string{
				"image_path",
			},
		},
		Func: func(ctx context.Context, args map[string]any) (string, error) {
			imagePath, ok := args["image_path"].(string)
			if !ok {
				return "", fmt.Errorf("image_path 参数错误")
			}
			fmt.Println("准备分析图片:", imagePath)

			// 读取图片
			// 读取图片：支持本地路径和网络 URL
			data, err := loadImage(ctx, imagePath)
			if err != nil {
				return "", fmt.Errorf("读取图片失败: %w", err)
			}
			encoded := base64.StdEncoding.EncodeToString(data)

			contentType := http.DetectContentType(data)

			imageURL := "data:" + contentType + ";base64," + encoded
			resp, err := visionClient.Chat.Completions.New(
				ctx,
				openai.ChatCompletionNewParams{
					Model: "qwen3.8-flash",
					Messages: []openai.ChatCompletionMessageParamUnion{
						{
							OfUser: &openai.ChatCompletionUserMessageParam{
								Content: openai.ChatCompletionUserMessageParamContentUnion{
									OfArrayOfContentParts: []openai.ChatCompletionContentPartUnionParam{
										//文字
										{
											OfText: &openai.ChatCompletionContentPartTextParam{
												Text: `
                      你是一个食物识别模型。
                      分析图片中的食物，只返回JSON。
                      不要输出Markdown。
                      不要输出解释文字。
                      返回格式:
                      {
                       "foods":[
                         {
                          "name":"食物名称",
                          "weight":100,
                          "calorie":200,
                          "protein":10,
                          "carbohydrate":30,
                          "fat":5
                         }
                       ]
                      }
                      规则:
                      1. 识别图片中的所有主要食物。
                      2. 估算每种食物重量（单位g）。
                      3. 估算每种食物营养数据。
                      4. 不确定时给合理估计。
                      5. 不要判断餐次。
                      6. 不要生成饮食建议。
                      7. 不要生成总结。`,
											},
										},
										//图片
										{
											OfImageURL: &openai.ChatCompletionContentPartImageParam{
												ImageURL: openai.ChatCompletionContentPartImageImageURLParam{
													URL: imageURL,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			)

			if err != nil {
				return "",
					fmt.Errorf(
						"调用视觉模型失败:%v",
						err,
					)
			}

			if len(resp.Choices) == 0 {
				return "",
					fmt.Errorf(
						"没有返回内容",
					)
			}

			content := resp.Choices[0].Message.Content

			var result FoodAnalysis

			err = json.Unmarshal(
				[]byte(content),
				&result,
			)

			if err != nil {
				return "",
					fmt.Errorf(
						"JSON解析失败:%v 原始内容:%s",
						err,
						content,
					)
			}

			jsondata, err := json.Marshal(result)

			if err != nil {
				return "",
					err
			}

			return string(jsondata), nil
		}}

}
func loadImage(ctx context.Context, imagePath string) ([]byte, error) {
	// 如果是网络图片
	if strings.HasPrefix(imagePath, "http://") ||
		strings.HasPrefix(imagePath, "https://") {

		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			imagePath,
			nil,
		)
		if err != nil {
			return nil, err
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("下载图片失败，HTTP 状态码：%d", resp.StatusCode)
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("读取网络图片内容失败: %w", err)
		}
		return data, nil
	}

	// 否则按本地文件路径读取
	return os.ReadFile(imagePath)
}
