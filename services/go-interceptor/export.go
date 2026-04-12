package main

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ExportFormat defines the output format for exports
type ExportFormat string

const (
	FormatJSON ExportFormat = "json"
	FormatCSV  ExportFormat = "csv"
)

// ExportRequest defines export parameters
type ExportRequest struct {
	Repository  string       `json:"repository"`
	Format      ExportFormat `json:"format"`
	StartDate   *time.Time   `json:"startDate,omitempty"`
	EndDate     *time.Time   `json:"endDate,omitempty"`
	MinSeverity string       `json:"minSeverity,omitempty"`
}

// ExportAnalysisDataToJSON exports audit records as JSON
func ExportAnalysisDataToJSON(records []PRAnalysisRecord, w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(map[string]interface{}{
		"exportedAt": time.Now().UTC(),
		"count":      len(records),
		"data":       records,
	})
}

// ExportAnalysisDataToCSV exports audit records as CSV
func ExportAnalysisDataToCSV(records []PRAnalysisRecord, w io.Writer) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	headers := []string{
		"PR Number",
		"Repository",
		"Recommendation",
		"Total Files",
		"AI Generated",
		"Critical",
		"High",
		"Medium",
		"Low",
	}
	if err := writer.Write(headers); err != nil {
		return err
	}

	// Write records
	for _, record := range records {
		row := []string{
			strconv.Itoa(record.PRNumber),
			record.Repository,
			record.Recommendation,
			strconv.Itoa(record.TotalFiles),
			strconv.Itoa(record.AIGenerated),
			strconv.Itoa(record.Critical),
			strconv.Itoa(record.High),
			strconv.Itoa(record.Medium),
			strconv.Itoa(record.Low),
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

// ExportHandler provides export functionality for audit data
func ExportHandler(repo Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}

		repository := r.URL.Query().Get("repository")
		if repository == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "repository parameter required"})
			return
		}

		format := r.URL.Query().Get("format")
		if format == "" {
			format = "json"
		}

		if repo == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "database not configured"})
			return
		}

		// Fetch audit logs
		limit := 5000 // High limit for exports
		records, err := repo.GetRecentAnalyses(r.Context(), limit, repository)
		if err != nil {
			slog.Error("failed to fetch analyses for export", "error", err, "repo", repository)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch data"})
			return
		}

		if records == nil {
			records = []PRAnalysisRecord{}
		}

		// Set response headers
		timestamp := time.Now().Format("20060102-150405")
		filename := ""
		contentType := ""

		switch strings.ToLower(format) {
		case "csv":
			contentType = "text/csv"
			filename = "tribunal-audit-" + timestamp + ".csv"
			w.Header().Set("Content-Type", contentType)
			w.Header().Set("Content-Disposition", "attachment; filename="+filename)

			if err := ExportAnalysisDataToCSV(records, w); err != nil {
				slog.Error("CSV export failed", "error", err)
			}

		case "json":
			fallthrough
		default:
			contentType = "application/json"
			filename = "tribunal-audit-" + timestamp + ".json"
			w.Header().Set("Content-Type", contentType)
			w.Header().Set("Content-Disposition", "attachment; filename="+filename)

			if err := ExportAnalysisDataToJSON(records, w); err != nil {
				slog.Error("JSON export failed", "error", err)
			}
		}

		slog.Info("audit data exported",
			"repo", repository,
			"format", format,
			"recordCount", len(records),
			"filename", filename,
		)
	}
}
