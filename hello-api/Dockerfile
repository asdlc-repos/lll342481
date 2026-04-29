# syntax=docker/dockerfile:1
#
# Multi-stage build for the hello-api Go service.
#
# Stage 1 (build) compiles a fully static Linux binary with the toolchain
# bundled in the official golang image. Stage 2 (runtime) copies just the
# binary into a minimal alpine image so the published artifact is small
# and contains no compiler or source code.

# ---------- build stage ----------
FROM golang:1.21 AS build
WORKDIR /src

# Copy module files first so dependency resolution can be cached
# independently of source changes. The service has no third-party
# dependencies today, but this layout keeps the cache friendly if any
# are added later.
COPY go.mod ./
RUN go mod download

# Now copy the rest of the source and compile.
COPY . .

# CGO_ENABLED=0 produces a statically linked binary so it runs on the
# slim alpine base without needing libc. -trimpath and -ldflags="-s -w"
# strip absolute paths and debug symbols for a smaller, more portable
# image.
ENV CGO_ENABLED=0 \
    GOOS=linux
RUN go build -trimpath -ldflags="-s -w" -o /out/hello-api .

# ---------- runtime stage ----------
FROM alpine:latest
WORKDIR /app

# ca-certificates is harmless to include and lets the service make HTTPS
# calls if it ever needs to. tini gives us PID-1 signal forwarding so
# graceful shutdown works correctly under `docker stop`.
RUN apk add --no-cache ca-certificates tini

# Run as a non-root user; the binary doesn't need any elevated
# privileges to bind 9090.
RUN addgroup -S app && adduser -S -G app app
USER app

COPY --from=build /out/hello-api /app/hello-api

EXPOSE 9090

ENTRYPOINT ["/sbin/tini", "--", "/app/hello-api"]
