# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy gen directories for all services first (needed for go mod download)
COPY user-service/gen/ ./user-service/gen/
COPY interview-service/gen/ ./interview-service/gen/
COPY candidate-service/gen/ ./candidate-service/gen/
COPY report-service/gen/ ./report-service/gen/
COPY technology-service/gen/ ./technology-service/gen/
COPY question-service/gen/ ./question-service/gen/

# Copy go mod files
COPY api-gateway/go.mod api-gateway/go.sum ./
RUN go mod download

# Copy source code (excluding service directories that will be copied separately)
COPY api-gateway/main.go ./
COPY api-gateway/internal/ ./internal/

# Update dependencies and build
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api-gateway .

# Final stage
FROM alpine:3.19

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN adduser -D -s /bin/sh appuser
USER appuser

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/api-gateway .

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

EXPOSE 8080

CMD ["./api-gateway"]
