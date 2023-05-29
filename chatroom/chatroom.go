package chatroom

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// Define the ChatMessage struct
type ChatMessage struct {
	ID        string    `bson:"_id,omitempty"`
	RoomID    string    `bson:"room_id"`
	Sender    string    `bson:"sender"`
	Message   string    `bson:"message"`
	Timestamp time.Time `bson:"timestamp"`
}

// Define the ChatRoom struct
type ChatRoom struct {
	ID        string   `bson:"_id,omitempty"`
	Name      string   `bson:"name"`
	Owner     string   `bson:"owner"`
	Members   []string `bson:"members"`
	CreatedAt time.Time
}

// Define the WebSocket message struct
type WebSocketMessage struct {
	RoomID  string `json:"room_id"`
	Sender  string `json:"sender"`
	Message string `json:"message"`
}

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	connectionsByRoom = make(map[string]map[*websocket.Conn]bool)
	connectionsLock   sync.Mutex
)

func HandleWebSocket(c *gin.Context) {
	roomID := c.Param("roomID")
	userID := c.Param("userID")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Failed to set up WebSocket connection:", err)
		return
	}

	// Register the WebSocket connection
	registerConnection(roomID, conn)

	defer func() {
		// Clean up the WebSocket connection when the goroutine exits
		unregisterConnection(roomID, conn)
		conn.Close()
	}()

	// Create a new goroutine to handle the WebSocket connection
	go handleChat(conn, roomID, userID)
}

func handleChat(conn *websocket.Conn, roomID, userID string) {
	// Read messages from the WebSocket connection
	for {
		var msg WebSocketMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println("Failed to read WebSocket message:", err)
			break
		}

		// Create a ChatMessage struct
		chatMessage := ChatMessage{
			RoomID:    roomID,
			Sender:    userID,
			Message:   msg.Message,
			Timestamp: time.Now(),
		}

		// Insert the message into the MongoDB collection
		err = insertChatMessage(chatMessage)
		if err != nil {
			log.Println("Failed to insert chat message:", err)
			break
		}

		// Broadcast the message to all connected clients in the same room
		broadcastMessage(roomID, chatMessage)
	}
}

func registerConnection(roomID string, conn *websocket.Conn) {
	connectionsLock.Lock()
	defer connectionsLock.Unlock()

	if connectionsByRoom[roomID] == nil {
		connectionsByRoom[roomID] = make(map[*websocket.Conn]bool)
	}

	connectionsByRoom[roomID][conn] = true
}

func unregisterConnection(roomID string, conn *websocket.Conn) {
	connectionsLock.Lock()
	defer connectionsLock.Unlock()

	if connectionsByRoom[roomID] != nil {
		delete(connectionsByRoom[roomID], conn)
		if len(connectionsByRoom[roomID]) == 0 {
			delete(connectionsByRoom, roomID)
		}
	}
}

func insertChatMessage(message ChatMessage) error {
	collection := mongoClient.Database("upbase_chatroom").Collection("test")
	_, err := collection.InsertOne(context.TODO(), message)
	return err
}

func broadcastMessage(roomID string, message ChatMessage) {
	connectionsLock.Lock()
	defer connectionsLock.Unlock()

	for conn := range connectionsByRoom[roomID] {
		err := conn.WriteJSON(message)
		if err != nil {
			log.Println("Failed to send WebSocket message:", err)
		}
	}
}

