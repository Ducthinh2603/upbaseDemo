package main

import (
	"fmt"
	"log"
	"net/http"
	"github.com/gorilla/mux"
	title "upbase/title"
	favicon "upbase/favicon"
)

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/title", title.GetHandler).Methods("POST")
	router.HandleFunc("/server-ip/favicon", favicon.GetFaviconURLHandler).Methods("POST")
	router.HandleFunc("/server-ip/public/files/{domainName}.png", favicon.GetFaviconImageHandler).Methods("GET")

	fmt.Println("Server listening on port 8000...")
	log.Fatal(http.ListenAndServe(":8000", router))
}