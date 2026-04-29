// Command hello-api runs a minimal HTTP service that exposes two
// endpoints used as a connectivity smoke-test:
//
//	GET /        -> {"message":"Hello, World!"}
//	GET /health  -> {"status":"ok"}
//
// The service depends only on the Go standard library, listens on port
// 9090 by default, and starts with no required environment variables. The
// PORT environment variable may override the listen port if set.
package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/asdlc-repos/lll342481/hello-api/internal/handlers"
	"github.com/asdlc-repos/lll342481/hello-api/internal/middleware"
)

const (
	// defaultPort is the TCP port the server listens on when the PORT
	// environment variable is unset or invalid. The component spec
	// pins this to 9090.
	defaultPort = 9090

	// readHeaderTimeout bounds how long the server waits for request
	// headers; it mitigates Slowloris-style resource exhaustion.
	readHeaderTimeout = 5 * time.Second

	// shutdownTimeout caps how long graceful shutdown waits for
	// in-flight requests before forcing the server closed.
	shutdownTimeout = 10 * time.Second
)

// resolvePort returns the port to listen on. It honours the PORT
// environment variable when present and parseable as a 1..65535 integer,
// otherwise it falls back to defaultPort. Invalid values are logged but
// never cause startup to fail — the service must boot with no required
// environment.
func resolvePort() int {
	raw, ok := os.LookupEnv("PORT")
	if !ok || raw == "" {
		return defaultPort
	}
	p, err := strconv.Atoi(raw)
	if err != nil || p < 1 || p > 65535 {
		log.Printf("warn: invalid PORT %q, falling back to %d", raw, defaultPort)
		return defaultPort
	}
	return p
}

// newRouter builds the HTTP handler tree: a plain http.ServeMux wrapped
// in the request-logging middleware. Defining it as a separate function
// keeps main() small and makes the routing easy to exercise from tests.
func newRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handlers.Health)
	// "/" is registered last because http.ServeMux uses longest-prefix
	// matching; /health must be its own entry to avoid being shadowed.
	mux.HandleFunc("/", handlers.Hello)
	return middleware.Logging(mux)
}

func main() {
	port := resolvePort()
	addr := net.JoinHostPort("0.0.0.0", strconv.Itoa(port))

	srv := &http.Server{
		Addr:              addr,
		Handler:           newRouter(),
		ReadHeaderTimeout: readHeaderTimeout,
	}

	// Run ListenAndServe in a goroutine so main can also wait for an
	// OS signal and trigger graceful shutdown on Ctrl-C / SIGTERM.
	serverErr := make(chan error, 1)
	go func() {
		log.Printf("hello-api listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		if err != nil {
			log.Fatalf("error: server failed: %v", err)
		}
	case sig := <-stop:
		log.Printf("received signal %s, shutting down", sig)
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("error: graceful shutdown failed: %v", err)
		}
	}

	log.Printf("hello-api stopped")
}
