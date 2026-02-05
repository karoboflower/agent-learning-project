package model

import (
	"context"
	"fmt"
	"time"
)

// PerformanceAnalyzer 性能分析器
type PerformanceAnalyzer struct {
	metricsHistory []PerformanceMetrics
	maxHistory     int
}

// NewPerformanceAnalyzer 创建性能分析器
func NewPerformanceAnalyzer(maxHistory int) *PerformanceAnalyzer {
	return &PerformanceAnalyzer{
		metricsHistory: make([]PerformanceMetrics, 0, maxHistory),
		maxHistory:     maxHistory,
	}
}

// RecordMetrics 记录性能指标
func (pa *PerformanceAnalyzer) RecordMetrics(metrics PerformanceMetrics) {
	pa.metricsHistory = append(pa.metricsHistory, metrics)

	// 保持历史记录数量
	if len(pa.metricsHistory) > pa.maxHistory {
		pa.metricsHistory = pa.metricsHistory[1:]
	}
}

// AnalyzePerformance 分析性能
func (pa *PerformanceAnalyzer) AnalyzePerformance(tenantID string) []OptimizationSuggestion {
	if len(pa.metricsHistory) == 0 {
		return nil
	}

	suggestions := make([]OptimizationSuggestion, 0)

	// 获取最新指标
	latest := pa.metricsHistory[len(pa.metricsHistory)-1]

	// 分析响应时间
	if latest.P95ResponseTime > 1*time.Second {
		suggestions = append(suggestions, OptimizationSuggestion{
			ID:          fmt.Sprintf("latency-%d", time.Now().Unix()),
			TenantID:    tenantID,
			Category:    "latency",
			Title:       "响应时间过高",
			Description: fmt.Sprintf("P95响应时间 %v 超过阈值 1s，建议优化", latest.P95ResponseTime),
			Impact:      "high",
			Effort:      "medium",
			Priority:    8,
			Metrics: map[string]interface{}{
				"p95_response_time": latest.P95ResponseTime.Milliseconds(),
				"threshold":         1000,
			},
			CreatedAt: time.Now(),
		})
	}

	// 分析错误率
	if latest.ErrorRate > 5.0 {
		suggestions = append(suggestions, OptimizationSuggestion{
			ID:          fmt.Sprintf("error-%d", time.Now().Unix()),
			TenantID:    tenantID,
			Category:    "reliability",
			Title:       "错误率过高",
			Description: fmt.Sprintf("错误率 %.2f%% 超过阈值 5%%，建议检查服务健康", latest.ErrorRate),
			Impact:      "high",
			Effort:      "high",
			Priority:    10,
			Metrics: map[string]interface{}{
				"error_rate": latest.ErrorRate,
				"threshold":  5.0,
			},
			CreatedAt: time.Now(),
		})
	}

	// 分析CPU使用率
	if latest.CPUUsage > 80.0 {
		suggestions = append(suggestions, OptimizationSuggestion{
			ID:          fmt.Sprintf("cpu-%d", time.Now().Unix()),
			TenantID:    tenantID,
			Category:    "resource",
			Title:       "CPU使用率过高",
			Description: fmt.Sprintf("CPU使用率 %.2f%% 超过阈值 80%%，建议扩容或优化", latest.CPUUsage),
			Impact:      "high",
			Effort:      "low",
			Priority:    9,
			Metrics: map[string]interface{}{
				"cpu_usage": latest.CPUUsage,
				"threshold": 80.0,
			},
			CreatedAt: time.Now(),
		})
	}

	// 分析内存使用率
	if latest.MemoryUsage > 85.0 {
		suggestions = append(suggestions, OptimizationSuggestion{
			ID:          fmt.Sprintf("memory-%d", time.Now().Unix()),
			TenantID:    tenantID,
			Category:    "resource",
			Title:       "内存使用率过高",
			Description: fmt.Sprintf("内存使用率 %.2f%% 超过阈值 85%%，建议扩容或优化", latest.MemoryUsage),
			Impact:      "high",
			Effort:      "low",
			Priority:    9,
			Metrics: map[string]interface{}{
				"memory_usage": latest.MemoryUsage,
				"threshold":    85.0,
			},
			CreatedAt: time.Now(),
		})
	}

	// 分析缓存命中率
	if latest.CacheHitRate < 50.0 {
		suggestions = append(suggestions, OptimizationSuggestion{
			ID:          fmt.Sprintf("cache-%d", time.Now().Unix()),
			TenantID:    tenantID,
			Category:    "cost",
			Title:       "缓存命中率过低",
			Description: fmt.Sprintf("缓存命中率 %.2f%% 低于目标 70%%，建议优化缓存策略", latest.CacheHitRate),
			Impact:      "medium",
			Effort:      "medium",
			Priority:    6,
			Metrics: map[string]interface{}{
				"cache_hit_rate": latest.CacheHitRate,
				"target":         70.0,
			},
			CreatedAt: time.Now(),
		})
	}

	// 分析吞吐量
	if latest.Throughput < 100.0 {
		suggestions = append(suggestions, OptimizationSuggestion{
			ID:          fmt.Sprintf("throughput-%d", time.Now().Unix()),
			TenantID:    tenantID,
			Category:    "throughput",
			Title:       "吞吐量偏低",
			Description: fmt.Sprintf("吞吐量 %.2f RPS 低于预期，建议优化性能", latest.Throughput),
			Impact:      "medium",
			Effort:      "high",
			Priority:    5,
			Metrics: map[string]interface{}{
				"throughput": latest.Throughput,
				"target":     500.0,
			},
			CreatedAt: time.Now(),
		})
	}

	// 分析队列积压
	if latest.QueuedTasks > 1000 {
		suggestions = append(suggestions, OptimizationSuggestion{
			ID:          fmt.Sprintf("queue-%d", time.Now().Unix()),
			TenantID:    tenantID,
			Category:    "throughput",
			Title:       "任务队列积压",
			Description: fmt.Sprintf("队列中有 %d 个任务等待处理，建议增加处理能力", latest.QueuedTasks),
			Impact:      "high",
			Effort:      "low",
			Priority:    8,
			Metrics: map[string]interface{}{
				"queued_tasks": latest.QueuedTasks,
				"threshold":    1000,
			},
			CreatedAt: time.Now(),
		})
	}

	return suggestions
}

