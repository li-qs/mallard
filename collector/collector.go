package collector

import (
	"encoding/json"
	"net/http"
	"sort"
	"sync"
)

type Span struct {
	TraceID   string `json:"trace_id"`
	SpanID    string `json:"span_id"`
	ParentID  string `json:"parent_id"`
	Service   string `json:"service"`
	Operation string `json:"operation"`
	StartTime int64  `json:"start_time"`
	Duration  int64  `json:"duration"` // 单位：纳秒
	Status    int    `json:"status"`   // HTTP 状态吗
	Error     string `json:"error,omitempty"`
}

type Collector struct {
	mu    sync.RWMutex
	spans map[string][]Span // key: TraceID
}

func NewCollector() *Collector {
	return &Collector{
		spans: make(map[string][]Span),
	}
}

func (c *Collector) HandleCollect(w http.ResponseWriter, r *http.Request) {
	var span Span
	if err := json.NewDecoder(r.Body).Decode(&span); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c.mu.Lock()
	c.spans[span.TraceID] = append(c.spans[span.TraceID], span)
	c.mu.Unlock()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (c *Collector) HandleQuery(w http.ResponseWriter, r *http.Request) {
	traceID := r.URL.Path[len("/trace/"):]
	if traceID == "" {
		http.Error(w, "missing trace_id", http.StatusBadRequest)
		return
	}

	c.mu.Lock()
	spans, ok := c.spans[traceID]
	c.mu.Unlock()

	if !ok {
		http.Error(w, "trace not found", http.StatusNotFound)
		return
	}

	sort.Slice(spans, func(i, j int) bool {
		return spans[i].StartTime < spans[j].StartTime
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(spans)
}
