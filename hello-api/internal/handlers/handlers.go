// Package handlers implements the HTTP request handlers for the hello-api
// service.
//
// Each handler:
//   - sets Content-Type: application/json on every response,
//   - serializes its response body using encoding/json,
//   - returns HTTP 200 on success, and
//   - falls back to HTTP 500 with a JSON error envelope if marshaling fails.
//
// Handlers depend only on the Go standard library, in keeping with the
// component's "zero external dependencies" constraint.
package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/asdlc-repos/lll342481/hello-api/internal/models"
)

const (
	// contentTypeJSON is the Content-Type used for every response body.
	contentTypeJSON = "application/json"

	// helloMessage is the canonical greeting returned by GET /.
	helloMessage = "Hello, World!"

	// statusOK is the canonical health value returned by GET /health.
	statusOK = "ok"
)

// writeJSON serializes v to w with the given status code and the
// application/json content type. If marshaling fails, it logs the error
// and writes a 500 response with a minimal JSON error envelope. Headers
// are written before the body, so callers must not write to w before
// calling this helper.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	body, err := json.Marshal(v)
	if err != nil {
		// Marshaling our own response types should never fail in
		// practice — the structs only contain strings — but if it
		// somehow does we still need to return *something* sensible
		// to the client and surface the failure in logs.
		log.Printf("error: json marshal failed: %v", err)
		w.Header().Set("Content-Type", contentTypeJSON)
		w.WriteHeader(http.StatusInternalServerError)
		// Hand-written JSON so a second marshaling failure can't
		// cascade. Safe because the literal contains no user input.
		_, _ = w.Write([]byte(`{"error":"internal server error"}`))
		return
	}

	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(status)
	if _, err := w.Write(body); err != nil {
		// The client likely disconnected; log for observability but
		// there is nothing else we can do at this point.
		log.Printf("error: failed to write response body: %v", err)
	}
}

// Hello handles GET / and returns {"message":"Hello, World!"}.
//
// Non-GET methods receive HTTP 405. The default mux already restricts the
// path to "/", but it does *not* filter by method, so we do that here.
func Hello(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeJSON(w, http.StatusMethodNotAllowed, models.ErrorResponse{
			Error: "method not allowed",
		})
		return
	}

	// http.ServeMux routes "/" as a catch-all; explicitly reject any
	// path that isn't exactly "/" so unknown URLs produce a 404 rather
	// than a misleading "Hello, World!".
	if r.URL.Path != "/" {
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{
			Error: "not found",
		})
		return
	}

	writeJSON(w, http.StatusOK, models.HelloResponse{Message: helloMessage})
}

// Health handles GET /health and returns {"status":"ok"}.
//
// Non-GET methods receive HTTP 405. The handler performs no external I/O,
// so "ok" simply indicates the process is up and serving traffic — which
// is exactly what a container-native liveness probe needs.
func Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeJSON(w, http.StatusMethodNotAllowed, models.ErrorResponse{
			Error: "method not allowed",
		})
		return
	}

	writeJSON(w, http.StatusOK, models.HealthResponse{Status: statusOK})
}
