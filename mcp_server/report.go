package main

import (
	"encoding/json"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ReportInput 是 generate_nutrition_report 工具接收的参数
type ReportInput struct {
	Date           string         `json:"date"`
	Records        []FoodRecord   `json:"records"`
	Summary        NutritionSum   `json:"summary"`
	WeeklyCalories []WeeklyCal    `json:"weeklyCalories"`
	Good           []string       `json:"good"`
	Warning        []string       `json:"warning"`
	Advice         []AdviceItem   `json:"advice"`
}

type FoodRecord struct {
	MealType     string  `json:"meal_type"`
	FoodName     string  `json:"food_name"`
	Weight       float64 `json:"weight"`
	Calorie      float64 `json:"calorie"`
	Protein      float64 `json:"protein"`
	Fat          float64 `json:"fat"`
	Carbohydrate float64 `json:"carbohydrate"`
}

type NutritionSum struct {
	Calories     float64 `json:"calories"`
	Protein      float64 `json:"protein"`
	Fat          float64 `json:"fat"`
	Carbohydrate float64 `json:"carbohydrate"`
}

type WeeklyCal struct {
	Day      string  `json:"day"`
	Calories float64 `json:"calories"`
}

type AdviceItem struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// ReportResult 是工具返回给 Agent 的结果
type ReportResult struct {
	Success    bool   `json:"success"`
	ReportName string `json:"report_name"`
	FilePath   string `json:"file_path"`
	Message    string `json:"message"`
}

// GenerateReport 生成营养分析 HTML 报告
// projectRoot: 项目根目录，用于定位 web 模板
// reportsDir: 报告输出目录
func GenerateReport(input ReportInput, projectRoot, reportsDir string) (*ReportResult, error) {
	// 1. 验证必要字段
	if input.Date == "" {
		return nil, fmt.Errorf("date 字段不能为空")
	}

	// 2. 生成报告目录名，避免路径穿越
	safeDate := strings.ReplaceAll(input.Date, "/", "-")
	safeDate = strings.ReplaceAll(safeDate, "\\", "-")
	timestamp := time.Now().Format("20060102-150405")
	reportName := fmt.Sprintf("nutrition-%s-%s", safeDate, timestamp)

	reportDir := filepath.Join(reportsDir, reportName)

	// 3. 读取原始 HTML 模板
	webDir := filepath.Join(projectRoot, "web")
	htmlPath := filepath.Join(webDir, "index.html")
	htmlContent, err := os.ReadFile(htmlPath)
	if err != nil {
		return nil, fmt.Errorf("读取 HTML 模板失败: %w", err)
	}

	// 4. 构建 nutritionData JSON（用于注入到页面）
	nutritionData := buildNutritionDataJSON(input)

	// 5. 将数据注入 HTML：替换 nutrition.js 中的模拟数据
	// 我们复制 nutrition.js 到报告目录，然后在 HTML 中通过内嵌脚本注入数据
	// 更简单的方式：直接在 HTML 中内嵌 nutritionData 变量

	htmlStr := string(htmlContent)

	// 将 ECharts CDN 替换为本地引用（如果本地存在）
	echartsLocalPath := filepath.Join(webDir, "assets", "echarts.min.js")
	if _, err := os.Stat(echartsLocalPath); err == nil {
		htmlStr = strings.Replace(
			htmlStr,
			`<script src="https://cdn.jsdelivr.net/npm/echarts/dist/echarts.min.js"></script>`,
			`<script src="./assets/echarts.min.js"></script>`,
			1,
		)
	}

	// 在 nutrition.js 之前注入 nutritionData 数据脚本，确保 JS 执行时数据已可用
	dataScript := fmt.Sprintf(`<script>window.__NUTRITION_DATA__ = %s;</script>`, nutritionData)
	htmlStr = strings.Replace(
		htmlStr,
		`<script src="./js/nutrition.js"></script>`,
		dataScript+"\n  <script src=\"./js/nutrition.js\"></script>",
		1,
	)

	// 6. 创建报告目录
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		return nil, fmt.Errorf("创建报告目录失败: %w", err)
	}

	// 7. 写入 index.html
	outHTMLPath := filepath.Join(reportDir, "index.html")
	if err := os.WriteFile(outHTMLPath, []byte(htmlStr), 0644); err != nil {
		return nil, fmt.Errorf("写入 HTML 文件失败: %w", err)
	}

	// 8. 复制 CSS
	cssSrc := filepath.Join(webDir, "css", "style.css")
	cssDst := filepath.Join(reportDir, "css", "style.css")
	if err := copyFile(cssSrc, cssDst); err != nil {
		return nil, fmt.Errorf("复制 CSS 失败: %w", err)
	}

	// 9. 复制 JS（修改后的 nutrition.js，从 window.__NUTRITION_DATA__ 读取数据）
	jsSrc := filepath.Join(webDir, "js", "nutrition.js")
	jsContent, err := os.ReadFile(jsSrc)
	if err != nil {
		return nil, fmt.Errorf("读取 JS 失败: %w", err)
	}

	// 修改 JS：从 window.__NUTRITION_DATA__ 读取数据，而不是使用模拟数据
	jsStr := string(jsContent)
	jsStr = replaceNutritionData(jsStr)

	jsDst := filepath.Join(reportDir, "js", "nutrition.js")
	if err := os.MkdirAll(filepath.Dir(jsDst), 0755); err != nil {
		return nil, fmt.Errorf("创建 JS 目录失败: %w", err)
	}
	if err := os.WriteFile(jsDst, []byte(jsStr), 0644); err != nil {
		return nil, fmt.Errorf("写入 JS 文件失败: %w", err)
	}

	// 10. 复制 ECharts（如果存在）
	if _, err := os.Stat(echartsLocalPath); err == nil {
		echartsDst := filepath.Join(reportDir, "assets", "echarts.min.js")
		if err := copyFile(echartsLocalPath, echartsDst); err != nil {
			return nil, fmt.Errorf("复制 ECharts 失败: %w", err)
		}
	}

	// 11. 检查文件是否成功创建
	if _, err := os.Stat(outHTMLPath); err != nil {
		return nil, fmt.Errorf("报告文件未成功创建: %w", err)
	}

	return &ReportResult{
		Success:    true,
		ReportName: reportName,
		FilePath:   outHTMLPath,
		Message:    "营养分析报告生成成功",
	}, nil
}

