# ---------- Build Stage ----------
FROM golang:1.22-alpine AS builder

WORKDIR /app

# 複製 source code
COPY server/ ./server/

# build binary
RUN go build -o server-app ./server/main.go

# ---------- Runtime Stage ----------
FROM alpine:latest

WORKDIR /app

# 從 builder 複製 binary
COPY --from=builder /app/server-app .

# 如果你的 server 用 8080
EXPOSE 8080

# run binary
CMD ["./server-app"]

