package database

import (
	"fmt"
	// "log"
	// "os"
	// "strconv"
)

// func GetDatabaseConfig() string {
// 	host := os.Getenv("POSTGRES_HOST")
// 	port, err := strconv.Atoi(os.Getenv("POSTGRES_PORT"))
// 	if err != nil {
// 		fmt.Println("Error converting string to int:", err)
// 		log.Panic("Problem with configuration!")
// 		os.Exit(1)
// 	}
// 	user := os.Getenv("POSTGRES_USER")
// 	password := os.Getenv("POSTGRES_PASSWORD")
// 	databaseName := os.Getenv("DATABASE_NAME")
	
	
// 	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, databaseName)
// }

func GetDatabaseConfig() string {
	host := "localhost"
	port := 1337
	user := "postgres"
	password := "postgres"
	databaseName := "upbase"
	
	
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, databaseName)
}