// buildNutritionDataJSON 将输入数据转换为网页需要的 JSON 格式
func buildNutritionDataJSON(input ReportInput) string {
	// 确保 records 至少为空数组
	records := input.Records
	if records == nil {
		records = []FoodRecord{}
	}

	weeklyCalories := input.WeeklyCalories
	if weeklyCalories == nil {
		weeklyCalories = []WeeklyCal{}
	}

	good := input.Good
	if good == nil {
		good = []string{}
	}

	warning := input.Warning
	if warning == nil {
		warning = []string{}
	}

	advice := input.Advice
	if advice == nil {
		advice = []AdviceItem{}
	}

	data := map[string]any{
		"date":           input.Date,
		"records":        records,
		"summary":        input.Summary,
		"weeklyCalories": weeklyCalories,
		"good":           good,
		"warning":        warning,
		"advice":         advice,
	}

	jsonBytes, _ := json.Marshal(data)
	return string(jsonBytes)
}

// replaceNutritionData 修改 nutrition.js，使其从 window.__NUTRITION_DATA__ 读取数据
func replaceNutritionData(js string) string {
	// 统一换行符，兼容 Windows \r\n
	js = strings.ReplaceAll(js, "\r\n", "\n")

	// 找到 const nutritionData = { 的起始位置
	const marker = "const nutritionData = {"
	idx := strings.Index(js, marker)
	if idx == -1 {
		return "const nutritionData = window.__NUTRITION_DATA__ || {};\n" + js
	}

	// 找到 nutritionData 对象的闭合：}; 后面跟着空行（下一行是注释）
	// 只匹配 "};\n\n\n"，保留后面的 "// =====" 注释不被消耗
	endMarkers := []string{
		"};\n\n\n",
		"};\n\n",
	}

	endIdx := -1
	var matchedMarker string
	for _, em := range endMarkers {
		if i := strings.Index(js[idx:], em); i != -1 {
			endIdx = i
			matchedMarker = em
			break
		}
	}

	if endIdx == -1 {
		return js
	}

	// 替换整个 nutritionData 定义，保留原有的空行和注释结构
	prefix := js[:idx]
	suffix := js[idx+endIdx+len(matchedMarker):]
	replacement := "const nutritionData = window.__NUTRITION_DATA__ || {};"
	return prefix + replacement + suffix
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

// escapeHTML 转义 HTML 特殊字符（用于安全输出）
func escapeHTML(s string) string {
	return html.EscapeString(s)
}
