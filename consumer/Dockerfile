# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.23-alpine3.20 AS builder

WORKDIR /build

# Install build dependencies
RUN apk update && apk upgrade && apk add --no-cache ca-certificates git tzdata

# Copy go mod files
COPY go.mod go.sum* ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o consumer .

# Final stage
FROM gcr.io/distroless/static-debian12:nonroot

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy binary
COPY --from=builder /build/consumer /consumer

# Expose port for health check
EXPOSE 8081

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD ["/consumer", "-health"]

# Run
ENTRYPOINT ["/consumer"]
