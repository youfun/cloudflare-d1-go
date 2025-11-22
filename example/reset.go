package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	cloudflare_d1_go "github.com/youfun/cloudflare-d1-go/client"
)

func main() {
	if err := godotenv.Load(".env"); err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")
	apiToken := os.Getenv("CLOUDFLARE_API_TOKEN")
	dbName := os.Getenv("CLOUDFLARE_DB_NAME")

	if accountID == "" || apiToken == "" || dbName == "" {
		log.Fatal("Please set CLOUDFLARE_ACCOUNT_ID, CLOUDFLARE_API_TOKEN, and CLOUDFLARE_DB_NAME environment variables")
	}

	client := cloudflare_d1_go.NewClient(accountID, apiToken)
	if err := client.ConnectDB(dbName); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	tables := []string{"user_departments", "users", "departments", "d1_migrations"}
	for _, t := range tables {
		fmt.Printf("Dropping table %s...\n", t)
		_, err := client.RemoveTable(t)
		if err != nil {
			fmt.Printf("Failed to drop table %s: %v\n", t, err)
		}
	}
	fmt.Println("Database reset complete.")
}
