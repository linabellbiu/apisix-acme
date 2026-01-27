FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o apisix-acme-service .

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/apisix-acme-service .
RUN apk add --no-cache ca-certificates

# Expose data volume
VOLUME ["/root/data"]

CMD ["./apisix-acme-service", "-c", "/root/data/config.yaml"]
