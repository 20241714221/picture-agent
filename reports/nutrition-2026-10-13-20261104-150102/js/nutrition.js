
// ===============================
// 1. 模拟营养数据
// ===============================

const nutritionData = window.__NUTRITION_DATA__ || {};// ===============================
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
