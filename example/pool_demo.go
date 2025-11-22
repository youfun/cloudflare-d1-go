package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/joho/godotenv"
	cloudflare_d1_go "github.com/youfun/cloudflare-d1-go/client"
)

func init() {
	if err := godotenv.Load(".env"); err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: Error loading .env file: %v", err)
	}
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())
}

type User struct {
	ID    int    `db:"id"`
	Name  string `db:"name"`
	Age   int    `db:"age"`
	Email string `db:"email"`
}

func main() {
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")
	apiToken := os.Getenv("CLOUDFLARE_API_TOKEN")

	if accountID == "" || apiToken == "" {
		log.Fatal("Please set CLOUDFLARE_ACCOUNT_ID and CLOUDFLARE_API_TOKEN environment variables")
	}

	// Create a connection pool (like sqlx.Open)
	fmt.Println("=== ConnectionPool Demo (Similar to sqlx.DB) ===\n")

	pool := cloudflare_d1_go.NewConnectionPool(accountID, apiToken)
	if pool == nil {
		log.Fatal("Failed to create connection pool")
	}

	// Set cache age to 1 hour (default is 24 hours)
	pool.SetCacheAge(1 * time.Hour)
	fmt.Println("✓ ConnectionPool created with 1-hour cache")

	// Connect to database (like pool.Connect("db_name"))
	// First time: calls API to fetch database ID
	fmt.Println("\n--- First connection (API call) ---")
	start := time.Now()
	err := pool.Connect("test")
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	elapsed := time.Since(start)
	fmt.Printf("✓ Connected to 'test' database in %v\n", elapsed)

	// Check if it's cached
	isCached := pool.IsCached("test")
	fmt.Printf("  - Is cached: %v\n", isCached)
	fmt.Printf("  - Cache info: %+v\n", pool.GetCacheInfo("test"))

	// Second connection to same database (from cache)
	// Should be instant, no API call
	fmt.Println("\n--- Second connection (from cache) ---")
	start = time.Now()
	err = pool.Connect("test")
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	elapsed = time.Since(start)
	fmt.Printf("✓ Connected to 'test' database in %v (from cache)\n", elapsed)

	// Execute queries like sqlx
	fmt.Println("\n=== Query Operations ===")

	// Create table
	fmt.Println("\n1. Creating table...")
	createTableQuery := "CREATE TABLE IF NOT EXISTS pool_demo (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, age INTEGER, email TEXT UNIQUE);"
	res, err := pool.CreateTable(createTableQuery)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}
	if !res.Success {
		log.Fatalf("Create table failed: %v", res.Errors)
	}
	fmt.Println("✓ Table created")

	// Insert data
	fmt.Println("\n2. Inserting data...")
	insertQuery := "INSERT INTO pool_demo (name, age, email) VALUES (?, ?, ?);"

	// Insert data with random suffix to avoid conflicts
	users := []struct {
		name  string
		age   int
		email string
	}{
		{"Alice", 30, "alice"},
		{"Bob", 25, "bob"},
		{"Charlie", 35, "charlie"},
	}

	for _, user := range users {
		// Generate random suffix to avoid email conflicts
		randomSuffix := rand.Intn(1000000)
		email := fmt.Sprintf("%s.%d@example.com", user.email, randomSuffix)

		res, err := pool.Query(insertQuery, []string{user.name, fmt.Sprintf("%d", user.age), email})
		if err != nil {
			log.Fatalf("Insert failed: %v", err)
		}
		result, _ := res.ToResult()
		lastID, _ := result.LastInsertId()
		fmt.Printf("  ✓ Inserted %s (ID: %d)\n", user.name, lastID)
	}

	// Query data using new sqlx-style Select method
	fmt.Println("\n3. Querying data with sqlx-style Select...")
	var allUsers []User
	err = pool.Select(&allUsers, "SELECT id, name, age, email FROM pool_demo ORDER BY age ASC")
	if err != nil {
		log.Fatalf("Select failed: %v", err)
	}

	fmt.Printf("Found %d users:\n", len(allUsers))
	for _, u := range allUsers {
		fmt.Printf("  - %s (Age: %d, Email: %s)\n", u.Name, u.Age, u.Email)
	}

	// Multiple conditions query using new sqlx-style Select (like sqlx)
	fmt.Println("\n4. Advanced query with Select (multiple WHERE conditions)...")
	var advancedUsers []User
	err = pool.Select(&advancedUsers, "SELECT id, name, age, email FROM pool_demo WHERE age > ? AND age < ? ORDER BY age ASC", 24, 33)
	if err != nil {
		log.Fatalf("Select failed: %v", err)
	}

	fmt.Println("Users aged between 24 and 33:")
	for _, u := range advancedUsers {
		fmt.Printf("  - %s (Age: %d)\n", u.Name, u.Age)
	}

	// Update with Exec (sqlx-style)
	fmt.Println("\n5. Update with sqlx-style Exec...")
	updateQuery := "UPDATE pool_demo SET age = ? WHERE age > ? AND name != ?;"
	rowsAffected, err := pool.Exec(updateQuery, 99, 28, "Alice")
	if err != nil {
		log.Fatalf("Exec failed: %v", err)
	}
	fmt.Printf("✓ Updated %d users\n", rowsAffected)

	// QueryOne using sqlx-style Get
	fmt.Println("\n6. Get single user with sqlx-style Get...")
	var singleUser User
	err = pool.Get(&singleUser, "SELECT id, name, age, email FROM pool_demo WHERE name = ?", "Alice")
	if err != nil {
		log.Fatalf("Get failed: %v", err)
	}
	fmt.Printf("✓ Found user: %s (Age: %d, Email: %s)\n", singleUser.Name, singleUser.Age, singleUser.Email)

	// QueryAll with multiple conditions
	fmt.Println("\n7. Select with multiple conditions (age > 25)...")
	var filteredUsers []User
	err = pool.Select(&filteredUsers, "SELECT id, name, age, email FROM pool_demo WHERE age > ? ORDER BY age ASC", 25)
	if err != nil {
		log.Fatalf("Select failed: %v", err)
	}
	fmt.Printf("✓ Found %d users with age > 25:\n", len(filteredUsers))
	for _, u := range filteredUsers {
		fmt.Printf("  - %s (Age: %d)\n", u.Name, u.Age)
	}

	// Show cache statistics
	fmt.Println("\n=== Cache Statistics ===")
	cachedDBs := pool.ListCachedDatabases()
	fmt.Printf("Cached databases: %v\n", cachedDBs)
	for _, dbName := range cachedDBs {
		info := pool.GetCacheInfo(dbName)
		fmt.Printf("  - %s: ID=%s, cached at %v\n", info.Name, info.DatabaseID, info.CachedAt.Format("15:04:05"))
	}

	// Advanced parameter type conversion examples
	fmt.Println("\n=== Advanced Parameter Type Conversion ===")

	// Example 1: String parameter with Get
	fmt.Println("\n8. Get with string parameter...")
	var stringParamUser User
	err = pool.Get(&stringParamUser, "SELECT id, name, age, email FROM pool_demo WHERE name = ?", "Bob")
	if err != nil {
		log.Printf("Get with string param failed: %v", err)
	} else {
		fmt.Printf("✓ Found: %s (Age: %d)\n", stringParamUser.Name, stringParamUser.Age)
	}

	// Example 2: Integer parameters with Select
	fmt.Println("\n9. Select with integer parameters...")
	var intParamUsers []User
	err = pool.Select(&intParamUsers, "SELECT id, name, age, email FROM pool_demo WHERE age > ? AND age < ? ORDER BY age ASC", 24, 33)
	if err != nil {
		log.Printf("Select with int params failed: %v", err)
	} else {
		fmt.Printf("✓ Found %d users between age 24-33\n", len(intParamUsers))
	}

	// Example 3: Multiple parameter types with Select
	fmt.Println("\n10. Select with multiple parameter types...")
	var mixedUsers []User
	err = pool.Select(&mixedUsers, "SELECT id, name, age, email FROM pool_demo WHERE age > ?", 25)
	if err != nil {
		log.Printf("Select with mixed params failed: %v", err)
	} else {
		fmt.Printf("✓ Found %d users with age > 25\n", len(mixedUsers))
	}

	// Example 4: Exec with multiple parameters
	fmt.Println("\n11. Exec with multiple parameters...")
	affected, err := pool.Exec("UPDATE pool_demo SET age = ? WHERE age = ?", 26, 25)
	if err != nil {
		log.Printf("Exec failed: %v", err)
	} else {
		fmt.Printf("✓ Updated %d users\n", affected)
	}

	// Example 5: Time.Time parameter
	fmt.Println("\n12. Time.Time parameter support...")
	now := time.Now()
	fmt.Printf("✓ Current time formatted: %s\n", now.Format("2006-01-02 15:04:05"))
	fmt.Println("  (In real usage, pass time.Time directly to Select/Get/Exec)")

	// Cleanup
	fmt.Println("\n=== Cleanup ===")
	pool.RemoveTable("pool_demo")
	pool.ClearAllCache()
	fmt.Println("✓ Table dropped and cache cleared")
}
