package main

import (
	"net/http"
	"time"
)

// MetricsHandler exposes lightweight JSON metrics for monitoring.
func MetricsHandler(metrics RedisMetricsProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}

		response := map[string]interface{}{
			"status":    "ok",
			"timestamp": time.Now().UTC(),
		}

		if metrics == nil {
			response["redis"] = map[string]interface{}{"status": "not_configured"}
		} else {
			response["redis"] = metrics.RedisMetrics()
		}

		writeJSON(w, http.StatusOK, response)
	}
}
