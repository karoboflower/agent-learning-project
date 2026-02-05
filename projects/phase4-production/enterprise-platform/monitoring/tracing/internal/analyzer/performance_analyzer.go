package analyzer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"
)

// JaegerClient Jaeger客户端
type JaegerClient struct {
	baseURL string
	client  *http.Client
}

// NewJaegerClient 创建Jaeger客户端
func NewJaegerClient(baseURL string) *JaegerClient {
	return &JaegerClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Trace 追踪数据
type Trace struct {
	TraceID   string  `json:"traceID"`
	Spans     []Span  `json:"spans"`
	Processes map[string]Process `json:"processes"`
}

// Span Span数据
type Span struct {
	TraceID       string      `json:"traceID"`
	SpanID        string      `json:"spanID"`
	OperationName string      `json:"operationName"`
	References    []Reference `json:"references"`
	StartTime     int64       `json:"startTime"`
	Duration      int64       `json:"duration"`
	Tags          []Tag       `json:"tags"`
	Logs          []Log       `json:"logs"`
	ProcessID     string      `json:"processID"`
}

// Reference Span引用
type Reference struct {
	RefType string `json:"refType"`
	TraceID string `json:"traceID"`
	SpanID  string `json:"spanID"`
}

// Tag 标签
type Tag struct {
	Key   string      `json:"key"`
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

// Log 日志
type Log struct {
	Timestamp int64 `json:"timestamp"`
	Fields    []Tag `json:"fields"`
}

// Process 进程
type Process struct {
	ServiceName string `json:"serviceName"`
	Tags        []Tag  `json:"tags"`
}

// GetTrace 获取追踪
func (jc *JaegerClient) GetTrace(ctx context.Context, traceID string) (*Trace, error) {
	url := fmt.Sprintf("%s/api/traces/%s", jc.baseURL, traceID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := jc.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get trace: %s", string(body))
	}

	var result struct {
		Data []Trace `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, fmt.Errorf("trace not found: %s", traceID)
	}

	return &result.Data[0], nil
}

// SearchTraces 搜索追踪
func (jc *JaegerClient) SearchTraces(ctx context.Context, service string, start, end time.Time, limit int) ([]Trace, error) {
	url := fmt.Sprintf("%s/api/traces?service=%s&start=%d&end=%d&limit=%d",
		jc.baseURL,
		service,
		start.UnixMicro(),
		end.UnixMicro(),
		limit,
	)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := jc.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to search traces: %s", string(body))
	}

	var result struct {
		Data []Trace `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

// PerformanceAnalyzer 性能分析器
type PerformanceAnalyzer struct {
	client *JaegerClient
}

// NewPerformanceAnalyzer 创建性能分析器
func NewPerformanceAnalyzer(jaegerURL string) *PerformanceAnalyzer {
	return &PerformanceAnalyzer{
		client: NewJaegerClient(jaegerURL),
	}
}

// TraceAnalysis 追踪分析结果
type TraceAnalysis struct {
	TraceID           string                 `json:"trace_id"`
	TotalDuration     time.Duration          `json:"total_duration"`
	SpanCount         int                    `json:"span_count"`
	ServiceCount      int                    `json:"service_count"`
	CriticalPath      []SpanSummary          `json:"critical_path"`
	Bottlenecks       []Bottleneck           `json:"bottlenecks"`
	ServiceBreakdown  map[string]DurationStat `json:"service_breakdown"`
	OperationBreakdown map[string]DurationStat `json:"operation_breakdown"`
	Errors            []ErrorInfo            `json:"errors"`
}

// SpanSummary Span摘要
type SpanSummary struct {
	SpanID        string        `json:"span_id"`
	OperationName string        `json:"operation_name"`
	ServiceName   string        `json:"service_name"`
	Duration      time.Duration `json:"duration"`
	StartTime     time.Time     `json:"start_time"`
}

// Bottleneck 瓶颈
type Bottleneck struct {
	SpanID         string        `json:"span_id"`
	OperationName  string        `json:"operation_name"`
	ServiceName    string        `json:"service_name"`
	Duration       time.Duration `json:"duration"`
	Percentage     float64       `json:"percentage"`
	Recommendation string        `json:"recommendation"`
}

// DurationStat 持续时间统计
type DurationStat struct {
	Count      int           `json:"count"`
	TotalTime  time.Duration `json:"total_time"`
	AvgTime    time.Duration `json:"avg_time"`
	MinTime    time.Duration `json:"min_time"`
	MaxTime    time.Duration `json:"max_time"`
	Percentage float64       `json:"percentage"`
}

// ErrorInfo 错误信息
type ErrorInfo struct {
	SpanID        string    `json:"span_id"`
	OperationName string    `json:"operation_name"`
	ServiceName   string    `json:"service_name"`
	ErrorMessage  string    `json:"error_message"`
	Timestamp     time.Time `json:"timestamp"`
}

// AnalyzeTrace 分析追踪
func (pa *PerformanceAnalyzer) AnalyzeTrace(ctx context.Context, traceID string) (*TraceAnalysis, error) {
	// 获取追踪数据
	trace, err := pa.client.GetTrace(ctx, traceID)
	if err != nil {
		return nil, err
	}

	analysis := &TraceAnalysis{
		TraceID:            traceID,
		SpanCount:          len(trace.Spans),
		ServiceBreakdown:   make(map[string]DurationStat),
		OperationBreakdown: make(map[string]DurationStat),
	}

	// 计算总持续时间
	if len(trace.Spans) > 0 {
		rootSpan := pa.findRootSpan(trace.Spans)
		analysis.TotalDuration = time.Duration(rootSpan.Duration) * time.Microsecond
	}

	// 统计服务数量
	services := make(map[string]bool)
	for _, span := range trace.Spans {
		if process, ok := trace.Processes[span.ProcessID]; ok {
			services[process.ServiceName] = true
		}
	}
	analysis.ServiceCount = len(services)

	// 分析关键路径
	analysis.CriticalPath = pa.findCriticalPath(trace)

	// 识别瓶颈
	analysis.Bottlenecks = pa.findBottlenecks(trace, analysis.TotalDuration)

	// 按服务分解
	analysis.ServiceBreakdown = pa.breakdownByService(trace)

	// 按操作分解
	analysis.OperationBreakdown = pa.breakdownByOperation(trace)

	// 提取错误
	analysis.Errors = pa.extractErrors(trace)

	return analysis, nil
}

// findRootSpan 查找根span
func (pa *PerformanceAnalyzer) findRootSpan(spans []Span) Span {
	for _, span := range spans {
		isRoot := true
		for _, ref := range span.References {
			if ref.RefType == "CHILD_OF" {
				isRoot = false
				break
			}
		}
		if isRoot {
			return span
		}
	}
	return spans[0]
}

// findCriticalPath 查找关键路径
func (pa *PerformanceAnalyzer) findCriticalPath(trace *Trace) []SpanSummary {
	// 构建span图
	spanMap := make(map[string]Span)
	for _, span := range trace.Spans {
		spanMap[span.SpanID] = span
	}

	// 查找根span
	rootSpan := pa.findRootSpan(trace.Spans)

	// DFS找到最长路径
	var criticalPath []SpanSummary
	pa.dfs(rootSpan, spanMap, trace.Processes, []SpanSummary{}, &criticalPath)

	return criticalPath
}

// dfs 深度优先搜索
func (pa *PerformanceAnalyzer) dfs(span Span, spanMap map[string]Span, processes map[string]Process, currentPath []SpanSummary, longestPath *[]SpanSummary) {
	// 添加当前span到路径
	summary := SpanSummary{
		SpanID:        span.SpanID,
		OperationName: span.OperationName,
		Duration:      time.Duration(span.Duration) * time.Microsecond,
		StartTime:     time.Unix(0, span.StartTime*1000),
	}
	if process, ok := processes[span.ProcessID]; ok {
		summary.ServiceName = process.ServiceName
	}
	currentPath = append(currentPath, summary)

	// 查找子span
	hasChildren := false
	for _, s := range spanMap {
		for _, ref := range s.References {
			if ref.RefType == "CHILD_OF" && ref.SpanID == span.SpanID {
				hasChildren = true
				pa.dfs(s, spanMap, processes, currentPath, longestPath)
			}
		}
	}

	// 如果是叶子节点，比较路径长度
	if !hasChildren {
		var currentDuration time.Duration
		for _, s := range currentPath {
			currentDuration += s.Duration
		}
		var longestDuration time.Duration
		for _, s := range *longestPath {
			longestDuration += s.Duration
		}
		if currentDuration > longestDuration {
			*longestPath = make([]SpanSummary, len(currentPath))
			copy(*longestPath, currentPath)
		}
	}
}

// findBottlenecks 查找瓶颈
func (pa *PerformanceAnalyzer) findBottlenecks(trace *Trace, totalDuration time.Duration) []Bottleneck {
	var bottlenecks []Bottleneck

	for _, span := range trace.Spans {
		duration := time.Duration(span.Duration) * time.Microsecond
		percentage := float64(duration) / float64(totalDuration) * 100

		// 持续时间超过总时间20%的span视为瓶颈
		if percentage > 20 {
			serviceName := ""
			if process, ok := trace.Processes[span.ProcessID]; ok {
				serviceName = process.ServiceName
			}

			bottleneck := Bottleneck{
				SpanID:        span.SpanID,
				OperationName: span.OperationName,
				ServiceName:   serviceName,
				Duration:      duration,
				Percentage:    percentage,
			}

			// 根据操作类型提供建议
			bottleneck.Recommendation = pa.getRecommendation(span.OperationName, duration)

			bottlenecks = append(bottlenecks, bottleneck)
		}
	}

	// 按持续时间排序
	sort.Slice(bottlenecks, func(i, j int) bool {
		return bottlenecks[i].Duration > bottlenecks[j].Duration
	})

	return bottlenecks
}

// getRecommendation 获取优化建议
func (pa *PerformanceAnalyzer) getRecommendation(operation string, duration time.Duration) string {
	if duration > 1*time.Second {
		if containsAny(operation, []string{"db", "database", "query", "sql"}) {
			return "Consider adding database indexes, optimizing queries, or implementing caching"
		}
		if containsAny(operation, []string{"http", "api", "request"}) {
			return "Consider implementing request batching, caching, or connection pooling"
		}
		if containsAny(operation, []string{"cache", "redis"}) {
			return "Check cache hit rate and consider cache warming or TTL optimization"
		}
		if containsAny(operation, []string{"llm", "agent", "model"}) {
			return "Consider prompt optimization, streaming responses, or model downgrade"
		}
	}
	return "Monitor this operation and consider optimization if it becomes a recurring bottleneck"
}

// breakdownByService 按服务分解
func (pa *PerformanceAnalyzer) breakdownByService(trace *Trace) map[string]DurationStat {
	breakdown := make(map[string]DurationStat)

	for _, span := range trace.Spans {
		serviceName := ""
		if process, ok := trace.Processes[span.ProcessID]; ok {
			serviceName = process.ServiceName
		}

		duration := time.Duration(span.Duration) * time.Microsecond

		stat := breakdown[serviceName]
		stat.Count++
		stat.TotalTime += duration

		if stat.MinTime == 0 || duration < stat.MinTime {
			stat.MinTime = duration
		}
		if duration > stat.MaxTime {
			stat.MaxTime = duration
		}

		breakdown[serviceName] = stat
	}

	// 计算平均值和百分比
	var totalDuration time.Duration
	for _, stat := range breakdown {
		totalDuration += stat.TotalTime
	}

	for service, stat := range breakdown {
		if stat.Count > 0 {
			stat.AvgTime = stat.TotalTime / time.Duration(stat.Count)
		}
		if totalDuration > 0 {
			stat.Percentage = float64(stat.TotalTime) / float64(totalDuration) * 100
		}
		breakdown[service] = stat
	}

	return breakdown
}

// breakdownByOperation 按操作分解
func (pa *PerformanceAnalyzer) breakdownByOperation(trace *Trace) map[string]DurationStat {
	breakdown := make(map[string]DurationStat)

	for _, span := range trace.Spans {
		duration := time.Duration(span.Duration) * time.Microsecond

		stat := breakdown[span.OperationName]
		stat.Count++
		stat.TotalTime += duration

		if stat.MinTime == 0 || duration < stat.MinTime {
			stat.MinTime = duration
		}
		if duration > stat.MaxTime {
			stat.MaxTime = duration
		}

		breakdown[span.OperationName] = stat
	}

	// 计算平均值和百分比
	var totalDuration time.Duration
	for _, stat := range breakdown {
		totalDuration += stat.TotalTime
	}

	for operation, stat := range breakdown {
		if stat.Count > 0 {
			stat.AvgTime = stat.TotalTime / time.Duration(stat.Count)
		}
		if totalDuration > 0 {
			stat.Percentage = float64(stat.TotalTime) / float64(totalDuration) * 100
		}
		breakdown[operation] = stat
	}

	return breakdown
}

// extractErrors 提取错误
func (pa *PerformanceAnalyzer) extractErrors(trace *Trace) []ErrorInfo {
	var errors []ErrorInfo

	for _, span := range trace.Spans {
		// 检查error标签
		for _, tag := range span.Tags {
			if tag.Key == "error" && tag.Value == true {
				serviceName := ""
				if process, ok := trace.Processes[span.ProcessID]; ok {
					serviceName = process.ServiceName
				}

				errorMsg := ""
				// 查找错误消息
				for _, t := range span.Tags {
					if t.Key == "error.message" || t.Key == "message" {
						if msg, ok := t.Value.(string); ok {
							errorMsg = msg
						}
					}
				}

				errors = append(errors, ErrorInfo{
					SpanID:        span.SpanID,
					OperationName: span.OperationName,
					ServiceName:   serviceName,
					ErrorMessage:  errorMsg,
					Timestamp:     time.Unix(0, span.StartTime*1000),
				})
			}
		}
	}

	return errors
}

// containsAny 检查字符串是否包含任意子串
func containsAny(s string, substrs []string) bool {
	for _, substr := range substrs {
		if len(s) >= len(substr) {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}

// CompareTraces 比较两个追踪
func (pa *PerformanceAnalyzer) CompareTraces(ctx context.Context, traceID1, traceID2 string) (*TraceComparison, error) {
	analysis1, err := pa.AnalyzeTrace(ctx, traceID1)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze trace 1: %w", err)
	}

	analysis2, err := pa.AnalyzeTrace(ctx, traceID2)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze trace 2: %w", err)
	}

	comparison := &TraceComparison{
		TraceID1:          traceID1,
		TraceID2:          traceID2,
		DurationDiff:      analysis2.TotalDuration - analysis1.TotalDuration,
		DurationDiffPct:   float64(analysis2.TotalDuration-analysis1.TotalDuration) / float64(analysis1.TotalDuration) * 100,
		SpanCountDiff:     analysis2.SpanCount - analysis1.SpanCount,
		ServiceCountDiff:  analysis2.ServiceCount - analysis1.ServiceCount,
	}

	return comparison, nil
}

// TraceComparison 追踪比较结果
type TraceComparison struct {
	TraceID1         string        `json:"trace_id_1"`
	TraceID2         string        `json:"trace_id_2"`
	DurationDiff     time.Duration `json:"duration_diff"`
	DurationDiffPct  float64       `json:"duration_diff_pct"`
	SpanCountDiff    int           `json:"span_count_diff"`
	ServiceCountDiff int           `json:"service_count_diff"`
}
