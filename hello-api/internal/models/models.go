// Package models contains JSON response shapes for the hello-api service.
//
// The schemas mirror the component's OpenAPI specification exactly: every
// field is required and serialized with a stable JSON tag. Keeping these
// types in one place makes the contract easy to audit and reuse from
// handlers and tests.
package models

// HelloResponse is the body returned by GET /.
//
// Per the OpenAPI spec, the `message` field is required and is the only
// property of the response object.
type HelloResponse struct {
	Message string `json:"message"`
}

// HealthResponse is the body returned by GET /health.
//
// Per the OpenAPI spec, the `status` field is required and is the only
// property of the response object. A value of "ok" indicates the service
// is healthy.
type HealthResponse struct {
	Status string `json:"status"`
}

// ErrorResponse is a generic error envelope used when a handler cannot
// produce its normal response (for example, a JSON marshaling failure).
//
// It is intentionally simple — a single `error` string — so it can be
// emitted without risking a second serialization failure.
type ErrorResponse struct {
	Error string `json:"error"`
}
