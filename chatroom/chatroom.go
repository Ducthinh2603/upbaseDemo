package chatroom

import (
	"context"
	"database/sql"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
	db "upbase/database"

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
	Owner     string   `bson:"owner"`
	Members   []string `bson:"members,omitempty"`
	CreatedAt time.Time
}

// Define the WebSocket message struct
type WebSocketMessage struct {
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

	var requestBody struct {
        UserID string `json:"user_id"`
    }

    // Bind the request body to the struct
    if err := c.ShouldBindJSON(&requestBody); err != nil {
		if err == io.EOF {
			log.Println("Starting connection!")
		} else {
        // Handle the error, e.g., return an error response
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
    }
	userID := requestBody.UserID
	log.Printf("roomId: %s, userId: %s\n", roomID, userID)
	// roomExists, err := verifyRoomIdByUserId(roomID, userID)
	roomExists, err := verifyRoomId(roomID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
	}
	if !roomExists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Room doesn't exist"})
	}

	// This is an attempt to solve "websocket: request origin not allowed by Upgrader.CheckOrigin"
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

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
	handleChat(conn, roomID)
}

func verifyRoomIdByUserId(roomId, userID string) (bool, error) {
	statement := "SELECT id, owner_id FROM upbase_chat_rooms WHERE owner_id = $1 AND id = $2"
	row := db.PgDb.QueryRow(statement, userID, roomId)

	var chatRoom ChatRoom
	err := row.Scan(&chatRoom.ID, &chatRoom.Owner)
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err // Database error
	}

	return true, nil
}

func verifyRoomId(roomId string) (bool, error) {
	statement := "SELECT id, owner_id FROM upbase_chat_rooms WHERE id = $1"
	row := db.PgDb.QueryRow(statement, roomId)

	var chatRoom ChatRoom
	err := row.Scan(&chatRoom.ID, &chatRoom.Owner)
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err // Database error
	}

	return true, nil
}

func handleChat(conn *websocket.Conn, roomID string) {
	// Read messages from the WebSocket connection
	for {
		var msg WebSocketMessage
		err := conn.ReadJSON(&msg)
		
		if err != nil {
			log.Println("Failed to read WebSocket message:", err)
			break
		}
		log.Printf("%s send: %s\n", conn.RemoteAddr(), msg.Message)

		// Create a ChatMessage struct
		chatMessage := ChatMessage{
			RoomID:    roomID,
			Sender:    msg.Sender,
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
		broadcastMessage(conn, roomID, chatMessage)
		// broadcastMessage(roomID, msg, msgType)
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
	collection := db.MongoClient.Database("upbase_chatroom").Collection("test")
	_, err := collection.InsertOne(context.TODO(), message)
	return err
}

func broadcastMessage(c *websocket.Conn, roomID string, message ChatMessage) {
	connectionsLock.Lock()
	defer connectionsLock.Unlock()

	for conn := range connectionsByRoom[roomID] {
		err := conn.WriteJSON(message)
		if err != nil {
			log.Println("Failed to send WebSocket message:", err)
		}
	}
}


// func broadcastMessage(roomID string, message []byte, messageType int) {
// 	connectionsLock.Lock()
// 	defer connectionsLock.Unlock()

// 	for conn := range connectionsByRoom[roomID] {
// 		err := conn.WriteMessage(messageType, message)
// 		if err != nil {
// 			log.Println("Failed to send WebSocket message:", err)
// 		}
// 	}
// }