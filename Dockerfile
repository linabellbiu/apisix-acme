FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/golang:1.25.3-alpine3.21 AS builder

WORKDIR /app

ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn,direct
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o apisix-acme-service .

FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/library/alpine:3.21
WORKDIR /root/
COPY --from=builder /app/apisix-acme-service .
RUN sed -i 's#https://dl-cdn.alpinelinux.org#https://mirrors.aliyun.com#g' /etc/apk/repositories && \
    apk add --no-cache ca-certificates

# Expose data volume
VOLUME ["/root/data"]

CMD ["./apisix-acme-service", "-c", "/root/data/config.yaml"]
