package chatroom

import (
	"log"
	"net/http"
	"time"
	db "upbase/database"

	"github.com/gin-gonic/gin"
)

// CreateChatRoom handles the creation of a new chat room
func CreateChatRoom(c *gin.Context) {
	var chatRoom ChatRoom
	if err := c.ShouldBindJSON(&chatRoom); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Insert the chat room into the database
	chatRoomID, err := insertChatRoom(chatRoom.Owner)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create chat room"})
		return
	}

	log.Printf("Chat room %s created successfully\n", chatRoomID)
	c.JSON(http.StatusOK, gin.H{"chat_room_id": chatRoomID, "message": "Chat room created successfully"})
}


func insertChatRoom(owner_id string) (string, error) {
	query := "INSERT INTO upbase_chat_rooms (owner_id, created_at) VALUES ($1, $2) RETURNING id"
	row := db.PgDb.QueryRow(query, owner_id, time.Now())

	var chatRoomID string
	err := row.Scan(&chatRoomID)
	if err != nil {
		log.Println("Failed to insert chat room:", err)
		return "", err
	}

	return chatRoomID, nil
}

