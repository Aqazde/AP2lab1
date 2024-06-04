package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

const (
	CONN_HOST = "localhost"
	CONN_port = "3335"
	CONN_type = "tcp"
)

var wg sync.WaitGroup

func main() {
	defer wg.Wait()

	config := &tls.Config{
		InsecureSkipVerify: true,
	}

	conn, err := tls.Dial(CONN_type, CONN_HOST+":"+CONN_port, config)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		os.Exit(1)
	}

	wg.Add(1)
	go readMessages(conn)

	wg.Add(1)
	go writeMessages(conn)
}

func readMessages(conn net.Conn) {
	defer wg.Done()

	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Disconnected from server.")
			break
		}
		fmt.Print(message)
	}
}
func writeMessages(conn net.Conn) {
	defer wg.Done()

	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(conn)
	for {
		fmt.Print("Enter message: ")
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			break
		}

		message = strings.TrimSpace(message)

		if _, err := writer.WriteString(message + "\n"); err != nil {
			fmt.Println("Error sending message:", err)
			break
		}

		if err := writer.Flush(); err != nil {
			fmt.Println("Error flushing writer:", err)
			break
		}

		if message == "/leave" {
			fmt.Println("Leaving chat...")
		}
	}
	conn.Close()
}
