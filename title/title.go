package title

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	
)

type TitleResponse struct {
	Title string `json:"title"`
}

type TitleRequest struct {
	Link string `json:"link"`
}

func GetHandler(w http.ResponseWriter, r *http.Request) {
	var request TitleRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error parsing request body: %s", err.Error())
		return
	}
	log.Println("Get title for: ", request.Link)

	title, err := getTitle(request.Link)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error: %s", err.Error())
		return
	}

	response := TitleResponse{Title: title}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func getTitle(link string) (string, error) {
	resp, err := http.Get(link)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	title := strings.TrimSpace(doc.Find("title").Text())
	return title, nil
}
