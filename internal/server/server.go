// Package server holds the small HTTP helpers shared by every control-plane
// service, so health checks, JSON encoding, logging, and tenant extraction are
// written once instead of duplicated per service.
package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

var (
	startTime    = time.Now()
	requestCount atomic.Int64
	errorCount   atomic.Int64
)

// JSON writes v as a JSON response with the given status.
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

// Error writes a JSON {"error": msg} body.
func Error(w http.ResponseWriter, status int, msg string) {
	JSON(w, status, map[string]string{"error": msg})
}

// Decode reads a JSON request body into v, returning false (and writing a 400)
// on failure.
func Decode(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		Error(w, http.StatusBadRequest, "invalid JSON body: "+err.Error())
		return false
	}
	return true
}

// TenantID extracts the tenant from the standard header ("" if absent).
func TenantID(r *http.Request) string { return r.Header.Get("X-Tenant-ID") }

// New builds an *http.Server with health endpoints and request logging already
// wired, so each service only registers its own routes.
func New(addr string, mux *http.ServeMux) *http.Server {
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		JSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})
	// Publishable health metrics in Prometheus text format (scrapeable; ships to
	// Prometheus/Grafana like any other CNCF workload).
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		fmt.Fprintf(w, "# HELP platform_up 1 if the service is up\n# TYPE platform_up gauge\nplatform_up 1\n")
		fmt.Fprintf(w, "# HELP platform_uptime_seconds Process uptime\n# TYPE platform_uptime_seconds gauge\nplatform_uptime_seconds %.0f\n", time.Since(startTime).Seconds())
		fmt.Fprintf(w, "# HELP platform_http_requests_total Total HTTP requests\n# TYPE platform_http_requests_total counter\nplatform_http_requests_total %d\n", requestCount.Load())
		fmt.Fprintf(w, "# HELP platform_http_errors_total Total HTTP 5xx responses\n# TYPE platform_http_errors_total counter\nplatform_http_errors_total %d\n", errorCount.Load())
	})
	return &http.Server{
		Addr:              addr,
		Handler:           logging(mux),
		ReadHeaderTimeout: 10 * time.Second,
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestCount.Add(1)
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		if rec.status >= 500 {
			errorCount.Add(1)
		}
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, rec.status, time.Since(start))
	})
}
