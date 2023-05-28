package main

import (
	"fmt"
	"log"
	"net/http"
	"github.com/gorilla/mux"
	title "upbase/title"
	// favicon "upbase/favicon"
	chatroom "upbase/chatroom"

	"github.com/gin-gonic/gin"
)


func main() {

	defer chatroom.DisconnectMongoClient()
	socket_router := gin.Default()

	// WebSocket endpoint for chat
	socket_router.GET("/ws/:roomID/:userID", chatroom.HandleWebSocket)

	// Start the server
	if err := socket_router.Run(":8080"); err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/title", title.GetHandler).Methods("POST")
	// router.HandleFunc("/server-ip/favicon", favicon.GetFaviconURLHandler).Methods("POST")
	// router.HandleFunc("/server-ip/public/files/{domainName}.png", favicon.GetFaviconImageHandler).Methods("GET")

	fmt.Println("Server listening on port 8000...")
	log.Fatal(http.ListenAndServe(":8000", router))
}