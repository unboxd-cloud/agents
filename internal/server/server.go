// Package server holds the small HTTP helpers shared by every control-plane
// service, so health checks, JSON encoding, logging, and tenant extraction are
// written once instead of duplicated per service.
package server

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
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
	return &http.Server{
		Addr:              addr,
		Handler:           logging(mux),
		ReadHeaderTimeout: 10 * time.Second,
	}
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}
