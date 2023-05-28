package chatroom

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}


func HandleWebSocket(c *gin.Context) {
	roomID := c.Param("roomID")
	userID := c.Param("userID")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Failed to set up WebSocket connection:", err)
		return
	}

	// Create a new goroutine to handle the WebSocket connection
	go handleChat(conn, roomID, userID)
}

func handleChat(conn *websocket.Conn, roomID, userID string) {
	// Register the WebSocket connection
	// You can use a map to store all active connections for each room
	// For simplicity, this example uses a single global map
	// Consider using a more robust solution for production
	connections := make(map[string]*websocket.Conn)

	connections[userID] = conn

	defer func() {
		// Clean up the WebSocket connection when the goroutine exits
		delete(connections, userID)
		conn.Close()
	}()

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
		broadcastMessage(connections, roomID, chatMessage)
	}
}

func insertChatMessage(message ChatMessage) error {
	collection := mongoClient.Database("upbase_chatroom").Collection("test")
	_, err := collection.InsertOne(nil, message)
	return err
}

func broadcastMessage(connections map[string]*websocket.Conn, roomID string, message ChatMessage) {
	// Iterate over all connections in the room and send the message
	for _, conn := range connections {
		err := conn.WriteJSON(message)
		if err != nil {
			log.Println("Failed to send WebSocket message:", err)
		}
	}
}

