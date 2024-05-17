package main

import (
	"bufio"
	"net"
	"os"
	"strings"
	"testing"
	"time"
)

func setup() {
	go main()
	time.Sleep(1 * time.Second)
}

func teardown() {
	// Teardown logic
}

// Helper function to create a client connection
func createClient(t *testing.T) net.Conn {
	conn, err := net.Dial(CONN_TYPE, CONN_HOST+CONN_PORT)
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	return conn
}

// 1 Test: Server startup and client connection
func TestServerStartup(t *testing.T) {
	setup()
	defer teardown()

	conn := createClient(t)
	defer conn.Close()

	_, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		t.Fatalf("Failed to read welcome message: %v", err)
	}
}

// 2 Test: Client can join lobby
func TestJoinLobby(t *testing.T) {
	defer teardown()

	conn := createClient(t)
	defer conn.Close()

	_, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		t.Fatalf("Failed to join lobby: %v", err)
	}
}

// 3 Test: Client can set username
func TestSetUsername(t *testing.T) {
	defer teardown()

	conn := createClient(t)
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	_, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("Failed to read welcome message: %v", err)
	}

	_, err = writer.WriteString("/setUsername testuser\n")
	if err != nil {
		t.Fatalf("Failed to send setUsername command: %v", err)
	}
	writer.Flush()

	message, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}
	if !strings.Contains(message, "Username set to: testuser") {
		t.Fatalf("Unexpected response: %v", message)
	}
}

// 4 Test: Client can create a chat room
func TestCreateChatRoom(t *testing.T) {
	defer teardown()

	conn := createClient(t)
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	_, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("Failed to read welcome message: %v", err)
	}

	_, err = writer.WriteString("/create testroom\n")
	if err != nil {
		t.Fatalf("Failed to send create command: %v", err)
	}
	writer.Flush()

	message, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}
	if !strings.Contains(message, "Created chat room \"testroom\"") {
		t.Fatalf("Unexpected response: %v", message)
	}
}

// 5 Test: Client can join a chat room
func TestJoinChatRoom(t *testing.T) {
	defer teardown()

	conn := createClient(t)
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	_, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("Failed to read welcome message: %v", err)
	}

	_, err = writer.WriteString("/create testroom\n")
	if err != nil {
		t.Fatalf("Failed to send create command: %v", err)
	}
	writer.Flush()

	_, err = reader.ReadString('\n')
	if err != nil {
		t.Fatalf("Failed to read response to create command: %v", err)
	}

	_, err = writer.WriteString("/join testroom\n")
	if err != nil {
		t.Fatalf("Failed to send join command: %v", err)
	}
	writer.Flush()

	message, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}
	if !strings.Contains(message, "Joined chat room \"testroom\"") {
		t.Fatalf("Unexpected response: %v", message)
	}
}

// 6 Test: Client can leave a chat room
func TestLeaveChatRoom(t *testing.T) {
	defer teardown()

	conn := createClient(t)
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	_, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("Failed to read welcome message: %v", err)
	}

	_, err = writer.WriteString("/create testroom\n")
	if err != nil {
		t.Fatalf("Failed to send create command: %v", err)
	}
	writer.Flush()

	_, err = reader.ReadString('\n')
	if err != nil {
		t.Fatalf("Failed to read response to create command: %v", err)
	}

	_, err = writer.WriteString("/join testroom\n")
	if err != nil {
		t.Fatalf("Failed to send join command: %v", err)
	}
	writer.Flush()

	_, err = reader.ReadString('\n')
	if err != nil {
		t.Fatalf("Failed to read response to join command: %v", err)
	}

	_, err = writer.WriteString("/leave\n")
	if err != nil {
		t.Fatalf("Failed to send leave command: %v", err)
	}
	writer.Flush()

	message, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}
	if !strings.Contains(message, "Left the chat room") {
		t.Fatalf("Unexpected response: %v", message)
	}
}

