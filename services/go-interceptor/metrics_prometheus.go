package main

import (
	"fmt"
	"net/http"
	"strings"
)

// PrometheusMetricsHandler exposes metrics in Prometheus text format.
func PrometheusMetricsHandler(metrics RedisMetricsProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}

		w.Header().Set("Content-Type", "text/plain; version=0.0.4")

		serviceUp := 1.0
		redisConfigured := 0.0
		var redisStats RedisStoreMetrics
		if metrics != nil {
			redisConfigured = 1.0
			redisStats = metrics.RedisMetrics()
		}

		lines := []string{
			"# HELP tribunal_service_up Whether the go-interceptor is up (1) or down (0).",
			"# TYPE tribunal_service_up gauge",
			fmt.Sprintf("tribunal_service_up %.0f", serviceUp),
			"# HELP tribunal_redis_configured Whether Redis session store is configured.",
			"# TYPE tribunal_redis_configured gauge",
			fmt.Sprintf("tribunal_redis_configured %.0f", redisConfigured),
			"# HELP tribunal_redis_operations_total Redis session store operations.",
			"# TYPE tribunal_redis_operations_total counter",
			fmt.Sprintf("tribunal_redis_operations_total %d", redisStats.Operations),
			"# HELP tribunal_redis_errors_total Redis session store errors.",
			"# TYPE tribunal_redis_errors_total counter",
			fmt.Sprintf("tribunal_redis_errors_total %d", redisStats.Errors),
			"# HELP tribunal_redis_last_ping_ok Last Redis ping result (1 ok, 0 error).",
			"# TYPE tribunal_redis_last_ping_ok gauge",
			fmt.Sprintf("tribunal_redis_last_ping_ok %.0f", boolToGauge(redisStats.LastPingOK)),
			"# HELP tribunal_redis_last_ping_latency_ms Last Redis ping latency in milliseconds.",
			"# TYPE tribunal_redis_last_ping_latency_ms gauge",
			fmt.Sprintf("tribunal_redis_last_ping_latency_ms %d", redisStats.LastPingLatencyMs),
			"# HELP tribunal_redis_session_ttl_seconds Redis session TTL seconds.",
			"# TYPE tribunal_redis_session_ttl_seconds gauge",
			fmt.Sprintf("tribunal_redis_session_ttl_seconds %d", redisStats.SessionTTLSeconds),
			"# HELP tribunal_redis_oauth_state_ttl_seconds Redis OAuth state TTL seconds.",
			"# TYPE tribunal_redis_oauth_state_ttl_seconds gauge",
			fmt.Sprintf("tribunal_redis_oauth_state_ttl_seconds %d", redisStats.OAuthStateTTLSeconds),
		}

		_, _ = fmt.Fprintln(w, strings.Join(lines, "\n"))
	}
}

func boolToGauge(value bool) float64 {
	if value {
		return 1
	}
	return 0
}
