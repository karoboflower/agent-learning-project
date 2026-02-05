package aggregator

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/olivere/elastic/v7"
)

// LogAggregator 日志聚合器
type LogAggregator struct {
	client *elastic.Client
}

// NewLogAggregator 创建日志聚合器
func NewLogAggregator(elasticsearchURL string) (*LogAggregator, error) {
	client, err := elastic.NewClient(
		elastic.SetURL(elasticsearchURL),
		elastic.SetSniff(false),
		elastic.SetHealthcheck(true),
		elastic.SetHealthcheckInterval(10*time.Second),
	)
	if err != nil {
		return nil, err
	}

	return &LogAggregator{
		client: client,
	}, nil
}

// LogQuery 日志查询参数
type LogQuery struct {
	StartTime   time.Time
	EndTime     time.Time
	Level       string
	Service     string
	TenantID    string
	UserID      string
	RequestID   string
	Message     string
	Limit       int
	Offset      int
	SortField   string
	SortOrder   string
}

// LogEntry 日志条目
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`
	Logger    string                 `json:"logger"`
	Message   string                 `json:"message"`
	Service   string                 `json:"service"`
	Version   string                 `json:"version"`
	Env       string                 `json:"env"`
	TenantID  string                 `json:"tenant_id,omitempty"`
	UserID    string                 `json:"user_id,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
	TraceID   string                 `json:"trace_id,omitempty"`
	SpanID    string                 `json:"span_id,omitempty"`
	Method    string                 `json:"method,omitempty"`
	Path      string                 `json:"path,omitempty"`
	Status    int                    `json:"status,omitempty"`
	Duration  float64                `json:"duration,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Extra     map[string]interface{} `json:"extra,omitempty"`
}

// SearchLogs 搜索日志
func (la *LogAggregator) SearchLogs(ctx context.Context, query *LogQuery) ([]*LogEntry, int64, error) {
	// 构建查询
	boolQuery := elastic.NewBoolQuery()

	// 时间范围
	if !query.StartTime.IsZero() || !query.EndTime.IsZero() {
		rangeQuery := elastic.NewRangeQuery("@timestamp")
		if !query.StartTime.IsZero() {
			rangeQuery.Gte(query.StartTime)
		}
		if !query.EndTime.IsZero() {
			rangeQuery.Lte(query.EndTime)
		}
		boolQuery.Filter(rangeQuery)
	}

	// 日志级别
	if query.Level != "" {
		boolQuery.Filter(elastic.NewTermQuery("level", query.Level))
	}

	// 服务
	if query.Service != "" {
		boolQuery.Filter(elastic.NewTermQuery("service", query.Service))
	}

	// 租户
	if query.TenantID != "" {
		boolQuery.Filter(elastic.NewTermQuery("tenant_id", query.TenantID))
	}

	// 用户
	if query.UserID != "" {
		boolQuery.Filter(elastic.NewTermQuery("user_id", query.UserID))
	}

	// 请求ID
	if query.RequestID != "" {
		boolQuery.Filter(elastic.NewTermQuery("request_id", query.RequestID))
	}

	// 消息搜索
	if query.Message != "" {
		boolQuery.Must(elastic.NewMatchQuery("message", query.Message))
	}

	// 设置默认值
	if query.Limit == 0 {
		query.Limit = 100
	}
	if query.SortField == "" {
		query.SortField = "@timestamp"
	}
	if query.SortOrder == "" {
		query.SortOrder = "desc"
	}

	// 执行搜索
	searchResult, err := la.client.Search().
		Index("app-logs-*").
		Query(boolQuery).
		Sort(query.SortField, query.SortOrder == "desc").
		From(query.Offset).
		Size(query.Limit).
		Pretty(true).
		Do(ctx)

	if err != nil {
		return nil, 0, err
	}

	// 解析结果
	logs := make([]*LogEntry, 0, len(searchResult.Hits.Hits))
	for _, hit := range searchResult.Hits.Hits {
		var log LogEntry
		if err := json.Unmarshal(hit.Source, &log); err != nil {
			continue
		}
		logs = append(logs, &log)
	}

	return logs, searchResult.TotalHits(), nil
}

// AggregateByLevel 按日志级别聚合
func (la *LogAggregator) AggregateByLevel(ctx context.Context, startTime, endTime time.Time) (map[string]int64, error) {
	// 时间范围查询
	rangeQuery := elastic.NewRangeQuery("@timestamp").
		Gte(startTime).
		Lte(endTime)

	// 聚合
	agg := elastic.NewTermsAggregation().Field("level").Size(10)

	// 执行聚合
	searchResult, err := la.client.Search().
		Index("app-logs-*").
		Query(rangeQuery).
		Aggregation("by_level", agg).
		Size(0).
		Do(ctx)

	if err != nil {
		return nil, err
	}

	// 解析结果
	result := make(map[string]int64)
	if agg, found := searchResult.Aggregations.Terms("by_level"); found {
		for _, bucket := range agg.Buckets {
			if key, ok := bucket.Key.(string); ok {
				result[key] = bucket.DocCount
			}
		}
	}

	return result, nil
}

// AggregateByService 按服务聚合
func (la *LogAggregator) AggregateByService(ctx context.Context, startTime, endTime time.Time) (map[string]int64, error) {
	rangeQuery := elastic.NewRangeQuery("@timestamp").
		Gte(startTime).
		Lte(endTime)

	agg := elastic.NewTermsAggregation().Field("service").Size(20)

	searchResult, err := la.client.Search().
		Index("app-logs-*").
		Query(rangeQuery).
		Aggregation("by_service", agg).
		Size(0).
		Do(ctx)

	if err != nil {
		return nil, err
	}

	result := make(map[string]int64)
	if agg, found := searchResult.Aggregations.Terms("by_service"); found {
		for _, bucket := range agg.Buckets {
			if key, ok := bucket.Key.(string); ok {
				result[key] = bucket.DocCount
			}
		}
	}

	return result, nil
}

// AggregateTimeline 按时间聚合
func (la *LogAggregator) AggregateTimeline(ctx context.Context, startTime, endTime time.Time, interval string) ([]TimelinePoint, error) {
	rangeQuery := elastic.NewRangeQuery("@timestamp").
		Gte(startTime).
		Lte(endTime)

	// 日期直方图聚合
	agg := elastic.NewDateHistogramAggregation().
		Field("@timestamp").
		FixedInterval(interval).
		MinDocCount(0).
		ExtendedBounds(&elastic.ExtendedBounds{
			Min: startTime.UnixNano() / int64(time.Millisecond),
			Max: endTime.UnixNano() / int64(time.Millisecond),
		})

	searchResult, err := la.client.Search().
		Index("app-logs-*").
		Query(rangeQuery).
		Aggregation("timeline", agg).
		Size(0).
		Do(ctx)

	if err != nil {
		return nil, err
	}

	// 解析结果
	points := make([]TimelinePoint, 0)
	if agg, found := searchResult.Aggregations.DateHistogram("timeline"); found {
		for _, bucket := range agg.Buckets {
			points = append(points, TimelinePoint{
				Timestamp: time.Unix(0, bucket.Key*int64(time.Millisecond)),
				Count:     bucket.DocCount,
			})
		}
	}

	return points, nil
}

// TimelinePoint 时间线点
type TimelinePoint struct {
	Timestamp time.Time `json:"timestamp"`
	Count     int64     `json:"count"`
}

// GetErrorStats 获取错误统计
func (la *LogAggregator) GetErrorStats(ctx context.Context, startTime, endTime time.Time) (*ErrorStats, error) {
	// 查询错误日志
	rangeQuery := elastic.NewRangeQuery("@timestamp").
		Gte(startTime).
		Lte(endTime)

	levelQuery := elastic.NewTermsQuery("level", "error", "fatal")

	boolQuery := elastic.NewBoolQuery().
		Filter(rangeQuery).
		Filter(levelQuery)

	// 聚合
	serviceAgg := elastic.NewTermsAggregation().Field("service").Size(10)
	errorTypeAgg := elastic.NewTermsAggregation().Field("error_class").Size(10)

	searchResult, err := la.client.Search().
		Index("app-logs-*").
		Query(boolQuery).
		Aggregation("by_service", serviceAgg).
		Aggregation("by_error_type", errorTypeAgg).
		Size(0).
		Do(ctx)

	if err != nil {
		return nil, err
	}

	stats := &ErrorStats{
		TotalErrors:   searchResult.TotalHits(),
		ByService:     make(map[string]int64),
		ByErrorType:   make(map[string]int64),
	}

	// 解析服务聚合
	if agg, found := searchResult.Aggregations.Terms("by_service"); found {
		for _, bucket := range agg.Buckets {
			if key, ok := bucket.Key.(string); ok {
				stats.ByService[key] = bucket.DocCount
			}
		}
	}

	// 解析错误类型聚合
	if agg, found := searchResult.Aggregations.Terms("by_error_type"); found {
		for _, bucket := range agg.Buckets {
			if key, ok := bucket.Key.(string); ok {
				stats.ByErrorType[key] = bucket.DocCount
			}
		}
	}

	return stats, nil
}

// ErrorStats 错误统计
type ErrorStats struct {
	TotalErrors int64            `json:"total_errors"`
	ByService   map[string]int64 `json:"by_service"`
	ByErrorType map[string]int64 `json:"by_error_type"`
}

// DeleteOldLogs 删除旧日志
func (la *LogAggregator) DeleteOldLogs(ctx context.Context, olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)

	rangeQuery := elastic.NewRangeQuery("@timestamp").
		Lt(cutoff)

	_, err := la.client.DeleteByQuery().
		Index("app-logs-*").
		Query(rangeQuery).
		Do(ctx)

	return err
}

// GetLogsByRequestID 根据请求ID获取日志
func (la *LogAggregator) GetLogsByRequestID(ctx context.Context, requestID string) ([]*LogEntry, error) {
	query := elastic.NewTermQuery("request_id", requestID)

	searchResult, err := la.client.Search().
		Index("app-logs-*").
		Query(query).
		Sort("@timestamp", true).
		Size(1000).
		Do(ctx)

	if err != nil {
		return nil, err
	}

	logs := make([]*LogEntry, 0, len(searchResult.Hits.Hits))
	for _, hit := range searchResult.Hits.Hits {
		var log LogEntry
		if err := json.Unmarshal(hit.Source, &log); err != nil {
			continue
		}
		logs = append(logs, &log)
	}

	return logs, nil
}

// GetLogsByTraceID 根据追踪ID获取日志
func (la *LogAggregator) GetLogsByTraceID(ctx context.Context, traceID string) ([]*LogEntry, error) {
	query := elastic.NewTermQuery("trace_id", traceID)

	searchResult, err := la.client.Search().
		Index("app-logs-*").
		Query(query).
		Sort("@timestamp", true).
		Size(1000).
		Do(ctx)

	if err != nil {
		return nil, err
	}

	logs := make([]*LogEntry, 0, len(searchResult.Hits.Hits))
	for _, hit := range searchResult.Hits.Hits {
		var log LogEntry
		if err := json.Unmarshal(hit.Source, &log); err != nil {
			continue
		}
		logs = append(logs, &log)
	}

	return logs, nil
}

// Close 关闭连接
func (la *LogAggregator) Close() error {
	la.client.Stop()
	return nil
}

// HealthCheck 健康检查
func (la *LogAggregator) HealthCheck(ctx context.Context) error {
	_, _, err := la.client.Ping(la.client.GetURL()[0]).Do(ctx)
	if err != nil {
		return fmt.Errorf("elasticsearch health check failed: %w", err)
	}
	return nil
}
