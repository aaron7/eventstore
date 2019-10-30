package store

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// These are the store API paths
const (
	APIPathEvents = "/events"
	APIPathQuery  = "/query"
)

var requestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "eventstore",
	Name:      "api_request_duration_seconds",
	Help:      "Time (in seconds) spent serving HTTP requests.",
	Buckets:   prometheus.DefBuckets,
}, []string{"method", "path", "status_code"})

func init() {
	prometheus.MustRegister(requestDuration)
}

// API serves the store API
type API struct {
	Store *Store
}

func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	iw := &interceptingWriter{http.StatusOK, w}
	w = iw
	defer func(begin time.Time) {
		requestDuration.WithLabelValues(
			r.Method,
			r.URL.Path,
			strconv.Itoa(iw.code),
		).Observe(time.Since(begin).Seconds())
	}(time.Now())

	method, path := r.Method, r.URL.Path
	switch {
	case method == "POST" && path == APIPathEvents:
		a.handlePostEvents(w, r)
	case method == "POST" && path == APIPathQuery:
		a.handleQuery(w, r)
	default:
		http.NotFound(w, r)
	}
}

// https://github.com/oklog/oklog/blob/master/pkg/store/api.go#L119
type interceptingWriter struct {
	code int
	http.ResponseWriter
}

func (iw *interceptingWriter) WriteHeader(code int) {
	iw.code = code
	iw.ResponseWriter.WriteHeader(code)
}

func (iw *interceptingWriter) Flush() {
	if f, ok := iw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Events is a list of events
type Events struct {
	Events []Event `json:"events"`
}

// Event is one sampled event
type Event struct {
	TS         int               `json:"ts"`
	Samplerate int               `json:"samplerate"`
	Data       map[string]string `json:"data"`
}

func (a *API) handlePostEvents(w http.ResponseWriter, r *http.Request) {
	var payload Events
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	a.Store.IngestEvents(payload.Events)
}

// Query is a query to the store
type Query struct {
	Data []Data `json:"data"`
}

// Data is ...
type Data struct {
	Name       string      `json:"name"`
	Keys       []string    `json:"keys"`
	Filters    []Filter    `json:"filters"`
	Operations []Operation `json:"operations"`
	HideData   bool        `json:"hideData"`
}

// Filter is a filter on a dimension
type Filter struct {
	Type  string `json:"type"`  // eq | regex
	Key   string `json:"key"`   // e.g. event_tag
	Value string `json:"value"` // e.g. product_view
}

// Operation operates on data
type Operation struct {
	Type string `json:"type"` // count | unique | group_by
	Key  string `json:"key"`  // e.g. user_id
}

func (a *API) handleQuery(w http.ResponseWriter, r *http.Request) {
	var query Query
	err := json.NewDecoder(r.Body).Decode(&query)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	eventIDs := a.Store.QueryEvents(query)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(eventIDs)
}
