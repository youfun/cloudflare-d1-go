# Cloudflare D1 Go Client ☁️

[中文版本](README.zh.md) | English

> 🔄 **Enhanced Fork Notice**
> 
> This is an enhanced version based on [ashayas/cloudflare-d1-go](https://github.com/ashayas/cloudflare-d1-go).
> 
> **Key Improvements:**
> - ✨ **sqlx-style API** - Clean methods like `Select()`, `Get()`, `Exec()` with automatic type conversion
> - ✨ **Automatic Parameter Conversion** - Pass int, bool, time.Time directly without []string conversion
> - ✨ **ConnectionPool with Caching** - Reduce API calls by 99% with intelligent database connection pooling
> - ✨ Enhanced data type handling (supports array-format rows from D1 API)
> - ✨ Advanced query support (JOIN queries, complex WHERE conditions)
> - ✨ UPSERT operations with full SQLite conflict resolution support
> - ✨ Improved error handling and data validation
> - ✨ StructScan with proper NULL value handling
> - ✨ Comprehensive examples for real-world scenarios
>
> The core HTTP request layer (`utils/request.go`) is based on the original project.





<p align="center">
<a href="https://pkg.go.dev/github.com/youfun/cloudflare-d1-go"><img src="https://pkg.go.dev/badge/github.com/youfun/cloudflare-d1-go.svg" alt="Go Reference"></a>
<img src="https://img.shields.io/github/go-mod/go-version/youfun/cloudflare-d1-go" alt="Go Version">
<img src="https://img.shields.io/badge/license-MIT-blue" alt="MIT License">
</p>

## Installation 📦

```bash
go get github.com/youfun/cloudflare-d1-go
```

## Quick Start 🚀

### ⭐ New: sqlx-Style API (Recommended)

The simplest way to query and scan results, just like sqlx:

#### Single Row Query with `Get()`
```go
type User struct {
    ID    int    `db:"id"`
    Name  string `db:"name"`
    Age   int    `db:"age"`
    Email string `db:"email"`
}

var user User
err := client.Get(
    &user,
    "SELECT * FROM users WHERE name = ?",
    "Alice",  // Direct parameter, no []string needed
)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("User: %s (Age: %d)\n", user.Name, user.Age)
```

#### Multiple Row Query with `Select()`
```go
var users []User
err := client.Select(
    &users,
    "SELECT * FROM users WHERE age > ? ORDER BY age ASC",
    25,  // Direct int parameter
)
if err != nil {
    log.Fatal(err)
}
for _, u := range users {
    fmt.Printf("%s (Age: %d)\n", u.Name, u.Age)
}
```

#### Execute Updates/Inserts with `Exec()`
```go
// Execute UPDATE and get rows affected
rowsAffected, err := client.Exec(
    "UPDATE users SET age = ? WHERE id = ?",
    30,   // Direct int
    123,  // Direct int
)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Updated %d rows\n", rowsAffected)
```

#### Multiple Parameter Types (Automatic Conversion)
```go
// String parameter
err := client.Get(&user, "SELECT * FROM users WHERE name = ?", "Alice")

// Integer parameter
err := client.Select(&users, "SELECT * FROM users WHERE age > ?", 25)

// Boolean parameter
err := client.Select(&results, "SELECT * FROM users WHERE active = ?", true)

// Time parameter (automatically formatted)
startTime := time.Now().Add(-24 * time.Hour)
err := client.Select(&users, "SELECT * FROM users WHERE created_at > ?", startTime)

// Mixed parameters
err := client.Select(
    &users,
    "SELECT * FROM users WHERE age > ? AND active = ? AND created_at > ?",
    25,         // int
    true,       // bool
    startTime,  // time.Time
)
```

**Benefits:**
- ✅ Clean, intuitive API similar to sqlx
- ✅ Automatic type conversion (int, bool, time.Time, string, etc.)
- ✅ One-line queries - no manual row iteration
- ✅ Automatic struct mapping using `db` tags
- ✅ Clean error handling
- ✅ Variadic parameters - just pass values directly

---

### ConnectionPool with sqlx-Style Methods (Recommended for Production)

The `ConnectionPool` provides sqlx-like methods with automatic caching:

```go
pool := cloudflare_d1_go.NewConnectionPool(accountID, apiToken)
pool.SetCacheAge(1 * time.Hour)

err := pool.Connect("database_name")
if err != nil {
    log.Fatal(err)
}

// Now use sqlx-style methods
var users []User
err = pool.Select(&users, "SELECT * FROM users WHERE age > ?", 25)

var user User
err = pool.Get(&user, "SELECT * FROM users WHERE id = ?", 123)

rowsAffected, err := pool.Exec("UPDATE users SET age = ? WHERE id = ?", 30, 123)
```

---

### Classic API (Still Supported)

#### Method 1: Direct Client (Basic)

#### Initialize the client 🔑

```go
client := cloudflare_d1_go.NewClient("account_id", "api_token")
```

#### Connect to a database 📁

```go
client.ConnectDB("database_name")
```

#### Query the database 🔍

```go
// Execute a SQL query with optional parameters
// query: SQL query string
// params: Array of parameter values to bind to the query (use ? placeholders in query)
client.Query("SELECT * FROM users WHERE age > ?", []string{"18"})
```

#### Example with parameters:
```go
// Find users in a specific city
client.Query("SELECT * FROM users WHERE city = ?", []string{"San Francisco"})

// Find products in a price range
client.Query("SELECT * FROM products WHERE price >= ? AND price <= ?", []string{"10.00", "50.00"})
```

#### Create a table 📄

```go
client.CreateTable("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, age INTEGER)")
```

#### Remove a table 🗑️

```go
client.RemoveTable("users")
```

#### Query specific database (Method 2) 🔀

```go
client := cloudflare_d1_go.NewClient("account_id", "api_token")
client.QueryDB(databaseID, "SELECT * FROM users", nil)
```

---

### Method 2: ConnectionPool (Recommended - Similar to sqlx.DB)

The `ConnectionPool` provides a sqlx-like experience with automatic caching and connection persistence. This is **recommended for production use** as it reduces API calls by up to 99%.

#### Initialize the connection pool

```go
pool := cloudflare_d1_go.NewConnectionPool("account_id", "api_token")

// Optional: Set cache age (default is 24 hours)
pool.SetCacheAge(1 * time.Hour)
```

#### Connect to database (with automatic caching)

```go
// First call: Makes an API request to fetch database ID (661ms)
// Subsequent calls: Instantly returns from cache (0ms)
err := pool.Connect("database_name")
if err != nil {
    log.Fatalf("Connection failed: %v", err)
}
```

#### Execute queries (just like sqlx)

```go
// Simple query
res, err := pool.Query("SELECT * FROM users", nil)

// Query with parameters
res, err := pool.Query("SELECT * FROM users WHERE age > ? AND age < ?", []string{"25", "35"})

// Insert with result info
res, err := pool.Query("INSERT INTO users (name, age) VALUES (?, ?)", []string{"Alice", "30"})
if err == nil {
    result, _ := res.ToResult()
    lastID, _ := result.LastInsertId()
    fmt.Printf("Inserted with ID: %d\n", lastID)
}

// Update with affected rows count
res, err := pool.Query("UPDATE users SET age = ? WHERE age > ?", []string{"99", "50"})
if err == nil {
    result, _ := res.ToResult()
    affected, _ := result.RowsAffected()
    fmt.Printf("Updated %d rows\n", affected)
}
```

#### Process query results (like sqlx)

```go
// Define struct with db tags
type User struct {
    ID    int    `db:"id"`
    Name  string `db:"name"`
    Age   int    `db:"age"`
    Email string `db:"email"`
}

// Query and scan results
res, err := pool.Query("SELECT id, name, age, email FROM users ORDER BY age ASC", nil)
if err != nil {
    log.Fatalf("Query failed: %v", err)
}

rows, _ := res.ToRows()
defer rows.Close()

for rows.Next() {
    var user User
    rows.StructScan(&user)
    fmt.Printf("%s (Age: %d)\n", user.Name, user.Age)
}
```

#### Cache management

```go
// Check if database is cached
if pool.IsCached("database_name") {
    fmt.Println("Using cached connection")
} else {
    fmt.Println("Making fresh API call")
}

// View cache info
info := pool.GetCacheInfo("database_name")
fmt.Printf("Database ID: %s, Cached at: %v\n", info.DatabaseID, info.CachedAt)

// List all cached databases
dbList := pool.ListCachedDatabases()
fmt.Printf("Cached databases: %v\n", dbList)

// Clear specific cache
pool.ClearCache("database_name")

// Clear all cache
pool.ClearAllCache()
```

#### Multiple databases

```go
pool := cloudflare_d1_go.NewConnectionPool(accountID, apiToken)

// Connect to multiple databases
pool.Connect("users_db")
pool.Connect("products_db")

// Query specific database (without switching current)
res, err := pool.QueryDB("users_db", "SELECT * FROM users", nil)

// Or switch current database
pool.Connect("products_db")  // Make it current
res, err := pool.Query("SELECT * FROM products", nil)  // Uses products_db
```

#### Performance Comparison

```
✅ First connection (API call):     661.9ms
✅ Subsequent connections (cached):  0ms
✅ Savings: 99% reduction in API calls!
```

**Recommended settings for different use cases:**

```go
// Web service (short-lived connections)
pool.SetCacheAge(1 * time.Hour)
pool.SetAutoReconnect(true)

// Long-running batch jobs
pool.SetCacheAge(24 * time.Hour)

// Development/testing
pool.SetCacheAge(5 * time.Minute)  // Fresh data frequently
```

## Advanced Features 🔧

### UPSERT Operations (Insert or Update)

D1 supports SQLite-based UPSERT operations similar to PostgreSQL. This is useful for data synchronization and deduplication scenarios.

#### Scenario 1: User Account Synchronization

When syncing user accounts from an external source, you want to update existing users or insert new ones:

```go
type User struct {
    ID    int    `db:"id"`
    Name  string `db:"name"`
    Email string `db:"email"`
    Age   int    `db:"age"`
}

// Upsert query - updates if email exists, inserts if not
upsertQuery := `
    INSERT INTO users (id, name, email, age) 
    VALUES (?, ?, ?, ?)
    ON CONFLICT(email) DO UPDATE 
    SET name = excluded.name, age = excluded.age;
`

// Sync user data
user := User{ID: 100, Name: "Alice", Email: "alice@example.com", Age: 30}
res, err := client.Query(upsertQuery, []string{
    fmt.Sprintf("%d", user.ID),
    user.Name,
    user.Email,
    fmt.Sprintf("%d", user.Age),
})

if err != nil {
    log.Fatalf("Upsert failed: %v", err)
}

result, _ := res.ToResult()
rowsAffected, _ := result.RowsAffected()
fmt.Printf("Synced user. Rows affected: %d\n", rowsAffected)
```

**Benefits:**
- ✅ No need to check if user exists first
- ✅ Atomic operation (no race conditions)
- ✅ Efficient single-query synchronization
- ✅ Automatic conflict resolution

#### Scenario 2: Data Deduplication (Skip Duplicates)

When importing data from multiple sources, you want to skip duplicate records:

```go
// Insert query with duplicate skip - inserts if email doesn't exist, ignores if it does
insertOrIgnoreQuery := "INSERT OR IGNORE INTO users (name, email, age) VALUES (?, ?, ?);"

// Try to insert duplicate records
emails := []string{"bob@example.com", "charlie@example.com", "bob@example.com"}
names := []string{"Bob", "Charlie", "Bob"}
ages := []string{"25", "35", "25"}

for i := 0; i < len(emails); i++ {
    res, err := client.Query(insertOrIgnoreQuery, []string{names[i], emails[i], ages[i]})
    if err != nil {
        log.Fatalf("Insert failed: %v", err)
    }
    
    result, _ := res.ToResult()
    rowsAffected, _ := result.RowsAffected()
    
    if rowsAffected > 0 {
        fmt.Printf("✓ Inserted user %s\n", names[i])
    } else {
        fmt.Printf("⊘ Skipped duplicate user %s\n", names[i])
    }
}
```

**Benefits:**
- ✅ Automatic duplicate detection
- ✅ No errors on duplicate inserts
- ✅ Clean batch import process
- ✅ Clear feedback on inserted vs skipped records

### UPSERT Syntax Comparison

D1 supports multiple UPSERT approaches. **All three methods have been verified and tested:**

```sql
-- Method 1: INSERT OR IGNORE (skips duplicate) ✓ Tested
INSERT OR IGNORE INTO users (id, name, email, age) 
VALUES (?, ?, ?, ?);

-- Method 2: INSERT ... ON CONFLICT ... DO UPDATE (selective update) ✓ Tested
INSERT INTO users (id, name, email, age) 
VALUES (?, ?, ?, ?)
ON CONFLICT(email) DO UPDATE 
SET name = excluded.name, age = excluded.age;

-- Method 3: INSERT OR REPLACE (replaces entire row) ✓ Tested
INSERT OR REPLACE INTO users (id, name, email, age) 
VALUES (?, ?, ?, ?);
```

> **Test Results:** All three UPSERT methods have been successfully tested with Cloudflare D1:
> - Method 1: Correctly skips duplicate inserts (returns 0 rows affected)
> - Method 2: Correctly updates existing records based on conflict column (returns 1 row affected)
> - Method 3: Correctly replaces entire rows with matching primary key (returns 1 row affected)

## Method Reference 📚

### Database Management
- `NewClient(accountID, apiToken string) *Client` - Creates a new D1 client
- `ListDB() (*APIResponse, error)` - Lists all databases in the account
- `CreateDB(name string) (*APIResponse, error)` - Creates a new database
- `DeleteDB(databaseID string) (*APIResponse, error)` - Deletes a database
- `ConnectDB(name string) error` - Connects to a database by name for subsequent operations

### Table Operations
- `CreateTable(createQuery string) (*APIResponse, error)` - Creates a table in the connected database
- `RemoveTable(tableName string) (*APIResponse, error)` - Removes a table from the connected database
- `CreateTableWithID(databaseID, createQuery string) (*APIResponse, error)` - Creates a table in a specific database
- `RemoveTableWithID(databaseID, tableName string) (*APIResponse, error)` - Removes a table from a specific database

### Query Execution
- `Query(query string, params []string) (*APIResponse, error)` - Executes a query on the connected database
  - Supports SELECT, INSERT, UPDATE, DELETE and all SQL operations
  - Parameters passed via array, corresponding to `?` placeholders in SQL
  - Example: `client.Query("INSERT INTO users (name, age) VALUES (?, ?)", []string{"Alice", "30"})`
  - Example: `client.Query("SELECT * FROM users WHERE age > ? AND age < ?", []string{"20", "40"})`
- `QueryDB(databaseID string, query string, params []string) (*APIResponse, error)` - Executes a query on a specific database
  - Same functionality as above, but for disconnected specific databases

#### sqlx-Style Convenience Methods (Recommended) ✨

**New Recommended Methods:**
- `Select(dest interface{}, query string, args ...interface{}) error` - Query multiple rows and scan into slice (sqlx-style)
  - `dest` must be a pointer to a slice, e.g., `&[]User{}`
  - `args` are variadic parameters (int, string, bool, time.Time, etc. - automatic conversion)
  - Returns empty slice if no rows found
  - Example: `client.Select(&users, "SELECT * FROM users WHERE age > ?", 25)`
  - Example: `client.Select(&users, "SELECT * FROM users WHERE age > ? AND active = ?", 25, true)`

- `Get(dest interface{}, query string, args ...interface{}) error` - Query a single row and scan into struct (sqlx-style)
  - `dest` must be a pointer to a struct, e.g., `&user`
  - `args` are variadic parameters (int, string, bool, time.Time, etc. - automatic conversion)
  - Returns error if no rows found
  - Example: `client.Get(&user, "SELECT * FROM users WHERE id = ?", 123)`
  - Example: `client.Get(&user, "SELECT * FROM users WHERE name = ?", "Alice")`

- `Exec(query string, args ...interface{}) (int64, error)` - Execute INSERT/UPDATE/DELETE and get rows affected (sqlx-style)
  - Returns the number of rows affected
  - `args` are variadic parameters (int, string, bool, time.Time, etc. - automatic conversion)
  - Example: `rowsAffected, err := client.Exec("UPDATE users SET age = ? WHERE id = ?", 30, 123)`
  - Example: `rowsAffected, err := client.Exec("INSERT INTO users (name, age) VALUES (?, ?)", "Alice", 30)`

**Parameter Type Support:**
All three methods support automatic parameter type conversion:
- String: `"Alice"` → `"Alice"`
- Integer types: `25` → `"25"` (int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64)
- Float types: `25.5` → `"25.5"` (float32, float64)
- Boolean: `true` → `"1"`, `false` → `"0"`
- time.Time: `time.Now()` → `"2006-01-02 15:04:05"` (formatted timestamp)
- Complex types: Automatically JSON marshaled
- Nil: `nil` → `""` (empty string for NULL values)

---

#### Legacy Convenience Methods (Deprecated - Use Select/Get/Exec instead)
- `QueryOne(query string, dest interface{}, args ...interface{}) error` - ⚠️ DEPRECATED: Use Get() instead
- `QueryAll(query string, dest interface{}, args ...interface{}) error` - ⚠️ DEPRECATED: Use Select() instead

#### Common SQL Operation Examples
**Single Insert:**
```go
res, err := client.Query(
    "INSERT INTO users (name, age, email) VALUES (?, ?, ?)",
    []string{"Alice", "30", "alice@example.com"},
)
```

**Single Delete:**
```go
res, err := client.Query(
    "DELETE FROM users WHERE id = ?",
    []string{"1"},
)
```

**Multiple Conditions Query (AND):**
```go
res, err := client.Query(
    "SELECT * FROM users WHERE age > ? AND age < ? ORDER BY age ASC",
    []string{"25", "35"},
)
```

**Multiple Conditions Query (OR):**
```go
res, err := client.Query(
    "SELECT * FROM users WHERE name = ? OR age >= ? ORDER BY name ASC",
    []string{"Alice", "30"},
)
```

**Multiple Conditions Update:**
```go
res, err := client.Query(
    "UPDATE users SET age = ? WHERE age > ? AND name != ?",
    []string{"31", "30", "Alice"},
)
result, _ := res.ToResult()
rowsAffected, _ := result.RowsAffected()
```

**Multiple Conditions Delete:**
```go
res, err := client.Query(
    "DELETE FROM users WHERE age < ? AND status = ?",
    []string{"18", "inactive"},
)
```

### Response Methods
- `ToRows() (*Rows, error)` - Converts SELECT query response to Rows for iteration
- `ToResult() (*Result, error)` - Converts INSERT/UPDATE/DELETE response to Result for metadata
- `Get(dest interface{}) error` - Scans first row into struct (sqlx-style)
  - `dest` must be a pointer to a struct
  - Returns error if no rows found
- `StructScanAll(dest interface{}) error` - Scans all rows into slice of structs (sqlx-style)
  - `dest` must be a pointer to a slice
  - Returns empty slice if no rows found

### Rows Methods (for SELECT queries)
- `Next() bool` - Prepares the next result row for reading
- `Scan(dest ...interface{}) error` - Copies columns in the current row to destination variables
- `StructScan(dest interface{}) error` - Scans current row into a struct using `db` tags
- `StructScanAll(dest interface{}) error` - Scans all remaining rows into a slice of structs (sqlx-style)
  - `dest` must be a pointer to a slice
  - Useful when you have existing Rows object
- `Columns() ([]string, error)` - Returns the column names
- `Close() error` - Closes the Rows

### Result Methods (for INSERT/UPDATE/DELETE)
- `LastInsertId() (int64, error)` - Returns the last inserted row ID
- `RowsAffected() (int64, error)` - Returns the number of rows affected

## Configuration 🔧

Set environment variables or use `.env` file:

```bash
CLOUDFLARE_ACCOUNT_ID=your_account_id
CLOUDFLARE_API_TOKEN=your_api_token
CLOUDFLARE_DB_NAME=your_database_name
```

See `example/.env.example` for detailed instructions.

## Examples 📖

Check the `example/` directory for comprehensive examples:

### `example/main.go` - Complete Feature Showcase
Demonstrates:
- ✅ **sqlx-style Select()** - Multi-row queries with automatic type conversion
- ✅ **sqlx-style Get()** - Single-row queries
- ✅ **sqlx-style Exec()** - INSERT/UPDATE/DELETE with rows affected
- ✅ Batch insert operations
- ✅ Multiple WHERE conditions (AND, OR)
- ✅ JOIN queries (LEFT JOIN, INNER JOIN)
- ✅ UPSERT operations (INSERT OR IGNORE, INSERT ON CONFLICT, INSERT OR REPLACE)
- ✅ Data synchronization scenarios
- ✅ Migrations with automatic schema management

Run it:
```bash
cd example
cp .env.example .env
# Edit .env with your Cloudflare credentials
go run main.go
```

### `example/pool_demo.go` - ConnectionPool with sqlx-Style API
Demonstrates:
- ✅ **ConnectionPool.Select()** - Query multiple rows
- ✅ **ConnectionPool.Get()** - Query single row
- ✅ **ConnectionPool.Exec()** - Execute updates and deletes
- ✅ Parameter type conversion (string, int, bool, time.Time)
- ✅ Cache management
- ✅ Multiple database handling

Run it:
```bash
cd example
go run pool_demo.go
```

## Migrations 🗄️

The migrations package provides a robust way to manage database schema changes.

### Features

- **Version Tracking**: Automatically tracks applied migrations in a `d1_migrations` table
- **Multiple Sources**: Supports multiple migration sources:
  - `FileMigrationSource`: Load from local directory (e.g., `migrations/`)
  - `EmbedFileSystemMigrationSource`: Use `embed.FS` for single-binary deployments
  - `MemoryMigrationSource`: In-memory migration list
- **SQL Format**: Compatible with sql-migrate format (`-- +migrate Up`, `-- +migrate Down`)
- **D1 Integration**: Works directly with cloudflare-d1-go client

### Create Migration Files

Create SQL files in a directory (e.g., `migrations/1_init.sql`):

```sql
-- +migrate Up
CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, age INTEGER);
CREATE TABLE posts (id INTEGER PRIMARY KEY, user_id INTEGER, title TEXT);

-- +migrate Down
DROP TABLE posts;
DROP TABLE users;
```

### Apply Migrations in Code

```go
import (
    "github.com/youfun/cloudflare-d1-go/migrations"
)

func main() {
    // Initialize client and connect to database
    client := cloudflare_d1_go.NewClient(accountID, apiToken)
    err := client.ConnectDB("database_name")
    if err != nil {
        log.Fatal(err)
    }

    // Define migration source
    source := &migrations.FileMigrationSource{
        Dir: "migrations",
    }

    // Apply migrations
    n, err := migrations.Exec(client, source, migrations.Up)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Applied %d migrations!\n", n)
}
```

### Using Embedded Migrations

For single-binary deployments:

```go
import (
    "embed"
    "github.com/youfun/cloudflare-d1-go/migrations"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

func main() {
    // ... client setup ...
    
    source := &migrations.EmbedFileSystemMigrationSource{
        FS: migrationFS,
        Dir: "migrations",
    }
    
    n, err := migrations.Exec(client, source, migrations.Up)
    if err != nil {
        log.Fatal(err)
    }
}
```

### Roll Back Migrations

```go
// Rollback last migration
n, err := migrations.Exec(client, source, migrations.Down)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Rolled back %d migrations\n", n)
```

## Testing 🧪

```bash
go test -v
```

## Known Limitations ⚠️

### Transaction Support

**D1 REST API does not support atomic transactions.** This is a platform-level limitation:

- ✅ **What Works:**
  - Single SQL statements (SELECT, INSERT, UPDATE, DELETE)
  - UPSERT operations using SQLite syntax (INSERT OR IGNORE, INSERT OR REPLACE, ON CONFLICT)
  - Complex queries (JOINs, subqueries, aggregations)

- ❌ **What's Not Supported:**
  - Multi-statement transactions (BEGIN/COMMIT/ROLLBACK)
  - Atomic operations across multiple queries
  - Sequential consistency guarantees across multiple requests

**Why?** The Cloudflare D1 REST API processes each query as an independent request. Transaction support is only available in the **Workers Binding API** (JavaScript/TypeScript), which can batch multiple statements together in a single request.

**Workarounds:**
1. **Use UPSERT Operations** - For atomic insert-or-update scenarios:
   ```go
   // This is atomic within a single query
   INSERT INTO users (id, name, email) VALUES (?, ?, ?)
   ON CONFLICT(email) DO UPDATE SET name = excluded.name
   ```

2. **Design Idempotent Operations** - Ensure your application logic can safely retry failed queries

3. **Use Workers Binding API** - If you need transactions, use Cloudflare Workers with D1 bindings (JavaScript/TypeScript)

4. **Application-Level Coordination** - Implement optimistic locking or version columns for concurrent updates

## TODO 📋

- Better error handling 🛡️
- More comprehensive test coverage 🧪

## Contributing 🤝

Contributions are welcome! Please feel free to submit a Pull Request.

## License 📄

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Support 💪

If you encounter any issues or have questions, please file an issue on GitHub.

## Acknowledgments 🙏

- **HTTP Request Layer**: Based on [ashayas/cloudflare-d1-go](https://github.com/ashayas/cloudflare-d1-go)
- **Migrations Package**: Based on [github.com/rubenv/sql-migrate](https://github.com/rubenv/sql-migrate)
