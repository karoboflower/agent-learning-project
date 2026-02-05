package forecaster

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/agent-learning/enterprise-platform/services/cost/internal/model"
	"github.com/agent-learning/enterprise-platform/services/cost/internal/repository"
)

// Forecaster 成本预测器
type Forecaster struct {
	repo *repository.CostRepository
}

// NewForecaster 创建预测器
func NewForecaster(repo *repository.CostRepository) *Forecaster {
	return &Forecaster{
		repo: repo,
	}
}

// ForecastNextMonth 预测下月成本
func (f *Forecaster) ForecastNextMonth(ctx context.Context, tenantID string) (*model.CostForecast, error) {
	// 获取历史数据（过去3个月）
	endDate := time.Now()
	startDate := endDate.AddDate(0, -3, 0)

	historicalData, err := f.repo.GetUsageHistory(ctx, tenantID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get historical data: %w", err)
	}

	if len(historicalData) < 7 {
		return nil, fmt.Errorf("insufficient historical data (need at least 7 days)")
	}

	// 使用线性回归预测
	forecast := f.linearRegression(historicalData)
	forecast.TenantID = tenantID
	forecast.ForecastDate = time.Now().AddDate(0, 1, 0)
	forecast.Method = "linear"

	return forecast, nil
}

// linearRegression 线性回归预测
func (f *Forecaster) linearRegression(data []*model.TokenUsage) *model.CostForecast {
	n := len(data)
	if n == 0 {
		return &model.CostForecast{Confidence: 0}
	}

	// 计算均值
	var sumX, sumY, sumXY, sumX2 float64
	for i, usage := range data {
		x := float64(i)
		y := usage.TotalCost

		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	// 计算斜率和截距
	slope := (float64(n)*sumXY - sumX*sumY) / (float64(n)*sumX2 - sumX*sumX)
	intercept := (sumY - slope*sumX) / float64(n)

	// 预测下个月（30天）的成本
	predictedDays := 30.0
	predictedCostPerDay := slope*float64(n) + intercept
	predictedTotalCost := predictedCostPerDay * predictedDays

	// 预测Token数（基于平均成本比例）
	avgCostPerToken := sumY / float64(n)
	var totalTokens int64
	for _, usage := range data {
		totalTokens += usage.TotalTokens
	}
	avgTokensPerDay := float64(totalTokens) / float64(n)
	predictedTokens := int64(avgTokensPerDay * predictedDays)

	// 计算置信度（基于R²）
	confidence := f.calculateR2(data, slope, intercept)

	return &model.CostForecast{
		PredictedTokens: predictedTokens,
		PredictedCost:   predictedTotalCost,
		Confidence:      confidence,
	}
}

// calculateR2 计算R²（决定系数）
func (f *Forecaster) calculateR2(data []*model.TokenUsage, slope, intercept float64) float64 {
	n := len(data)
	if n == 0 {
		return 0
	}

	// 计算平均值
	var sumY float64
	for _, usage := range data {
		sumY += usage.TotalCost
	}
	meanY := sumY / float64(n)

	// 计算总平方和和残差平方和
	var ssTot, ssRes float64
	for i, usage := range data {
		x := float64(i)
		y := usage.TotalCost
		predicted := slope*x + intercept

		ssTot += math.Pow(y-meanY, 2)
		ssRes += math.Pow(y-predicted, 2)
	}

	if ssTot == 0 {
		return 0
	}

	r2 := 1 - (ssRes / ssTot)
	return math.Max(0, math.Min(1, r2)) // 限制在0-1之间
}

// ExponentialSmoothing 指数平滑预测
func (f *Forecaster) ExponentialSmoothing(data []*model.TokenUsage, alpha float64) *model.CostForecast {
	if len(data) == 0 {
		return &model.CostForecast{Confidence: 0}
	}

	// 初始预测值为第一个数据点
	forecast := data[0].TotalCost

	// 应用指数平滑
	for i := 1; i < len(data); i++ {
		actual := data[i].TotalCost
		forecast = alpha*actual + (1-alpha)*forecast
	}

	// 预测下个月
	predictedCostPerDay := forecast
	predictedTotalCost := predictedCostPerDay * 30

	// 估算Token数
	var totalTokens int64
	for _, usage := range data {
		totalTokens += usage.TotalTokens
	}
	avgTokensPerDay := float64(totalTokens) / float64(len(data))
	predictedTokens := int64(avgTokensPerDay * 30)

	return &model.CostForecast{
		PredictedTokens: predictedTokens,
		PredictedCost:   predictedTotalCost,
		Confidence:      0.7, // 经验值
		Method:          "exponential",
	}
}

// DetectAnomaly 检测成本异常
func (f *Forecaster) DetectAnomaly(ctx context.Context, tenantID string) (bool, string, error) {
	// 获取过去7天的数据
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -7)

	data, err := f.repo.GetUsageHistory(ctx, tenantID, startDate, endDate)
	if err != nil {
		return false, "", err
	}

	if len(data) < 3 {
		return false, "", nil // 数据不足，无法判断
	}

	// 计算均值和标准差
	var sum, sumSq float64
	for _, usage := range data {
		sum += usage.TotalCost
		sumSq += usage.TotalCost * usage.TotalCost
	}

	n := float64(len(data))
	mean := sum / n
	variance := (sumSq / n) - (mean * mean)
	stdDev := math.Sqrt(variance)

	// 检查最后一天是否异常（超过2个标准差）
	lastDay := data[len(data)-1]
	threshold := mean + 2*stdDev

	if lastDay.TotalCost > threshold {
		message := fmt.Sprintf("检测到成本异常：今日成本 $%.2f 超过正常范围 (均值: $%.2f, 阈值: $%.2f)",
			lastDay.TotalCost, mean, threshold)
		return true, message, nil
	}

	return false, "", nil
}

// TrendAnalysis 趋势分析
func (f *Forecaster) TrendAnalysis(ctx context.Context, tenantID string, days int) (string, float64, error) {
	// 获取历史数据
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	data, err := f.repo.GetUsageHistory(ctx, tenantID, startDate, endDate)
	if err != nil {
		return "", 0, err
	}

	if len(data) < 2 {
		return "stable", 0, nil
	}

	// 计算线性回归的斜率
	n := len(data)
	var sumX, sumY, sumXY, sumX2 float64
	for i, usage := range data {
		x := float64(i)
		y := usage.TotalCost

		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	slope := (float64(n)*sumXY - sumX*sumY) / (float64(n)*sumX2 - sumX*sumX)

	// 计算变化率
	avgCost := sumY / float64(n)
	changeRate := (slope / avgCost) * 100

	// 判断趋势
	var trend string
	if changeRate > 10 {
		trend = "increasing"
	} else if changeRate < -10 {
		trend = "decreasing"
	} else {
		trend = "stable"
	}

	return trend, changeRate, nil
}

// SeasonalAnalysis 季节性分析
func (f *Forecaster) SeasonalAnalysis(data []*model.TokenUsage) map[string]float64 {
	if len(data) < 7 {
		return nil
	}

	// 按星期几分组
	weekdayCosts := make(map[time.Weekday][]float64)
	for _, usage := range data {
		weekday := usage.Timestamp.Weekday()
		weekdayCosts[weekday] = append(weekdayCosts[weekday], usage.TotalCost)
	}

	// 计算每个星期几的平均成本
	result := make(map[string]float64)
	for weekday, costs := range weekdayCosts {
		var sum float64
		for _, cost := range costs {
			sum += cost
		}
		avg := sum / float64(len(costs))
		result[weekday.String()] = avg
	}

	return result
}
