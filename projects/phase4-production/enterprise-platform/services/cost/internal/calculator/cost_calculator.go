package calculator

import (
	"fmt"

	"github.com/agent-learning/enterprise-platform/services/cost/internal/model"
)

// CostCalculator 成本计算器
type CostCalculator struct {
	pricingMap map[string]model.ModelPricing
}

// NewCostCalculator 创建成本计算器
func NewCostCalculator() *CostCalculator {
	return &CostCalculator{
		pricingMap: model.GetModelPricing(),
	}
}

// Calculate 计算Token使用成本
func (c *CostCalculator) Calculate(modelName string, inputTokens, outputTokens int64) (*model.TokenUsage, error) {
	pricing, ok := c.pricingMap[modelName]
	if !ok {
		return nil, fmt.Errorf("unknown model: %s", modelName)
	}

	usage := &model.TokenUsage{
		Model:        modelName,
		Provider:     pricing.Provider,
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
	}

	usage.CalculateCost(pricing)

	return usage, nil
}

// CompareCosts 比较不同模型的成本
func (c *CostCalculator) CompareCosts(inputTokens, outputTokens int64) []model.ModelPricing {
	comparisons := make([]model.ModelPricing, 0, len(c.pricingMap))

	for _, pricing := range c.pricingMap {
		comparisons = append(comparisons, pricing)
	}

	return comparisons
}

// GetCheapestModel 获取最便宜的模型
func (c *CostCalculator) GetCheapestModel(inputTokens, outputTokens int64) (string, float64, error) {
	if len(c.pricingMap) == 0 {
		return "", 0, fmt.Errorf("no models available")
	}

	var cheapestModel string
	var lowestCost float64 = -1

	for modelName, pricing := range c.pricingMap {
		inputCost := float64(inputTokens) / 1000.0 * pricing.InputPricePerK
		outputCost := float64(outputTokens) / 1000.0 * pricing.OutputPricePerK
		totalCost := inputCost + outputCost

		if lowestCost == -1 || totalCost < lowestCost {
			lowestCost = totalCost
			cheapestModel = modelName
		}
	}

	return cheapestModel, lowestCost, nil
}

// EstimateMonthlyCost 估算月度成本
func (c *CostCalculator) EstimateMonthlyCost(modelName string, dailyRequests int64, avgInputTokens, avgOutputTokens int64) (float64, error) {
	pricing, ok := c.pricingMap[modelName]
	if !ok {
		return 0, fmt.Errorf("unknown model: %s", modelName)
	}

	// 计算单次请求成本
	inputCost := float64(avgInputTokens) / 1000.0 * pricing.InputPricePerK
	outputCost := float64(avgOutputTokens) / 1000.0 * pricing.OutputPricePerK
	costPerRequest := inputCost + outputCost

	// 估算月度成本（30天）
	monthlyCost := costPerRequest * float64(dailyRequests) * 30

	return monthlyCost, nil
}

// CalculateSavings 计算切换模型可节省的成本
func (c *CostCalculator) CalculateSavings(currentModel, targetModel string, inputTokens, outputTokens int64) (float64, float64, error) {
	currentPricing, ok := c.pricingMap[currentModel]
	if !ok {
		return 0, 0, fmt.Errorf("unknown current model: %s", currentModel)
	}

	targetPricing, ok := c.pricingMap[targetModel]
	if !ok {
		return 0, 0, fmt.Errorf("unknown target model: %s", targetModel)
	}

	// 计算当前成本
	currentInputCost := float64(inputTokens) / 1000.0 * currentPricing.InputPricePerK
	currentOutputCost := float64(outputTokens) / 1000.0 * currentPricing.OutputPricePerK
	currentCost := currentInputCost + currentOutputCost

	// 计算目标成本
	targetInputCost := float64(inputTokens) / 1000.0 * targetPricing.InputPricePerK
	targetOutputCost := float64(outputTokens) / 1000.0 * targetPricing.OutputPricePerK
	targetCost := targetInputCost + targetOutputCost

	// 计算节省
	savings := currentCost - targetCost
	savingsPercent := 0.0
	if currentCost > 0 {
		savingsPercent = (savings / currentCost) * 100
	}

	return savings, savingsPercent, nil
}

