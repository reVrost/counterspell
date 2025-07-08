package counterspell

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/revrost/counterspell/internal/db"
)

// APIHandler handles HTTP requests for the counterspell API
type APIHandler struct {
	queries *db.Queries
	db      *sql.DB
}

// NewAPIHandler creates a new API handler
func NewAPIHandler(database *sql.DB) *APIHandler {
	return &APIHandler{
		queries: db.New(database),
		db:      database,
	}
}

// APIResponse represents the standard API response format
type APIResponse struct {
	Metadata map[string]any `json:"metadata"`
	Data     any            `json:"data"`
}

// TraceListItem represents a trace in the list view
type TraceListItem struct {
	TraceID        string  `json:"trace_id"`
	RootSpanName   string  `json:"root_span_name"`
	TraceStartTime int64   `json:"trace_start_time"`
	DurationMs     float64 `json:"duration_ms"`
	SpanCount      int64   `json:"span_count"`
	ErrorCount     int64   `json:"error_count"`
	HasError       bool    `json:"has_error"`
}

// TraceDetail represents detailed trace information
type TraceDetail struct {
	TraceID string     `json:"trace_id"`
	Spans   []SpanItem `json:"spans"`
}

// SpanItem represents a span in the trace detail view
type SpanItem struct {
	SpanID       string         `json:"span_id"`
	TraceID      string         `json:"trace_id"`
	ParentSpanID string         `json:"parent_span_id"`
	Name         string         `json:"name"`
	StartTime    int64          `json:"start_time"`
	EndTime      int64          `json:"end_time"`
	DurationNs   int64          `json:"duration_ns"`
	Attributes   map[string]any `json:"attributes"`
	ServiceName  string         `json:"service_name"`
	HasError     bool           `json:"has_error"`
}

// LogItem represents a log entry
type LogItem struct {
	ID         int64          `json:"id"`
	Timestamp  int64          `json:"timestamp"`
	Level      string         `json:"level"`
	Message    string         `json:"message"`
	TraceID    string         `json:"trace_id"`
	SpanID     string         `json:"span_id"`
	Attributes map[string]any `json:"attributes"`
}

// QueryLogs handles GET /counterspell/api/logs
func (h *APIHandler) QueryLogs(c echo.Context) error {
	// Parse query parameters
	limit := parseIntParam(c.QueryParam("limit"), 100)
	offset := parseIntParam(c.QueryParam("offset"), 0)
	levelStr := c.QueryParam("level")
	q := c.QueryParam("q") // Full-text search
	startTimeStr := c.QueryParam("start_time")
	endTimeStr := c.QueryParam("end_time")
	traceIDStr := c.QueryParam("trace_id")

	ctx := c.Request().Context()

	// If we have filters, use the filtered query, otherwise use the simple one
	var logs []db.Log
	var total int64
	var err error

	if levelStr != "" || startTimeStr != "" || endTimeStr != "" || traceIDStr != "" {
		// Set nullable parameters
		var level sql.NullString
		if levelStr != "" {
			level = sql.NullString{String: levelStr, Valid: true}
		}
		var traceID sql.NullString
		if traceIDStr != "" {
			traceID = sql.NullString{String: traceIDStr, Valid: true}
		}
		var startTime sql.NullInt64
		if startTimeStr != "" {
			parsedStartTime, err := strconv.ParseInt(startTimeStr, 10, 64)
			if err == nil {
				startTime = sql.NullInt64{Int64: parsedStartTime, Valid: true}
			}
		}
		var endTime sql.NullInt64
		if endTimeStr != "" {
			parsedEndTime, err := strconv.ParseInt(endTimeStr, 10, 64)
			if err == nil {
				endTime = sql.NullInt64{Int64: parsedEndTime, Valid: true}
			}
		}

		logs, err = h.queries.GetLogsWithFilters(ctx, level, traceID, startTime, endTime, int32(limit), int32(offset))
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to query logs")
		}

		total, err = h.queries.CountLogsWithFilters(ctx, level, traceID, startTime, endTime)
		if err != nil {
			total = 0
		}
	} else {
		// Use simple query
		logs, err = h.queries.GetLogs(ctx, int32(limit), int32(offset))
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to query logs")
		}

		total, err = h.queries.CountLogs(ctx)
		if err != nil {
			total = 0
		}
	}

	// Convert to API format and apply text search if needed
	apiLogs := []LogItem{}
	for _, log := range logs {
		apiLog := LogItem{
			ID:        log.ID,
			Timestamp: log.Timestamp,
			Level:     log.Level,
			Message:   log.Message,
			TraceID:   log.TraceID,
			SpanID:    log.SpanID,
		}

		// Parse attributes JSON
		if len(log.Attributes) > 0 {
			var attrs map[string]any
			if json.Unmarshal(log.Attributes, &attrs) == nil {
				apiLog.Attributes = attrs
			} else {
				apiLog.Attributes = make(map[string]any)
			}
		} else {
			apiLog.Attributes = make(map[string]any)
		}

		// Apply text search filter if provided
		if q != "" {
			searchTerm := strings.ToLower(q)
			if !strings.Contains(strings.ToLower(apiLog.Message), searchTerm) &&
				!strings.Contains(strings.ToLower(string(log.Attributes)), searchTerm) {
				continue
			}
		}

		apiLogs = append(apiLogs, apiLog)
	}

	response := APIResponse{
		Metadata: map[string]any{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
		Data: apiLogs,
	}

	return c.JSON(http.StatusOK, response)
}

