package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

var connections []net.Conn
var mu sync.Mutex
var total int64

func createConn(target string) {
	conn, err := net.Dial("tcp", target)
	if err != nil {
		fmt.Println("connect error:", err)
		return
	}

	fmt.Fprintf(conn, "GET /health HTTP/1.1\r\n")
	fmt.Fprintf(conn, "Host: localhost\r\n")

	mu.Lock()
	connections = append(connections, conn)
	mu.Unlock()

	atomic.AddInt64(&total, 1)

	// 持續送 header
	go func(c net.Conn) {
		for {
			fmt.Fprintf(c, "X-Test: slow\r\n")
			time.Sleep(10 * time.Second)
		}
	}(conn)
}

func createBatch(target string, n int) {
	for i := 0; i < n; i++ {
		go createConn(target)
	}
}

func sendHeaderAll() {
	mu.Lock()
	defer mu.Unlock()

	for _, conn := range connections {
		fmt.Fprintf(conn, "X-Test: keep-alive\r\n")
	}

	fmt.Println("header sent to all connections")
}

func printStatus() {
	fmt.Printf("Current connections: %d\n", atomic.LoadInt64(&total))
}

func main() {

	target := "127.0.0.1:8083"

	fmt.Println("Starting with 50 connections...")
	createBatch(target, 50)

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("cmd> ")

		if !scanner.Scan() {
			break
		}

		cmd := scanner.Text()

		switch cmd {

		case "add":
			createBatch(target, 50)
			fmt.Println("Added 50 connections")

		case "cmd":
			sendHeaderAll()

		case "stat":
			printStatus()

		case "quit":
			fmt.Println("exit")
			return

		default:
			fmt.Println("commands: add | cmd | stat | quit")
		}
	}
}

