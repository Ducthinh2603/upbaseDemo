package main

import (
	// "fmt"
	"log"
	"net/http"
	// "net/http"
	// "github.com/gorilla/mux"
	// title "upbase/title"
	// favicon "upbase/favicon"
	chatroom "upbase/chatroom"
	db "upbase/database"
	user "upbase/user"

	"github.com/gin-gonic/gin"
)


func main() {

	defer db.DisconnectMongoClient()
	router := gin.Default()

	// WebSocket endpoint for chat
	router.GET("/:roomID/:userID", chatroom.HandleWebSocket)
	router.LoadHTMLGlob("templates/*.html")
	router.GET("/chatroom", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})
	// User registration and login endpoints
	router.POST("/users/register", user.RegisterUser)
	router.POST("/users/login", user.LoginUser)
	router.POST("/users/createChatroom", chatroom.CreateChatRoom)

	// Start the server
	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}

	// router := mux.NewRouter()
	// router.HandleFunc("/title", title.GetHandler).Methods("POST")
	// router.HandleFunc("/server-ip/favicon", favicon.GetFaviconURLHandler).Methods("POST")
	// router.HandleFunc("/server-ip/public/files/{domainName}.png", favicon.GetFaviconImageHandler).Methods("GET")

	// fmt.Println("Server listening on port 8000...")
	// log.Fatal(http.ListenAndServe(":8000", router))
}