// QueryTraces handles GET /counterspell/api/traces
func (h *APIHandler) QueryTraces(c echo.Context) error {
	// Parse query parameters
	limit := parseIntParam(c.QueryParam("limit"), 100)
	offset := parseIntParam(c.QueryParam("offset"), 0)
	q := c.QueryParam("q")                     // search root span name
	hasErrorParam := c.QueryParam("has_error") // filter by error status

	ctx := c.Request().Context()

	// Get root spans
	rootSpans, err := h.queries.GetRootSpans(ctx, int32(limit), int32(offset))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to query traces")
	}

	// Get trace statistics
	traceStats, err := h.queries.GetTraceStats(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to query trace stats")
	}

	// Create a map for quick lookup of trace stats
	statsMap := make(map[string]db.GetTraceStatsRow)
	for _, stat := range traceStats {
		statsMap[stat.TraceID] = stat
	}

	// Build trace list items
	traces := []TraceListItem{}
	for _, rootSpan := range rootSpans {
		// Apply name filter if provided
		if q != "" && !strings.Contains(strings.ToLower(rootSpan.Name), strings.ToLower(q)) {
			continue
		}

		// Get stats for this trace
		stats, exists := statsMap[rootSpan.TraceID]
		if !exists {
			stats = db.GetTraceStatsRow{
				TraceID:    rootSpan.TraceID,
				SpanCount:  1,
				ErrorCount: 0,
			}
		}

		errorCount := stats.ErrorCount

		// Apply error filter if provided
		if hasErrorParam == "true" && errorCount == 0 {
			continue
		}

		// Calculate duration in milliseconds
		// Convert nanoseconds to time.Time for duration calculation
		startTime := time.Unix(0, rootSpan.StartTime)
		endTime := time.Unix(0, rootSpan.EndTime)
		durationMs := float64(endTime.Sub(startTime).Nanoseconds()) / 1000000

		trace := TraceListItem{
			TraceID:        rootSpan.TraceID,
			RootSpanName:   rootSpan.Name,
			TraceStartTime: rootSpan.StartTime,
			DurationMs:     durationMs,
			SpanCount:      stats.SpanCount,
			ErrorCount:     errorCount,
			HasError:       errorCount > 0,
		}

		traces = append(traces, trace)
	}

	// Get total count
	total, err := h.queries.CountTraces(ctx)
	if err != nil {
		total = 0
	}

	response := APIResponse{
		Metadata: map[string]any{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
		Data: traces,
	}

	return c.JSON(http.StatusOK, response)
}

// GetTraceDetails handles GET /counterspell/api/traces/:trace_id
func (h *APIHandler) GetTraceDetails(c echo.Context) error {
	traceID := c.Param("trace_id")
	if traceID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "trace_id is required")
	}

	ctx := c.Request().Context()

	// Get all spans for the trace using the generated query
	spans, err := h.queries.GetTraceDetails(ctx, traceID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to query trace details")
	}

	if len(spans) == 0 {
		return echo.NewHTTPError(http.StatusNotFound, "Trace not found")
	}

	// Convert to API format
	apiSpans := make([]SpanItem, 0, len(spans))
	for _, span := range spans {
		apiSpan := SpanItem{
			SpanID:       span.SpanID,
			TraceID:      span.TraceID,
			ParentSpanID: span.ParentSpanID,
			Name:         span.Name,
			StartTime:    span.StartTime,
			EndTime:      span.EndTime,
			DurationNs:   span.DurationNs,
			ServiceName:  span.ServiceName,
			HasError:     span.HasError,
		}

		// Parse attributes JSON
		if len(span.Attributes) > 0 {
			var attrs map[string]any
			if json.Unmarshal(span.Attributes, &attrs) == nil {
				apiSpan.Attributes = attrs
			} else {
				apiSpan.Attributes = make(map[string]any)
			}
		} else {
			apiSpan.Attributes = make(map[string]any)
		}

		apiSpans = append(apiSpans, apiSpan)
	}

	response := TraceDetail{
		TraceID: traceID,
		Spans:   apiSpans,
	}

	return c.JSON(http.StatusOK, response)
}

// parseIntParam parses an integer parameter with a default value
func parseIntParam(param string, defaultValue int) int {
	if param == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(param)
	if err != nil {
		return defaultValue
	}

	return value
}
