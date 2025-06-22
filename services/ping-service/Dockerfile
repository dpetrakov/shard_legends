FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN go build -o ping-service .

FROM alpine:latest

RUN apk --no-cache add ca-certificates wget
WORKDIR /root/

COPY --from=builder /app/ping-service .

EXPOSE 8080

# Health check using the dedicated /health endpoint
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD wget -q -O - http://localhost:8080/health || exit 1

CMD ["./ping-service"]