// GetTrend 获取趋势
func (pa *PerformanceAnalyzer) GetTrend(metric string, duration time.Duration) (string, float64) {
	if len(pa.metricsHistory) < 2 {
		return "stable", 0
	}

	// 获取时间范围内的数据
	cutoff := time.Now().Add(-duration)
	var values []float64

	for _, m := range pa.metricsHistory {
		if m.Timestamp.After(cutoff) {
			switch metric {
			case "response_time":
				values = append(values, float64(m.AvgResponseTime.Milliseconds()))
			case "error_rate":
				values = append(values, m.ErrorRate)
			case "cpu_usage":
				values = append(values, m.CPUUsage)
			case "memory_usage":
				values = append(values, m.MemoryUsage)
			case "throughput":
				values = append(values, m.Throughput)
			case "cache_hit_rate":
				values = append(values, m.CacheHitRate)
			}
		}
	}

	if len(values) < 2 {
		return "stable", 0
	}

	// 计算趋势
	first := values[0]
	last := values[len(values)-1]
	change := ((last - first) / first) * 100

	var trend string
	if change > 10 {
		trend = "increasing"
	} else if change < -10 {
		trend = "decreasing"
	} else {
		trend = "stable"
	}

	return trend, change
}

// GetAverageMetrics 获取平均指标
func (pa *PerformanceAnalyzer) GetAverageMetrics(duration time.Duration) *PerformanceMetrics {
	if len(pa.metricsHistory) == 0 {
		return nil
	}

	cutoff := time.Now().Add(-duration)
	var count int
	var totalResponseTime time.Duration
	var totalErrorRate float64
	var totalCPU float64
	var totalMemory float64
	var totalThroughput float64
	var totalCacheHitRate float64

	for _, m := range pa.metricsHistory {
		if m.Timestamp.After(cutoff) {
			count++
			totalResponseTime += m.AvgResponseTime
			totalErrorRate += m.ErrorRate
			totalCPU += m.CPUUsage
			totalMemory += m.MemoryUsage
			totalThroughput += m.Throughput
			totalCacheHitRate += m.CacheHitRate
		}
	}

	if count == 0 {
		return nil
	}

	return &PerformanceMetrics{
		AvgResponseTime: totalResponseTime / time.Duration(count),
		ErrorRate:       totalErrorRate / float64(count),
		CPUUsage:        totalCPU / float64(count),
		MemoryUsage:     totalMemory / float64(count),
		Throughput:      totalThroughput / float64(count),
		CacheHitRate:    totalCacheHitRate / float64(count),
		Timestamp:       time.Now(),
	}
}

