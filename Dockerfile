# Build stage: compile a static Linux binary for amd64.
FROM golang:1.23 AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY main.go auth.go metrics.go handlers.go system_info.go services.go \
     process.go extended_metrics.go health.go cron.go security.go maintenance.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /osctl \
    main.go auth.go metrics.go handlers.go system_info.go services.go \
    process.go extended_metrics.go health.go cron.go security.go maintenance.go

# Runtime stage: minimal image with just the binary.
FROM alpine:3.24

COPY --from=builder /osctl /usr/local/bin/osctl

# HTTP API server listens here by default (OSCTL_PORT).
EXPOSE 12000

# NOTE: container mode is limited — osctl wraps host commands (systemctl,
# docker, journalctl, ...), so most endpoints only work when the container
# is run with appropriate mounts/privileges, e.g.:
#   docker run -v /run/systemd:/run/systemd -p 12000:12000 osctl api
ENTRYPOINT ["/usr/local/bin/osctl"]
CMD ["api"]
