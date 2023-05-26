package favicon

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

var db *sql.DB
var err error

func init() {
	db, err = sql.Open("postgres", GetDatabaseConfig())
	if err != nil {
		log.Fatal(err)
	}
	statement := ""
	f, err := os.Open("favicon/init.sql")
	if err != nil {
		log.Fatal("Can't open entry point: ", err)
	}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		statement += scanner.Text() + "\n"
	}
	_, err = db.Exec(statement)
	if err != nil {
		log.Fatal("Can't initiate database: ", err)
	}
}

type FaviconRequest struct {
	Link string `json:"link"`
}

type FaviconResponse struct {
	FaviconURL string `json:"favicon_url"`
}

func GetFaviconURLHandler(w http.ResponseWriter, r *http.Request) {
	var request FaviconRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error parsing request body: %s", err.Error())
		return
	}

	domainName, err := extractDomain(request.Link)

	log.Printf("Type of domainName: %T", domainName)
	if err != nil || strings.TrimSpace(domainName) == "" {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Enter")
		fmt.Fprintf(w, "Can't extract domain name from URL!")
		return
	}

	isDownloaded, err := isFaviconDownloaded(domainName)
	if err != nil {
		log.Println("Can't query from database!!!")
	}

	if isDownloaded {
		log.Printf("%s doesn't need to be downloaded", domainName)
		response := FaviconResponse{FaviconURL: contructFaviconURL(r.Host, domainName)}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		return
	}

	log.Println("Saving image for", domainName, "...")
	err = saveURLFavicon(domainName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, err)
		return

	}
	log.Println("Saving completed!")
	response := FaviconResponse{FaviconURL: contructFaviconURL(r.Host, domainName)}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

}

func GetFaviconImageHandler(w http.ResponseWriter, r *http.Request) {
	var buffer bytes.Buffer
	vars := mux.Vars(r)
	domainName := vars["domainName"]
	filePath := fmt.Sprintf("favicon/image/%s.png", domainName)
	f, err := os.Open(filePath)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(`{"message" : "This URL doesn't exist"}`)
		return
	}

	_, err = buffer.ReadFrom(f)
	if err != nil {
		http.Error(w, "Failed to read image", http.StatusInternalServerError)
		return
	}

	imageData := buffer.Bytes()
	w.Header().Set("Content-Type", "image/png")
	_, err = w.Write(imageData)
	if err != nil {
		http.Error(w, "Failed to write image", http.StatusInternalServerError)
		return
	}
}

func extractDomain(urlStr string) (string, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	domain := parsedURL.Host
	return domain, nil
}

func saveImageToFile(domainName string, resp *http.Response) error {
	file, err := os.Create("favicon/image/" + domainName + ".png")
	if err != nil {
		log.Fatalf("Can't create file image: %s", err)
		return err
	}
	defer file.Close()

	// Copy the response body to the file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		log.Fatal("Can't save image to file: ", err)
		return err
	}
	log.Println("Image downloaded and saved to " + domainName + "image.png")
	return nil
}

func saveImageToDB(domainName string, resp *http.Response) error {
	var buffer bytes.Buffer
	_, err = buffer.ReadFrom(resp.Body)
	if err != nil {
		log.Fatal("Can't read the image: ", err)
		return err
	}

	// Get the byte slice from the buffer
	imageData := buffer.Bytes()
	_, err = db.Exec("INSERT INTO favicon (domain_name, image_data) VALUES ($1, $2)", domainName, imageData)
	if err != nil {
		log.Fatal("Can't save the image to DB: ", err)
		return err
	}
	log.Printf("Image %s saved to DB\n", domainName)
	return nil
}

func saveURLFavicon(domainName string) error {
	url := fmt.Sprintf("https://www.google.com/s2/favicons?domain=%s&sz=256", domainName)

	resp, err := http.Get(url)
	if err != nil {
		log.Fatal("Something wrong with ", url)
		return err
	}
	defer resp.Body.Close()

	// Attemp to using goroutine but it fails to save to file.

	// errCh := make(chan error)
	// doneCh := make(chan error)

	// go func() {
	// 	err := saveImageToFile(domainName, resp)
	// 	errCh <- err
	// 	log.Println("Finish save to file")
	// 	if err != nil {
	// 		log.Fatal("Can't save image to file!")
	// 	}
	// }()

	// go func() {
	// 	err := saveImageToDB(domainName, resp)
	// 	errCh <- err
	// 	log.Println("Finish save to DB")
	// 	if err != nil {
	// 		log.Fatal("Can't save image to file!")
	// 	}
	// }()

	// // Wait for Goroutines to finish
	// go func() {
	// 	err1 := <- errCh
	// 	err2 := <- errCh
	// 	log.Println("Receive result from saving action!")
	// 	if err1 != nil && err2 != nil {
	// 		doneCh <- fmt.Errorf("Server can't save image!!")
	// 	} else {
	// 		doneCh <- nil
	// 	}
	// }()

	// // Wait for all Goroutines to finish
	// err = <-doneCh
	// log.Println("Poll successfully from done channel!")
	// return err

	err1 := saveImageToFile(domainName, resp)
	if err1 != nil {
		log.Fatal("Can't save image to file!")
	}
	err2 := saveImageToDB(domainName, resp)
	if err2 != nil {
		log.Fatal("Can't save image to file!")
	}
	if err1 != nil && err2 != nil {
		return fmt.Errorf("Server can't save image!!")
	}
	return nil
}

func contructFaviconURL(domainName string, fileName string) string {
	return fmt.Sprintf("http://%s/server-ip/public/files/%s.png", domainName, fileName)
}

func isFaviconDownloaded(domainName string) (bool, error) {
	var name string
	err := db.QueryRow("SELECT domain_name FROM favicon WHERE domain_name = $1", domainName).Scan(&name)
	if err != nil {
		return false, err
	}
	if name == "" {
		return false, nil
	}
	return true, nil
}
