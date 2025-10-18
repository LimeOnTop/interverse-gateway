# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy common proto files first
COPY services/proto/ ./services/proto/

# Copy gen directories for all services
COPY services/user-service/gen/ ./services/user-service/gen/
COPY services/interview-service/gen/ ./services/interview-service/gen/
COPY services/candidate-service/gen/ ./services/candidate-service/gen/
COPY services/report-service/gen/ ./services/report-service/gen/
COPY services/technology-service/gen/ ./services/technology-service/gen/
COPY services/question-service/gen/ ./services/question-service/gen/

# Copy go mod files
COPY services/api-gateway/go.mod services/api-gateway/go.sum ./
RUN go mod download

# Copy source code
COPY services/api-gateway/ ./

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
