
// ===============================
// 1. 模拟营养数据
// ===============================

const nutritionData = {
    date: "2026-10-13",

    // 对应 daily_food_records 表中的食物记录
    records: [
        {
            id: 1,
            user_id: 1,
            eat_date: "2026-10-13",
            meal_type: "早餐",
            food_name: "白米饭饭团",
            weight: 110,
            calorie: 202,
            protein: 3.5,
            fat: 0.3,
            carbohydrate: 44.5
        },
        {
            id: 2,
            user_id: 1,
            eat_date: "2026-10-13",
            meal_type: "早餐",
            food_name: "紫米杂粮饭团",
            weight: 110,
            calorie: 215,
            protein: 4.5,
            fat: 0.8,
            carbohydrate: 46
        },
        {
            id: 3,
            user_id: 1,
            eat_date: "2026-10-13",
            meal_type: "早餐",
            food_name: "清炒四季豆",
            weight: 30,
            calorie: 25,
            protein: 1.2,
            fat: 1,
            carbohydrate: 3.5
        }
    ],

    // 网页顶部的营养统计
    summary: {
        calories: 442,
        protein: 9.2,
        fat: 2.1,
        carbohydrate: 94
    },

    weeklyCalories: [
        { day: "周一", calories: 1800 },
        { day: "周二", calories: 1950 },
        { day: "周三", calories: 1700 },
        { day: "周四", calories: 2100 },
        { day: "周五", calories: 1900 },
        { day: "周六", calories: 2200 },
        { day: "周日", calories: 1850 }
    ],

    good: [
        "本次记录中，碳水化合物主要来自米饭类食物。",
        "已经记录了多种食物，有利于了解饮食组成。"
    ],

    warning: [
        "当前只展示部分早餐记录，不能据此判断全天营养是否充足。"
    ],

    advice: [
        {
            title: "丰富食物种类",
            content: "可以搭配蔬菜、蛋类和优质蛋白质来源。"
        },
        {
            title: "完善饮食记录",
            content: "尽量记录当天每一餐，方便进行完整的营养分析。"
        }
    ]
};


// ===============================
// 2. 填充顶部营养数据卡片
// ===============================

function renderSummary() {
    const summary = nutritionData.summary;

    const cards = document.querySelectorAll(".nutrition-card");

    const values = [
        {
            name: "总热量",
            value: summary.calories,
            unit: "kcal"
        },
        {
            name: "蛋白质",
            value: summary.protein,
            unit: "g"
        },
        {
            name: "碳水化合物",
            value: summary.carbohydrate,
            unit: "g"
        },
        {
            name: "脂肪",
            value: summary.fat,
            unit: "g"
        }
    ];

    cards.forEach((card, index) => {
        const item = values[index];

        card.innerHTML = `
            <div>
                <p class="card-label">${item.name}</p>
                <div class="card-value">
                    ${item.value}
                    <span>${item.unit}</span>
                </div>
            </div>
        `;
    });
}


// ===============================
// 3. 营养亮点
// ===============================

function renderGood() {
    const container = document.getElementById("good-content");

    container.innerHTML = nutritionData.good.map(item => `
        <div class="analysis-item good-item">
            <span class="item-icon">✓</span>
            <span>${item}</span>
        </div>
    `).join("");
}


// ===============================
// 4. 需要改善
// ===============================

function renderWarning() {
    const container = document.getElementById("warning-content");

    container.innerHTML = nutritionData.warning.map(item => `
        <div class="analysis-item warning-item">
            <span class="item-icon">!</span>
            <span>${item}</span>
        </div>
    `).join("");
}


// ===============================
// 5. 饮食建议
// ===============================

function renderAdvice() {
    const container = document.getElementById("advice-content");

    container.innerHTML = nutritionData.advice.map(item => `
        <div class="advice-item">
            <h3>${item.title}</h3>
            <p>${item.content}</p>
        </div>
    `).join("");
}


// ===============================
// 6. 营养素占比环形图
// ===============================

function renderNutritionChart() {
    const summary = nutritionData.summary;

    const chart = echarts.init(
        document.getElementById("nutrition-chart")
    );

    // 按照三大营养素提供的热量计算占比
    const proteinCalories = summary.protein * 4;
    const carbohydrateCalories = summary.carbohydrate * 4;
    const fatCalories = summary.fat * 9;

    const option = {
        title: {
            text: "三大营养素供能比例",
            left: "center",
            top: 10,
            textStyle: {
                fontSize: 17,
                color: "#294638"
            }
        },

        tooltip: {
            trigger: "item",
            formatter: "{b}: {c} kcal ({d}%)"
        },

        legend: {
            bottom: 10,
            left: "center"
        },

        series: [
            {
                type: "pie",
                radius: ["45%", "68%"],
                center: ["50%", "52%"],

                data: [
                    {
                        value: proteinCalories,
                        name: "蛋白质",
                        itemStyle: { color: "#55B88A" }
                    },
                    {
                        value: carbohydrateCalories,
                        name: "碳水化合物",
                        itemStyle: { color: "#6CA9E8" }
                    },
                    {
                        value: fatCalories,
                        name: "脂肪",
                        itemStyle: { color: "#F0B65B" }
                    }
                ],

                label: {
                    formatter: "{d}%",
                    fontSize: 13
                }
            }
        ]
    };

    chart.setOption(option);

    return chart;
}


// ===============================
// 7. 最近七天热量趋势图
// ===============================

function renderCalorieChart() {
    const chart = echarts.init(
        document.getElementById("calorie-chart")
    );

    const days = nutritionData.weeklyCalories.map(
        item => item.day
    );

    const calories = nutritionData.weeklyCalories.map(
        item => item.calories
    );

    const option = {
        title: {
            text: "近七天热量摄入趋势",
            left: "center",
            top: 10,
            textStyle: {
                fontSize: 17,
                color: "#294638"
            }
        },

        tooltip: {
            trigger: "axis"
        },

        grid: {
            left: "10%",
            right: "8%",
            top: "22%",
            bottom: "15%"
        },

        xAxis: {
            type: "category",
            data: days
        },

        yAxis: {
            type: "value",
            name: "kcal"
        },

        series: [
            {
                name: "热量摄入",
                type: "line",
                data: calories,
                smooth: true,

                symbolSize: 8,

                lineStyle: {
                    width: 3,
                    color: "#55B88A"
                },

                itemStyle: {
                    color: "#55B88A"
                },

                areaStyle: {
                    color: "rgba(85, 184, 138, 0.12)"
                }
            }
        ]
    };

    chart.setOption(option);

    return chart;
}


// ===============================
// 8. 页面初始化
// ===============================

function initDashboard() {
    renderSummary();
    renderGood();
    renderWarning();
    renderAdvice();

    const charts = [
        renderNutritionChart(),
        renderCalorieChart()
    ];

    // 浏览器窗口大小变化时，自动调整图表
    window.addEventListener("resize", () => {
        charts.forEach(chart => chart.resize());
    });
}

// 页面加载完成后执行
document.addEventListener("DOMContentLoaded", initDashboard);
