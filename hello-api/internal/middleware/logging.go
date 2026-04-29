// Package middleware contains HTTP middleware used by the hello-api
// service.
//
// The middleware here is intentionally tiny — just request logging — so
// the service can satisfy its observability requirement without taking
// on any third-party dependencies.
package middleware

import (
	"log"
	"net/http"
	"time"
)

// statusRecorder wraps http.ResponseWriter so middleware can observe the
// status code chosen by the inner handler. Without it, WriteHeader is
// invisible to the caller.
type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

// WriteHeader records the status before delegating to the wrapped writer.
// Subsequent calls are forwarded as-is so net/http's own "superfluous
// WriteHeader" warning still fires for buggy handlers.
func (s *statusRecorder) WriteHeader(code int) {
	if !s.wroteHeader {
		s.status = code
		s.wroteHeader = true
	}
	s.ResponseWriter.WriteHeader(code)
}

// Write ensures that handlers which call Write without an explicit
// WriteHeader still record the implicit 200 status that net/http applies.
func (s *statusRecorder) Write(b []byte) (int, error) {
	if !s.wroteHeader {
		s.status = http.StatusOK
		s.wroteHeader = true
	}
	return s.ResponseWriter.Write(b)
}

// Logging returns middleware that logs the method, path, status, and
// duration of every request to stdout via the standard log package. The
// format is intentionally simple and human-readable so it works well in
// container logs without requiring a structured-log shipper.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rec, r)

		log.Printf("%s %s %d %s", r.Method, r.URL.Path, rec.status, time.Since(start))
	})
}
