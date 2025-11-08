# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build terminal apps first
RUN mkdir -p terminal-apps-exe && \
    for app in terminal-apps/*; do \
      if [ -d "$app" ] && [ -f "$app/main.go" ]; then \
        echo "Building $(basename $app)..." && \
        cd "$app" && \
        CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -trimpath -o ../terminal-apps-exe/$(basename $app) . && \
        cd ..; \
      fi; \
    done

# Build main server
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -trimpath -o /app/server .

# Runtime stage - minimal image
FROM alpine:latest

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/server .
COPY --from=builder /app/terminal-apps-exe ./terminal-apps-exe

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    chown -R appuser:appuser /app

USER appuser

EXPOSE 8080

CMD ["./server"]