// 7 Test: Broadcast message to chat room
func TestBroadcastMessage(t *testing.T) {
	defer teardown()

	conn1 := createClient(t)
	defer conn1.Close()
	conn2 := createClient(t)
	defer conn2.Close()

	reader1 := bufio.NewReader(conn1)
	writer1 := bufio.NewWriter(conn1)
	reader2 := bufio.NewReader(conn2)
	writer2 := bufio.NewWriter(conn2)

	_, err := reader1.ReadString('\n')
	if err != nil {
		t.Fatalf("Failed to read welcome message: %v", err)
	}
	_, err = reader2.ReadString('\n')
	if err != nil {
		t.Fatalf("Failed to read welcome message: %v", err)
	}

	_, err = writer1.WriteString("/create testroom\n")
	if err != nil {
		t.Fatalf("Failed to send create command: %v", err)
	}
	writer1.Flush()

	_, err = reader1.ReadString('\n')
	if err != nil {
		t.Fatalf("Failed to read response to create command: %v", err)
	}

	_, err = writer1.WriteString("/join testroom\n")
	if err != nil {
		t.Fatalf("Failed to send join command: %v", err)
	}
	writer1.Flush()
	_, err = writer2.WriteString("/join testroom\n")
	if err != nil {
		t.Fatalf("Failed to send join command: %v", err)
	}
	writer2.Flush()

	_, err = reader1.ReadString('\n')
	if err != nil {
		t.Fatalf("Failed to read response to join command: %v", err)
	}
	_, err = reader2.ReadString('\n')
	if err != nil {
		t.Fatalf("Failed to read response to join command: %v", err)
	}

	_, err = writer1.WriteString("Hello everyone\n")
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}
	writer1.Flush()

	message, err := reader2.ReadString('\n')
	if err != nil {
		t.Fatalf("Failed to read broadcast message: %v", err)
	}
	if !strings.Contains(message, "Hello everyone") {
		t.Fatalf("Unexpected broadcast message: %v", message)
	}
}

// 8 Test: Save message to history
func TestSaveMessageToHistory(t *testing.T) {
	defer teardown()

	conn := createClient(t)
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	_, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("Failed to read welcome message: %v", err)
	}

	_, err = writer.WriteString("/create testroom\n")
	if err != nil {
		t.Fatalf("Failed to send create command: %v", err)
	}
	writer.Flush()

	_, err = reader.ReadString('\n')
	if err != nil {
		t.Fatalf("Failed to read response to create command: %v", err)
	}

	_, err = writer.WriteString("/join testroom\n")
	if err != nil {
		t.Fatalf("Failed to send join command: %v", err)
	}
	writer.Flush()

	_, err = reader.ReadString('\n')
	if err != nil {
		t.Fatalf("Failed to read response to join command: %v", err)
	}

	_, err = writer.WriteString("Test message for history\n")
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}
	writer.Flush()

	time.Sleep(1 * time.Second)

	history, err := os.ReadFile("chat_history.txt")
	if err != nil {
		t.Fatalf("Failed to read history file: %v", err)
	}
	if !strings.Contains(string(history), "Test message for history") {
		t.Fatalf("Message not found in history: %v", string(history))
	}
}

// 9 Test: Multiple clients in lobby
func TestMultipleClientsInLobby(t *testing.T) {
	defer teardown()

	conn1 := createClient(t)
	defer conn1.Close()
	conn2 := createClient(t)
	defer conn2.Close()

	_, err := bufio.NewReader(conn1).ReadString('\n')
	if err != nil {
		t.Fatalf("Failed to read welcome message: %v", err)
	}
	_, err = bufio.NewReader(conn2).ReadString('\n')
	if err != nil {
		t.Fatalf("Failed to read welcome message: %v", err)
	}
}

// 10 Failing test
func TestFailFast(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	t.Fatal("This test is designed to fail fast.")
}