// DetectAnomalies 检测异常
func (pa *PerformanceAnalyzer) DetectAnomalies() []string {
	if len(pa.metricsHistory) < 10 {
		return nil
	}

	anomalies := make([]string, 0)
	latest := pa.metricsHistory[len(pa.metricsHistory)-1]

	// 计算历史平均值和标准差
	avgMetrics := pa.GetAverageMetrics(24 * time.Hour)
	if avgMetrics == nil {
		return nil
	}

	// 检测响应时间异常（超过2倍标准差）
	if latest.AvgResponseTime > avgMetrics.AvgResponseTime*2 {
		anomalies = append(anomalies, fmt.Sprintf("响应时间异常：当前 %v，平均 %v",
			latest.AvgResponseTime, avgMetrics.AvgResponseTime))
	}

	// 检测错误率异常
	if latest.ErrorRate > avgMetrics.ErrorRate*3 {
		anomalies = append(anomalies, fmt.Sprintf("错误率异常：当前 %.2f%%，平均 %.2f%%",
			latest.ErrorRate, avgMetrics.ErrorRate))
	}

	// 检测CPU使用率异常
	if latest.CPUUsage > avgMetrics.CPUUsage*1.5 {
		anomalies = append(anomalies, fmt.Sprintf("CPU使用率异常：当前 %.2f%%，平均 %.2f%%",
			latest.CPUUsage, avgMetrics.CPUUsage))
	}

	return anomalies
}

// GenerateReport 生成性能报告
func (pa *PerformanceAnalyzer) GenerateReport(ctx context.Context, tenantID string, duration time.Duration) map[string]interface{} {
	avgMetrics := pa.GetAverageMetrics(duration)
	if avgMetrics == nil {
		return nil
	}

	latest := pa.metricsHistory[len(pa.metricsHistory)-1]

	report := map[string]interface{}{
		"tenant_id": tenantID,
		"period":    duration.String(),
		"generated_at": time.Now(),
		"current_metrics": map[string]interface{}{
			"avg_response_time": latest.AvgResponseTime.Milliseconds(),
			"p95_response_time": latest.P95ResponseTime.Milliseconds(),
			"p99_response_time": latest.P99ResponseTime.Milliseconds(),
			"error_rate":        latest.ErrorRate,
			"throughput":        latest.Throughput,
			"cpu_usage":         latest.CPUUsage,
			"memory_usage":      latest.MemoryUsage,
			"cache_hit_rate":    latest.CacheHitRate,
			"active_connections": latest.ActiveConnections,
			"queued_tasks":      latest.QueuedTasks,
		},
		"average_metrics": map[string]interface{}{
			"avg_response_time": avgMetrics.AvgResponseTime.Milliseconds(),
			"error_rate":        avgMetrics.ErrorRate,
			"throughput":        avgMetrics.Throughput,
			"cpu_usage":         avgMetrics.CPUUsage,
			"memory_usage":      avgMetrics.MemoryUsage,
			"cache_hit_rate":    avgMetrics.CacheHitRate,
		},
		"trends": map[string]interface{}{},
		"anomalies": pa.DetectAnomalies(),
		"suggestions": pa.AnalyzePerformance(tenantID),
	}

	// 添加趋势分析
	metrics := []string{"response_time", "error_rate", "cpu_usage", "memory_usage", "throughput", "cache_hit_rate"}
	trends := make(map[string]interface{})

	for _, metric := range metrics {
		trend, change := pa.GetTrend(metric, duration)
		trends[metric] = map[string]interface{}{
			"trend":  trend,
			"change": fmt.Sprintf("%.2f%%", change),
		}
	}

	report["trends"] = trends

	return report
}
