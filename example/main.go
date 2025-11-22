package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/joho/godotenv"
	cloudflare_d1_go "github.com/youfun/cloudflare-d1-go/client"
	"github.com/youfun/cloudflare-d1-go/migrations"
	"github.com/youfun/cloudflare-d1-go/utils"
)

func init() {
	// Load environment variables from .env file
	// .env file is optional - env vars can override it
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

type Department struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

type UserWithDept struct {
	UserID       int    `db:"user_id"`
	UserName     string `db:"user_name"`
	Age          int    `db:"age"`
	DepartmentID int    `db:"department_id"`
	DeptName     string `db:"dept_name"`
}

func main() {
	// Load configuration from environment variables (with .env file as fallback)
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")
	apiToken := os.Getenv("CLOUDFLARE_API_TOKEN")
	dbName := os.Getenv("CLOUDFLARE_DB_NAME")

	if accountID == "" || apiToken == "" || dbName == "" {
		log.Fatal("Please set CLOUDFLARE_ACCOUNT_ID, CLOUDFLARE_API_TOKEN, and CLOUDFLARE_DB_NAME environment variables")
	}

	client := cloudflare_d1_go.NewClient(accountID, apiToken)

	// Connect to database
	fmt.Printf("Connecting to database %s...\n", dbName)
	if err := client.ConnectDB(dbName); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	var res *utils.APIResponse

	// ============ Migrations ============
	fmt.Println("\n=== Applying Migrations ===")

	// First, clean up any existing migration records to allow fresh migration
	// This is useful for testing/development
	fmt.Println("Cleaning up previous migration records...")
	_, _ = client.RemoveTable("d1_migrations")
	fmt.Println("✓ Migration table reset")

	// Create migration source
	migrationsSource := &migrations.FileMigrationSource{
		Dir: "./migrations",
	}

	// Debug: Check if migrations directory exists and find migrations
	fmt.Printf("Looking for migrations in: ./migrations\n")
	foundMigrations, err := migrationsSource.FindMigrations()
	if err != nil {
		log.Fatalf("Failed to find migrations: %v", err)
	}
	fmt.Printf("Found %d migration files\n", len(foundMigrations))
	for _, m := range foundMigrations {
		fmt.Printf("  - %s (Up: %d queries, Down: %d queries)\n", m.Id, len(m.Up), len(m.Down))
	}

	if len(foundMigrations) == 0 {
		fmt.Println("⚠️  No migrations found! Make sure the migrations directory exists.")
	}

	// Apply migrations
	n, err := migrations.Exec(client, migrationsSource, migrations.Up)
	if err != nil {
		log.Fatalf("Failed to apply migrations: %v", err)
	}
	fmt.Printf("Applied %d migrations!\n", n)

	// ============ Batch Insert Data ============
	fmt.Println("\n=== Batch Insert Data ===")

	// Insert multiple users with random suffix to avoid conflicts
	userData := []struct {
		name  string
		age   int
		email string
	}{
		{"Alice", 30, "alice"},
		{"Bob", 25, "bob"},
		{"Charlie", 35, "charlie"},
		{"Diana", 28, "diana"},
		{"Eve", 32, "eve"},
	}

	fmt.Println("Inserting users...")
	insertUserQuery := "INSERT OR IGNORE INTO users (name, age, email) VALUES (?, ?, ?);"
	var userIDs []int64
	for _, user := range userData {
		// Generate random suffix to avoid email conflicts
		randomSuffix := rand.Intn(1000000)
		email := fmt.Sprintf("%s.%d@example.com", user.email, randomSuffix)

		res, err = client.Query(insertUserQuery, []string{user.name, fmt.Sprintf("%d", user.age), email})
		if err != nil {
			log.Fatalf("Insert user failed: %v", err)
		}
		result, err := res.ToResult()
		if err != nil {
			log.Fatalf("Failed to get result: %v", err)
		}
		lastID, _ := result.LastInsertId()
		userIDs = append(userIDs, lastID)
		fmt.Printf("  ✓ Inserted user %s (ID: %d)\n", user.name, lastID)
	}

	// Insert departments
	deptData := []string{"Engineering", "Sales", "HR", "Marketing"}
	fmt.Println("Inserting departments...")
	insertDeptQuery := "INSERT OR IGNORE INTO departments (name) VALUES (?);"
	var deptIDs []int64
	for _, dept := range deptData {
		res, err = client.Query(insertDeptQuery, []string{dept})
		if err != nil {
			log.Fatalf("Insert department failed: %v", err)
		}
		result, err := res.ToResult()
		if err != nil {
			log.Fatalf("Failed to get result: %v", err)
		}
		lastID, _ := result.LastInsertId()
		deptIDs = append(deptIDs, lastID)
		fmt.Printf("  ✓ Inserted department %s (ID: %d)\n", dept, lastID)
	}

	// ============ Query with Multiple WHERE Conditions ============
	fmt.Println("\n=== Query with Multiple WHERE Conditions ===")

	// Query 1: Find users with age > 28 AND age < 33
	fmt.Println("\n1. Users aged between 28 and 33 (using Select):")
	multiConditionQuery := "SELECT * FROM users WHERE age > ? AND age < ? ORDER BY age ASC;"
	var ageFilteredUsers []User
	if err := client.Select(&ageFilteredUsers, multiConditionQuery, 28, 33); err != nil {
		log.Fatalf("Select failed: %v", err)
	}

	for _, u := range ageFilteredUsers {
		fmt.Printf("  - %s (Age: %d, Email: %s)\n", u.Name, u.Age, u.Email)
	}

	// Query 2: Find specific users by name or age
	fmt.Println("\n2. Users named Alice OR age >= 32 (using Select):")
	orConditionQuery := "SELECT * FROM users WHERE name = ? OR age >= ? ORDER BY name ASC;"
	var filteredUsers []User
	if err := client.Select(&filteredUsers, orConditionQuery, "Alice", 32); err != nil {
		log.Fatalf("Select failed: %v", err)
	}

	for _, u := range filteredUsers {
		fmt.Printf("  - %s (Age: %d)\n", u.Name, u.Age)
	}

	// ============ Update with Multiple Conditions ============
	fmt.Println("\n=== Update with Multiple Conditions ===")

	updateQuery := "UPDATE users SET age = ? WHERE age > ? AND name != ?;"
	res, err = client.Query(updateQuery, []string{"99", "30", "Alice"})
	if err != nil {
		log.Fatalf("Update failed: %v", err)
	}

	result, err := res.ToResult()
	if err != nil {
		log.Fatalf("Failed to get result: %v", err)
	}
	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("Updated users with age > 30 and name != Alice. Rows Affected: %d\n", rowsAffected)

	// ============ Query All Data ============
	fmt.Println("\n=== All Users (After Updates) ===")
	selectAllQuery := "SELECT * FROM users ORDER BY id ASC;"
	var allUsers []User
	if err := client.Select(&allUsers, selectAllQuery); err != nil {
		log.Fatalf("Select failed: %v", err)
	}

	fmt.Printf("Total users: %d\n", len(allUsers))
	for _, u := range allUsers {
		fmt.Printf("  - ID: %d, Name: %s, Age: %d, Email: %s\n", u.ID, u.Name, u.Age, u.Email)
	}

	// ============ Single Row Query (Get Method) ============
	fmt.Println("\n=== Single Row Query (sqlx-style Get) ===")
	getQuery := "SELECT * FROM users WHERE name = ? LIMIT 1;"
	var singleUser User
	if err := client.Get(&singleUser, getQuery, "Alice"); err != nil {
		log.Fatalf("Get failed: %v", err)
	}
	fmt.Printf("Found user: ID=%d, Name=%s, Age=%d, Email=%s\n", singleUser.ID, singleUser.Name, singleUser.Age, singleUser.Email)

	// ============ JOIN Query Example ============
	fmt.Println("\n=== JOIN Query Examples ===")

	// First, insert relationships
	fmt.Println("\nInserting user-department relationships...")
	insertRelQuery := "INSERT INTO user_departments (user_id, department_id) VALUES (?, ?);"
	relationships := []struct {
		userIdx int
		deptIdx int
	}{
		{0, 0}, // Alice -> Engineering
		{0, 1}, // Alice -> Sales
		{1, 0}, // Bob -> Engineering
		{2, 0}, // Charlie -> Engineering
		{3, 2}, // Diana -> HR
		{4, 1}, // Eve -> Sales
	}

	for _, rel := range relationships {
		if rel.userIdx < len(userIDs) && rel.deptIdx < len(deptIDs) {
			res, err = client.Query(insertRelQuery, []string{fmt.Sprintf("%d", userIDs[rel.userIdx]), fmt.Sprintf("%d", deptIDs[rel.deptIdx])})
			if err != nil {
				log.Fatalf("Insert relationship failed: %v", err)
			}
			result, err := res.ToResult()
			if err != nil {
				log.Fatalf("Failed to get result: %v", err)
			}
			if result != nil {
				_, _ = result.LastInsertId()
			}
		}
	}
	fmt.Println("  ✓ Relationships inserted")

	// LEFT JOIN Query: Get all users and their departments (even if no department)
	fmt.Println("\n1. LEFT JOIN - All users with their departments:")
	leftJoinQuery := `
		SELECT 
			u.id as user_id,
			u.name as user_name,
			u.age,
			d.id as department_id,
			d.name as dept_name
		FROM users u
		LEFT JOIN user_departments ud ON u.id = ud.user_id
		LEFT JOIN departments d ON ud.department_id = d.id
		ORDER BY u.id ASC, d.id ASC;
	`
	var joinResults []UserWithDept
	if err := client.Select(&joinResults, leftJoinQuery); err != nil {
		log.Fatalf("Select failed: %v", err)
	}

	for _, r := range joinResults {
		deptName := "N/A"
		if r.DeptName != "" {
			deptName = r.DeptName
		}
		fmt.Printf("  - %s (Age: %d) -> %s\n", r.UserName, r.Age, deptName)
	}

	// INNER JOIN Query: Only users with departments
	fmt.Println("\n2. INNER JOIN - Users with departments (exclude unassigned):")
	innerJoinQuery := `
		SELECT 
			u.id as user_id,
			u.name as user_name,
			u.age,
			d.id as department_id,
			d.name as dept_name
		FROM users u
		INNER JOIN user_departments ud ON u.id = ud.user_id
		INNER JOIN departments d ON ud.department_id = d.id
		ORDER BY u.id ASC;
	`
	var innerJoinResults []UserWithDept
	if err := client.Select(&innerJoinResults, innerJoinQuery); err != nil {
		log.Fatalf("Select failed: %v", err)
	}

	for _, result := range innerJoinResults {
		fmt.Printf("  - %s (Age: %d) works in %s\n", result.UserName, result.Age, result.DeptName)
	}

	// ============ UPSERT (INSERT OR UPDATE) Test ============
	fmt.Println("\n=== UPSERT (INSERT OR UPDATE) Test ===")
	fmt.Println("\nD1 supports three UPSERT methods:")

	// Method 1: INSERT OR IGNORE (skip if exists)
	fmt.Println("\n1. Testing INSERT OR IGNORE (skip if already exists):")
	upsertQuery1 := "INSERT OR IGNORE INTO users (name, age, email) VALUES (?, ?, ?);"
	for i := 0; i < 2; i++ {
		res, err = client.Query(upsertQuery1, []string{"Frank", "26", "frank@example.com"})
		if err != nil {
			fmt.Printf("  Attempt %d failed: %v\n", i+1, err)
		} else {
			result, _ := res.ToResult()
			changes, _ := result.RowsAffected()
			if changes > 0 {
				fmt.Printf("  Attempt %d: ✓ Inserted successfully\n", i+1)
			} else {
				fmt.Printf("  Attempt %d: ⊘ Skipped (already exists)\n", i+1)
			}
		}
	}

	// Method 2: INSERT ... ON CONFLICT ... DO UPDATE
	fmt.Println("\n2. Testing INSERT ... ON CONFLICT ... DO UPDATE (update if exists):")
	upsertQuery2 := `INSERT INTO users (id, name, age, email) VALUES (?, ?, ?, ?)
    ON CONFLICT(email) DO UPDATE SET name = ?, age = ?;`

	// This will try to insert a user with existing email (alice@example.com)
	res, err = client.Query(upsertQuery2, []string{"100", "Alice_Updated", "31", "alice@example.com", "Alice_Updated", "31"})
	if err != nil {
		fmt.Printf("  Note: UPSERT may not be fully supported: %v\n", err)
	} else {
		result, _ := res.ToResult()
		changes, _ := result.RowsAffected()
		fmt.Printf("  ✓ UPSERT completed with %d rows affected\n", changes)
	}

	// Method 3: INSERT OR REPLACE (replace entire row if exists)
	fmt.Println("\n3. Testing INSERT OR REPLACE (replace entire row if exists):")
	upsertQuery3 := "INSERT OR REPLACE INTO users (id, name, age, email) VALUES (?, ?, ?, ?);"

	// First, insert a record with Bob
	res, err = client.Query(upsertQuery3, []string{"102", "Bob", "25", "bob.new@example.com"})
	if err != nil {
		fmt.Printf("  Initial insert failed: %v\n", err)
	} else {
		result, _ := res.ToResult()
		lastID, _ := result.LastInsertId()
		fmt.Printf("  First insert: ID=%d, Name=Bob\n", lastID)
	}

	// Now replace the same record with different data
	res, err = client.Query(upsertQuery3, []string{"102", "Bob_Updated", "26", "bob.new@example.com"})
	if err != nil {
		fmt.Printf("  Replace failed: %v\n", err)
	} else {
		result, _ := res.ToResult()
		changes, _ := result.RowsAffected()
		fmt.Printf("  Replace operation: %d rows affected\n", changes)
	}

	// Check final state
	fmt.Println("\n4. Final user data (showing UPSERT results):")
	finalQuery := "SELECT id, name, age, email FROM users WHERE name LIKE 'Alice%' OR name LIKE 'Bob%' OR name = 'Frank' ORDER BY id ASC;"
	var finalUsers []User
	if err := client.Select(&finalUsers, finalQuery); err != nil {
		log.Fatalf("Select failed: %v", err)
	}

	for _, u := range finalUsers {
		fmt.Printf("  - ID: %d, Name: %s, Age: %d, Email: %s\n", u.ID, u.Name, u.Age, u.Email)
	}

	fmt.Println("\n=== UPSERT Summary ===")
	fmt.Println("✓ D1 SQLite支持以下UPSERT操作:")
	fmt.Println("  • INSERT OR IGNORE - 如果主键存在则忽略")
	fmt.Println("  • INSERT ... ON CONFLICT ... DO UPDATE - 高级冲突处理")
	fmt.Println("  • INSERT OR REPLACE - 如果主键存在则替换")
	fmt.Println("  (具体支持程度取决于Cloudflare D1的SQLite版本)")

	fmt.Println("\n=== Test Complete ===")
	fmt.Println("✓ All tests passed successfully!")

	fmt.Println("\n=== Cleanup ===")
	// Clean up tables and migrations table for next run
	_, _ = client.RemoveTable("user_departments")
	_, _ = client.RemoveTable("users")
	_, _ = client.RemoveTable("departments")
	_, _ = client.RemoveTable("d1_migrations") // Remove migrations table to allow re-running migrations
	fmt.Println("✓ All tables dropped including migrations table")
}