// OptimizationRecommendations 生成成本优化建议
func (c *CostCalculator) OptimizationRecommendations(currentModel string, totalTokens, totalRequests int64, currentCost float64, cacheHitRate float64) []model.CostOptimization {
	recommendations := make([]model.CostOptimization, 0)

	// 建议1: 模型切换
	cheapestModel, cheapestCost, _ := c.GetCheapestModel(totalTokens/totalRequests, totalTokens/totalRequests)
	if cheapestModel != currentModel && cheapestCost < currentCost/float64(totalRequests) {
		estimatedNewCost := cheapestCost * float64(totalRequests)
		savings := currentCost - estimatedNewCost
		savingsPercent := (savings / currentCost) * 100

		recommendations = append(recommendations, model.CostOptimization{
			Type:           "model_switch",
			Title:          fmt.Sprintf("切换到 %s 模型", cheapestModel),
			Description:    fmt.Sprintf("当前使用 %s，建议切换到成本更低的 %s", currentModel, cheapestModel),
			CurrentCost:    currentCost,
			EstimatedCost:  estimatedNewCost,
			Savings:        savings,
			SavingsPercent: savingsPercent,
			Priority:       c.getPriority(savingsPercent),
		})
	}

	// 建议2: 提高缓存命中率
	if cacheHitRate < 50.0 {
		targetHitRate := 70.0
		potentialSavings := currentCost * ((targetHitRate - cacheHitRate) / 100.0)

		recommendations = append(recommendations, model.CostOptimization{
			Type:           "cache",
			Title:          "提高缓存命中率",
			Description:    fmt.Sprintf("当前缓存命中率 %.1f%%，建议优化到 %.1f%%", cacheHitRate, targetHitRate),
			CurrentCost:    currentCost,
			EstimatedCost:  currentCost - potentialSavings,
			Savings:        potentialSavings,
			SavingsPercent: ((targetHitRate - cacheHitRate) / 100.0) * 100,
			Priority:       c.getPriority(((targetHitRate - cacheHitRate) / 100.0) * 100),
		})
	}

	// 建议3: 批量处理
	if totalRequests > 1000 {
		batchSavings := currentCost * 0.15 // 假设批量处理可节省15%

		recommendations = append(recommendations, model.CostOptimization{
			Type:           "batch",
			Title:          "使用批量处理",
			Description:    "将多个小请求合并为批量请求，减少API调用次数",
			CurrentCost:    currentCost,
			EstimatedCost:  currentCost - batchSavings,
			Savings:        batchSavings,
			SavingsPercent: 15.0,
			Priority:       "medium",
		})
	}

	// 建议4: Prompt优化
	avgTokensPerRequest := totalTokens / totalRequests
	if avgTokensPerRequest > 3000 {
		optimizationSavings := currentCost * 0.20 // 假设Prompt优化可节省20%

		recommendations = append(recommendations, model.CostOptimization{
			Type:           "prompt_optimization",
			Title:          "优化Prompt长度",
			Description:    fmt.Sprintf("平均每次请求 %d tokens，建议优化Prompt减少Token消耗", avgTokensPerRequest),
			CurrentCost:    currentCost,
			EstimatedCost:  currentCost - optimizationSavings,
			Savings:        optimizationSavings,
			SavingsPercent: 20.0,
			Priority:       "high",
		})
	}

	return recommendations
}

// getPriority 根据节省百分比确定优先级
func (c *CostCalculator) getPriority(savingsPercent float64) string {
	if savingsPercent >= 30 {
		return "high"
	} else if savingsPercent >= 15 {
		return "medium"
	}
	return "low"
}

// GetModelPricing 获取模型定价
func (c *CostCalculator) GetModelPricing(modelName string) (model.ModelPricing, error) {
	pricing, ok := c.pricingMap[modelName]
	if !ok {
		return model.ModelPricing{}, fmt.Errorf("unknown model: %s", modelName)
	}
	return pricing, nil
}

// ListModels 列出所有模型
func (c *CostCalculator) ListModels() []model.ModelPricing {
	models := make([]model.ModelPricing, 0, len(c.pricingMap))
	for _, pricing := range c.pricingMap {
		models = append(models, pricing)
	}
	return models
}

// GetProviderModels 获取指定提供商的所有模型
func (c *CostCalculator) GetProviderModels(provider string) []model.ModelPricing {
	models := make([]model.ModelPricing, 0)
	for _, pricing := range c.pricingMap {
		if pricing.Provider == provider {
			models = append(models, pricing)
		}
	}
	return models
}
