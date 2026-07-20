package tracer

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type key string

const TraceIDHeader = "X-Trace-ID"
const ParentIDHeader = "X-Parent-ID"
const TraceIDKey key = "trace_id"

var httpClient = &http.Client{
	Timeout:   3 * time.Second,
	Transport: &http.Transport{MaxIdleConnsPerHost: 100},
}

func TracingMiddleware(serviceName, collectorURL string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			traceID := r.Header.Get(TraceIDHeader)
			if traceID == "" {
				traceID = GenerateTraceID()
			}
			w.Header().Set(TraceIDHeader, traceID)

			ctx := context.WithValue(r.Context(), TraceIDKey, traceID)
			r = r.WithContext(ctx)

			start := time.Now()

			wrapped := &responseWriter{ResponseWriter: w, status: 200}
			next.ServeHTTP(wrapped, r)

			span := Span{
				SpanID:    generateSpanID(),
				TraceID:   traceID,
				ParentID:  r.Header.Get(ParentIDHeader),
				Service:   serviceName,
				Operation: r.Method + " " + r.URL.Path,
				StartTime: start.UnixNano(),
				Duration:  time.Since(start).Nanoseconds(),
				Status:    wrapped.status,
			}

			go reportSpan(span, collectorURL)
		})
	}
}

func reportSpan(span Span, collectorURL string) {
	data, err := json.Marshal(span)
	if err != nil {
		log.Printf("marshal span failed: %v", err)
		return
	}

	resp, err := httpClient.Post(collectorURL, "application/json", bytes.NewReader(data))
	if err != nil {
		log.Printf("report span failed: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("collector returned non-200: %d", resp.StatusCode)
	}
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (w *responseWriter) WithHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}
