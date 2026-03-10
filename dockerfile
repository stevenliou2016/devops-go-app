# ---------- Build Stage ----------
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY server/ ./server/

RUN go build -o server-app ./server/main.go

# ---------- Runtime Stage ----------
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/server-app .

EXPOSE 8080

# run binary
CMD ["./server-app"]

