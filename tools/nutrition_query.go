package tools

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/openai/openai-go/v3"
)

// FoodRecord 对应 daily_food_records 表中的单条食物记录
type FoodRecord struct {
	MealType     string  `json:"meal_type"`
	FoodName     string  `json:"food_name"`
	Weight       float64 `json:"weight"`
	Calorie      float64 `json:"calorie"`
	Protein      float64 `json:"protein"`
	Fat          float64 `json:"fat"`
	Carbohydrate float64 `json:"carbohydrate"`
}

// NutritionSummary 当日营养汇总
type NutritionSummary struct {
	Calories     float64 `json:"calories"`
	Protein      float64 `json:"protein"`
	Fat          float64 `json:"fat"`
	Carbohydrate float64 `json:"carbohydrate"`
}

// WeeklyCalorie 近七天每日热量
type WeeklyCalorie struct {
	Day      string  `json:"day"`
	Calories float64 `json:"calories"`
}

func NewNutritionQueryTool(db *sql.DB) *ToolDef {
	return &ToolDef{
		Name:        "nutrition_query",
		Description: "查询指定日期的饮食营养摄入情况，返回当日所有食物明细、营养汇总和近七天热量趋势",
		Params: openai.FunctionParameters{
			"type": "object",
			"properties": map[string]any{
				"date": map[string]any{
					"type":        "string",
					"description": "查询日期，例如2026-10-13",
				},
			},
			"required": []string{
				"date",
			},
		},
		Func: func(ctx context.Context, args map[string]any) (string, error) {
			date, _ := args["date"].(string)

			// 1. 查询当日所有食物明细
			rows, err := db.QueryContext(ctx, `
				SELECT meal_type, food_name, weight, calorie, protein, fat, carbohydrate
				FROM daily_food_records
				WHERE eat_date = ? AND user_id = ?
				ORDER BY id ASC
			`, date, 1)
			if err != nil {
				return "", fmt.Errorf("查询食物明细失败: %w", err)
			}
			defer rows.Close()

			var records []FoodRecord
			for rows.Next() {
				var r FoodRecord
				if err := rows.Scan(&r.MealType, &r.FoodName, &r.Weight, &r.Calorie, &r.Protein, &r.Fat, &r.Carbohydrate); err != nil {
					return "", fmt.Errorf("扫描食物明细失败: %w", err)
				}
				records = append(records, r)
			}

			// 2. 查询当日营养汇总
			var summary NutritionSummary
			err = db.QueryRowContext(ctx, `
				SELECT
					COALESCE(SUM(calorie), 0),
					COALESCE(SUM(protein), 0),
					COALESCE(SUM(fat), 0),
					COALESCE(SUM(carbohydrate), 0)
				FROM daily_food_records
				WHERE eat_date = ? AND user_id = ?
			`, date, 1).Scan(&summary.Calories, &summary.Protein, &summary.Fat, &summary.Carbohydrate)
			if err != nil {
				return "", fmt.Errorf("查询营养汇总失败: %w", err)
			}

			// 3. 查询近七天热量趋势（含查询日期在内的最近7天）
			weeklyCalories, err := queryWeeklyCalories(ctx, db, date)
			if err != nil {
				return "", fmt.Errorf("查询近七天热量失败: %w", err)
			}

			result := map[string]any{
				"date":            date,
				"records":         records,
				"summary":         summary,
				"weeklyCalories":  weeklyCalories,
			}
			data, _ := json.Marshal(result)
			return string(data), nil
		},
	}
}

// queryWeeklyCalories 查询指定日期往前推6天共7天的每日热量
func queryWeeklyCalories(ctx context.Context, db *sql.DB, targetDate string) ([]WeeklyCalorie, error) {
	t, err := time.Parse("2006-01-02", targetDate)
	if err != nil {
		return nil, fmt.Errorf("日期格式错误: %w", err)
	}

	weekdayNames := []string{"周日", "周一", "周二", "周三", "周四", "周五", "周六"}
	result := make([]WeeklyCalorie, 0, 7)

	for i := 6; i >= 0; i-- {
		day := t.AddDate(0, 0, -i)
		dateStr := day.Format("2006-01-02")

		var calories float64
		err := db.QueryRowContext(ctx, `
			SELECT COALESCE(SUM(calorie), 0)
			FROM daily_food_records
			WHERE eat_date = ? AND user_id = ?
		`, dateStr, 1).Scan(&calories)
		if err != nil {
			return nil, err
		}

		result = append(result, WeeklyCalorie{
			Day:      weekdayNames[day.Weekday()],
			Calories: calories,
		})
	}

	return result, nil
}
