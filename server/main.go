package main

import (
    "context"
    "errors"
    "log"
    "net/http"
    _ "net/http/pprof"
    "os"
    "os/signal"
    "strconv"
    "syscall"
    "time"
)

func main() {
    // ===== Config =====
    port := getEnv("PORT", "8083")
    shutdownTimeoutSec := getEnvAsInt("SHUTDOWN_TIMEOUT", 10)

    // ===== Router =====
    mux := http.NewServeMux()
    mux.HandleFunc("/health", healthHandler)
    mux.HandleFunc("/version", versionHandler)

    // ===== HTTP Server with Timeouts =====
    server := &http.Server{
        Addr:              ":" + port,
        Handler:           loggingMiddleware(mux),
    }

    // ===== Server Start (non-blocking) =====
    go func() {
        log.Printf("server starting on port %s\n", port)
        if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
            log.Fatalf("listen error: %v", err)
        }
    }()

    // ===== Debug Server Start (non-blocking) =====
    go http.ListenAndServe("127.0.0.1:6060", nil)

    // ===== Signal Handling =====
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

    <-stop
    log.Println("shutdown signal received")

    // ===== Graceful Shutdown =====
    ctx, cancel := context.WithTimeout(context.Background(), time.Duration(shutdownTimeoutSec)*time.Second)
    defer cancel()

    if err := server.Shutdown(ctx); err != nil {
        log.Printf("graceful shutdown failed: %v\n", err)
    } else {
        log.Println("server exited properly")
    }
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("ok"))
}

func versionHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("v1.0.0"))
}

func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)
        log.Printf("%s %s %s\n", r.Method, r.URL.Path, time.Since(start))
    })
}

func getEnv(key, fallback string) string {
    if val, ok := os.LookupEnv(key); ok {
        return val
    }
    return fallback
}

func getEnvAsInt(key string, fallback int) int {
    if valStr, ok := os.LookupEnv(key); ok {
        if val, err := strconv.Atoi(valStr); err == nil {
            return val
        }
    }
    return fallback
}

