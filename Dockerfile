# === build stage ===
FROM golang:1.24 AS builder
WORKDIR /src

# 先複製 go.mod / go.sum (利於快取)
COPY go.mod go.sum ./
RUN go mod download

# 複製其餘程式碼並編譯 ./cmd/api
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /bin/goapp ./cmd/api

# === run stage ===
FROM alpine:3.20

# 非 root 執行
RUN adduser -D -H appuser
USER appuser

COPY --from=builder /bin/goapp /bin/goapp

EXPOSE 8080

# 執行
ENTRYPOINT ["/bin/goapp"]
