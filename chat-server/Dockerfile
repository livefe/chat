FROM --platform=linux/amd64 golang:1.24-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git ca-certificates tzdata
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server \
 && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/migration ./cmd/migration \
 && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/worker ./cmd/worker
FROM --platform=linux/amd64 alpine:3.20
RUN apk add --no-cache ca-certificates tzdata wget \
 && adduser -D -H -u 10001 appuser
WORKDIR /app
COPY --from=builder /out/server /out/migration /out/worker /app/
COPY configs/config.yaml configs/scheduler.yaml /app/configs/
RUN mkdir -p /app/data /app/logs \
 && chown -R appuser:appuser /app
USER appuser
ENV TZ=Asia/Shanghai
EXPOSE 8080
CMD ["/app/server"]
