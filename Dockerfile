# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY interverse-gateway/go.mod interverse-gateway/go.sum ./
COPY interverse-contracts /interverse-contracts
RUN go mod edit -replace=github.com/LimeOnTop/interverse-contracts=/interverse-contracts
RUN go mod download

COPY interverse-gateway/main.go ./
COPY interverse-gateway/internal/ ./internal/

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api-gateway .

FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
RUN adduser -D -s /bin/sh appuser
USER appuser
WORKDIR /app
COPY --from=builder /app/api-gateway .
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1
EXPOSE 8080
CMD ["./api-gateway"]
