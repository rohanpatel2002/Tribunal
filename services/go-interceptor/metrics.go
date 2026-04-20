package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// MetricsCollector tracks application metrics
type MetricsCollector struct {
	// Request counters
	TotalRequests      int64
	RequestsByEndpoint map[string]int64

	// Response times (in milliseconds)
	AvgResponseTime map[string]float64

	// Business metrics
	TotalAnalyses          int64
	TotalCriticalRisks     int64
	TotalHighRisks         int64
	TotalAIGeneratedScores map[string]float64
	AverageAIScore         float64

	// Error tracking
	TotalErrors      int64
	ErrorsByEndpoint map[string]int64
}

// NewMetricsCollector initializes metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		RequestsByEndpoint:     make(map[string]int64),
		AvgResponseTime:        make(map[string]float64),
		TotalAIGeneratedScores: make(map[string]float64),
		ErrorsByEndpoint:       make(map[string]int64),
	}
}

// RecordRequest logs an API request
func (mc *MetricsCollector) RecordRequest(endpoint string, startTime time.Time, statusCode int) {
	duration := time.Since(startTime).Milliseconds()
	mc.TotalRequests++
	mc.RequestsByEndpoint[endpoint]++

	// Update average response time
	currentAvg := mc.AvgResponseTime[endpoint]
	count := float64(mc.RequestsByEndpoint[endpoint])
	mc.AvgResponseTime[endpoint] = (currentAvg*(count-1) + float64(duration)) / count

	if statusCode >= 400 {
		mc.TotalErrors++
		mc.ErrorsByEndpoint[endpoint]++
	}

	slog.Debug("request recorded",
		"endpoint", endpoint,
		"status", statusCode,
		"duration_ms", duration,
		"total_requests", mc.TotalRequests,
	)
}

// RecordAnalysis logs a completed analysis
func (mc *MetricsCollector) RecordAnalysis(aiScore float64, critical, high int, repo string) {
	mc.TotalAnalyses++
	mc.TotalCriticalRisks += int64(critical)
	mc.TotalHighRisks += int64(high)

	// Track AI scores per repository
	currentScoreSum := mc.TotalAIGeneratedScores[repo]
	count := float64(mc.TotalAnalyses)
	mc.AverageAIScore = (currentScoreSum + aiScore) / count
	mc.TotalAIGeneratedScores[repo] = currentScoreSum + aiScore
}

// GetPrometheusMetrics returns metrics in Prometheus format
func (mc *MetricsCollector) GetPrometheusMetrics() string {
	return fmt.Sprintf(`# HELP tribunal_total_requests Total HTTP requests received
# TYPE tribunal_total_requests counter
tribunal_total_requests %d

# HELP tribunal_total_errors Total HTTP errors (4xx, 5xx)
# TYPE tribunal_total_errors counter
tribunal_total_errors %d

# HELP tribunal_total_analyses Total code analyses performed
# TYPE tribunal_total_analyses counter
tribunal_total_analyses %d

# HELP tribunal_total_critical_risks Total critical-severity findings
# TYPE tribunal_total_critical_risks counter
tribunal_total_critical_risks %d

# HELP tribunal_total_high_risks Total high-severity findings
# TYPE tribunal_total_high_risks counter
tribunal_total_high_risks %d

# HELP tribunal_average_ai_score Average AI confidence score (0-1)
# TYPE tribunal_average_ai_score gauge
tribunal_average_ai_score %.4f

# HELP tribunal_avg_response_time_ms Average response time by endpoint
# TYPE tribunal_avg_response_time_ms gauge
%s
`,
		mc.TotalRequests,
		mc.TotalErrors,
		mc.TotalAnalyses,
		mc.TotalCriticalRisks,
		mc.TotalHighRisks,
		mc.AverageAIScore,
		mc.formatEndpointMetrics(),
	)
}

// formatEndpointMetrics formats per-endpoint metrics
func (mc *MetricsCollector) formatEndpointMetrics() string {
	metrics := ""
	for endpoint, avgTime := range mc.AvgResponseTime {
		metrics += fmt.Sprintf(`tribunal_avg_response_time_ms{endpoint="%s"} %.2f
`, endpoint, avgTime)
	}
	return metrics
}

// MetricsMiddleware wraps HTTP handler with metrics collection
func MetricsMiddleware(metrics *MetricsCollector) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()

			// Wrap response writer to capture status code
			wrappedWriter := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next(wrappedWriter, r)

			// Record metrics
			metrics.RecordRequest(r.URL.Path, startTime, wrappedWriter.statusCode)
		}
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// GetMetricsHandler returns Prometheus metrics endpoint
func GetMetricsHandler(metrics *MetricsCollector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		fmt.Fprint(w, metrics.GetPrometheusMetrics())
	}
}

// GrafanaAlerts defines alert rules for Grafana
var GrafanaAlerts = `
{
  "alerts": [
    {
      "name": "HighCriticalRiskRate",
      "description": "Alert when critical risks exceed 10 in 1 hour",
      "threshold": 10,
      "condition": "tribunal_total_critical_risks > 10"
    },
    {
      "name": "HighResponseTime",
      "description": "Alert when avg response time exceeds 500ms",
      "threshold": 500,
      "condition": "tribunal_avg_response_time_ms > 500"
    },
    {
      "name": "ErrorRateHigh",
      "description": "Alert when error rate exceeds 5%",
      "threshold": 0.05,
      "condition": "tribunal_total_errors / tribunal_total_requests > 0.05"
    },
    {
      "name": "HighAIDetectionRate",
      "description": "Alert when average AI score exceeds 0.75",
      "threshold": 0.75,
      "condition": "tribunal_average_ai_score > 0.75"
    }
  ]
}
`
