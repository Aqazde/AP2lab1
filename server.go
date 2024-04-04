package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	CONN_PORT = ":3335"
	CONN_TYPE = "tcp"
)

var (
	lobby        = NewLobby()
	historyFile  *os.File
	historyMutex sync.Mutex
	chatRooms    = make(map[string]*ChatRoom)
	chatMutex    sync.Mutex
)

func init() {
	var err error
	historyFile, err = os.Create("chat_history.txt")
	if err != nil {
		log.Fatal(err)
	}
}

type Client struct {
	conn     net.Conn
	reader   *bufio.Reader
	writer   *bufio.Writer
	userName string
	chatRoom *ChatRoom
}

func NewClient(conn net.Conn) *Client {
	return &Client{
		conn:     conn,
		reader:   bufio.NewReader(conn),
		writer:   bufio.NewWriter(conn),
		userName: "Anonymous",
	}
}

func (c *Client) JoinLobby() {
	lobby.Join(c)
}

func (c *Client) Listen() {
	defer func() {
		c.conn.Close()
		lobby.Remove(c)
		if c.chatRoom != nil {
			c.chatRoom.RemoveClient(c)
		}
	}()

	for {
		message, err := c.reader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading from client [%s]: %v", c.userName, err)
			return
		}
		message = strings.TrimSpace(message)
		if strings.HasPrefix(message, "/") {
			c.processCommand(message)
		} else {
			saveMessageToHistory(message)
			c.chatRoom.Broadcast(fmt.Sprintf("[%s] %s: %s", time.Now().Format("2006-01-02 15:04:05"), c.userName, message))
		}
	}
}

func (c *Client) processCommand(command string) {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return
	}

	switch parts[0] {
	case "/join":
		if len(parts) != 2 {
			c.sendMessage("Usage: /join <chat_name>")
			return
		}
		roomName := parts[1]
		chatMutex.Lock()
		room, exists := chatRooms[roomName]
		chatMutex.Unlock()
		if !exists {
			c.sendMessage(fmt.Sprintf("Error: A chat room with name '%s' does not exist.", roomName))
			return
		}
		if c.chatRoom != nil {
			c.chatRoom.RemoveClient(c)
		}
		c.chatRoom = room
		room.AddClient(c)
		c.sendMessage(fmt.Sprintf("Notice: \"%s\" joined the chat room.", c.userName))
		room.Broadcast(fmt.Sprintf("Notice: \"%s\" joined the chat room.", c.userName))
	case "/create":
		if len(parts) != 2 {
			c.sendMessage("Usage: /create <chat_name>")
			return
		}
		roomName := parts[1]
		chatMutex.Lock()
		_, exists := chatRooms[roomName]
		if exists {
			c.sendMessage(fmt.Sprintf("Error: Chat room '%s' already exists.", roomName))
			chatMutex.Unlock()
			return
		}
		room := NewChatRoom(roomName)
		chatRooms[roomName] = room
		chatMutex.Unlock()
		c.sendMessage(fmt.Sprintf("Notice: Created chat room \"%s\".", roomName))
	case "/leave":
		if c.chatRoom != nil {
			c.chatRoom.Broadcast(fmt.Sprintf("Notice: \"%s\" has left the chat room.", c.userName))
			c.chatRoom.RemoveClient(c)
			c.chatRoom = nil
		}
	default:
		c.sendMessage(fmt.Sprintf("Unknown command: %s", parts[0]))
	}
}

func (c *Client) sendMessage(message string) {
	c.writer.WriteString(message + "\n")
	c.writer.Flush()
}

type Lobby struct {
	clients []*Client
	sync.RWMutex
}

func NewLobby() *Lobby {
	return &Lobby{}
}

func (l *Lobby) Join(client *Client) {
	l.Lock()
	defer l.Unlock()
	l.clients = append(l.clients, client)
	go client.Listen()
}

func (l *Lobby) Remove(client *Client) {
	l.Lock()
	defer l.Unlock()
	for i, c := range l.clients {
		if c == client {
			// This removes the client from the lobby slice.
			l.clients = append(l.clients[:i], l.clients[i+1:]...)
			break
		}
	}
}

// Broadcast sends a message to all clients in the lobby
func (l *Lobby) Broadcast(message string) {
	l.RLock()
	defer l.RUnlock()
	for _, client := range l.clients {
		client.sendMessage(message)
	}
}

// ChatRoom represents a chat room
type ChatRoom struct {
	Name    string
	Clients []*Client
	sync.RWMutex
}

// NewChatRoom creates a new ChatRoom instance
func NewChatRoom(name string) *ChatRoom {
	return &ChatRoom{Name: name}
}

// AddClient adds a new client to the chat room
func (cr *ChatRoom) AddClient(client *Client) {
	cr.Lock()
	defer cr.Unlock()
	cr.Clients = append(cr.Clients, client)
}

// RemoveClient removes a client from the chat room
func (cr *ChatRoom) RemoveClient(client *Client) {
	cr.Lock()
	defer cr.Unlock()
	for i, c := range cr.Clients {
		if c == client {
			cr.Clients = append(cr.Clients[:i], cr.Clients[i+1:]...)
			break
		}
	}
}

// Broadcast sends a message to all clients in the chat room
func (cr *ChatRoom) Broadcast(message string) {
	cr.RLock()
	defer cr.RUnlock()
	for _, client := range cr.Clients {
		client.sendMessage(message)
	}
}

func saveMessageToHistory(message string) {
	historyMutex.Lock()
	defer historyMutex.Unlock()
	_, err := historyFile.WriteString(message + "\n")
	if err != nil {
		log.Printf("Error writing to history file: %v", err)
	}
}

func main() {
	defer historyFile.Close()

	// Start listening on the specified port
	listener, err := net.Listen(CONN_TYPE, CONN_PORT)
	if err != nil {
		log.Fatal("Error starting TCP server:", err)
	}
	defer listener.Close()
	log.Println("Listening on " + CONN_PORT)

	// Main loop to accept incoming connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}

		client := NewClient(conn)
		client.JoinLobby()
		client.sendMessage("Welcome to the server! Type \"/help\" to get a list of commands.")
	}
}
