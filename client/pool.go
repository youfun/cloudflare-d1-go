package cloudflared1

import (
	"fmt"
	"sync"
	"time"

	"github.com/youfun/cloudflare-d1-go/utils"
)

// ConnectionInfo holds database connection metadata
type ConnectionInfo struct {
	DatabaseID string
	Name       string
	CachedAt   time.Time
}

// ConnectionPool manages database connections with caching and persistence
// Similar to sqlx.DB but for Cloudflare D1
type ConnectionPool struct {
	accountID       string
	apiToken        string
	connections     map[string]*ConnectionInfo
	currentDB       string
	mu              sync.RWMutex
	maxCacheAge     time.Duration
	autoReconnect   bool
	lastHealthCheck time.Time
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool(accountID, apiToken string) *ConnectionPool {
	if accountID == "" || apiToken == "" {
		return nil
	}
	return &ConnectionPool{
		accountID:     accountID,
		apiToken:      apiToken,
		connections:   make(map[string]*ConnectionInfo),
		maxCacheAge:   24 * time.Hour, // Cache for 24 hours by default
		autoReconnect: true,
	}
}

// Connect connects to a database by name, with automatic caching
// If cached, returns immediately without API call
// Like sqlx: pool.Connect("database_name")
func (p *ConnectionPool) Connect(dbName string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if already connected and cache is valid
	if connInfo, exists := p.connections[dbName]; exists {
		if time.Since(connInfo.CachedAt) < p.maxCacheAge {
			p.currentDB = dbName
			return nil // Return from cache
		}
	}

	// Cache miss or expired, fetch from API
	client := &Client{
		AccountID: p.accountID,
		APIToken:  p.apiToken,
	}

	if err := client.ConnectDB(dbName); err != nil {
		return fmt.Errorf("failed to connect to database %s: %w", dbName, err)
	}

	// Cache the connection info
	p.connections[dbName] = &ConnectionInfo{
		DatabaseID: client.DatabaseID,
		Name:       dbName,
		CachedAt:   time.Now(),
	}

	p.currentDB = dbName
	return nil
}

// ConnectWithID connects directly using database ID
// Useful when you already know the database ID
func (p *ConnectionPool) ConnectWithID(dbName, databaseID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.connections[dbName] = &ConnectionInfo{
		DatabaseID: databaseID,
		Name:       dbName,
		CachedAt:   time.Now(),
	}

	p.currentDB = dbName
	return nil
}

// Query executes a query on the currently connected database
// Like sqlx: result := pool.Query("SELECT * FROM users")
func (p *ConnectionPool) Query(query string, params []string) (*utils.APIResponse, error) {
	p.mu.RLock()
	connInfo, exists := p.connections[p.currentDB]
	p.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no database connected, call Connect first")
	}

	client := &Client{
		AccountID:  p.accountID,
		APIToken:   p.apiToken,
		DatabaseID: connInfo.DatabaseID,
	}

	return client.Query(query, params)
}

// Select executes a query and scans all results into a slice, similar to sqlx.Select
// Like sqlx: pool.Select(&users, "SELECT * FROM users WHERE age > ?", 25)
func (p *ConnectionPool) Select(dest interface{}, query string, args ...interface{}) error {
	p.mu.RLock()
	connInfo, exists := p.connections[p.currentDB]
	p.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no database connected, call Connect first")
	}

	client := &Client{
		AccountID:  p.accountID,
		APIToken:   p.apiToken,
		DatabaseID: connInfo.DatabaseID,
	}

	return client.Select(dest, query, args...)
}

// Get executes a query and scans the first result into a struct, similar to sqlx.Get
// Like sqlx: pool.Get(&user, "SELECT * FROM users WHERE id = ?", 123)
func (p *ConnectionPool) Get(dest interface{}, query string, args ...interface{}) error {
	p.mu.RLock()
	connInfo, exists := p.connections[p.currentDB]
	p.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no database connected, call Connect first")
	}

	client := &Client{
		AccountID:  p.accountID,
		APIToken:   p.apiToken,
		DatabaseID: connInfo.DatabaseID,
	}

	return client.Get(dest, query, args...)
}

// Exec executes a query and returns the number of rows affected, similar to sqlx.Exec
// Like sqlx: rowsAffected, err := pool.Exec("UPDATE users SET age = ? WHERE id = ?", 30, 123)
func (p *ConnectionPool) Exec(query string, args ...interface{}) (int64, error) {
	p.mu.RLock()
	connInfo, exists := p.connections[p.currentDB]
	p.mu.RUnlock()

	if !exists {
		return 0, fmt.Errorf("no database connected, call Connect first")
	}

	client := &Client{
		AccountID:  p.accountID,
		APIToken:   p.apiToken,
		DatabaseID: connInfo.DatabaseID,
	}

	return client.Exec(query, args...)
}

// QueryDB executes a query on a specific database in the pool
// Like sqlx: result := pool.QueryDB(dbName, "SELECT * FROM users")
func (p *ConnectionPool) QueryDB(dbName string, query string, params []string) (*utils.APIResponse, error) {
	p.mu.RLock()
	connInfo, exists := p.connections[dbName]
	p.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("database %s not connected, call Connect first", dbName)
	}

	client := &Client{
		AccountID:  p.accountID,
		APIToken:   p.apiToken,
		DatabaseID: connInfo.DatabaseID,
	}

	return client.Query(query, params)
}

// CreateTable creates a table in the currently connected database
func (p *ConnectionPool) CreateTable(createQuery string) (*utils.APIResponse, error) {
	p.mu.RLock()
	connInfo, exists := p.connections[p.currentDB]
	p.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no database connected, call Connect first")
	}

	client := &Client{
		AccountID:  p.accountID,
		APIToken:   p.apiToken,
		DatabaseID: connInfo.DatabaseID,
	}

	return client.CreateTable(createQuery)
}

// RemoveTable removes a table from the currently connected database
func (p *ConnectionPool) RemoveTable(tableName string) (*utils.APIResponse, error) {
	p.mu.RLock()
	connInfo, exists := p.connections[p.currentDB]
	p.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no database connected, call Connect first")
	}

	client := &Client{
		AccountID:  p.accountID,
		APIToken:   p.apiToken,
		DatabaseID: connInfo.DatabaseID,
	}

	return client.RemoveTable(tableName)
}

// RemoveTableDB removes a table from a specific database in the pool
func (p *ConnectionPool) RemoveTableDB(dbName, tableName string) (*utils.APIResponse, error) {
	p.mu.RLock()
	connInfo, exists := p.connections[dbName]
	p.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("database %s not connected, call Connect first", dbName)
	}

	client := &Client{
		AccountID:  p.accountID,
		APIToken:   p.apiToken,
		DatabaseID: connInfo.DatabaseID,
	}

	return client.RemoveTable(tableName)
}

// CreateTableDB creates a table in a specific database in the pool
func (p *ConnectionPool) CreateTableDB(dbName, createQuery string) (*utils.APIResponse, error) {
	p.mu.RLock()
	connInfo, exists := p.connections[dbName]
	p.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("database %s not connected, call Connect first", dbName)
	}

	client := &Client{
		AccountID:  p.accountID,
		APIToken:   p.apiToken,
		DatabaseID: connInfo.DatabaseID,
	}

	return client.CreateTable(createQuery)
}

// GetCurrentDB returns the name of the currently connected database
func (p *ConnectionPool) GetCurrentDB() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.currentDB
}

// GetDatabaseID returns the ID of a cached database connection
func (p *ConnectionPool) GetDatabaseID(dbName string) string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if connInfo, exists := p.connections[dbName]; exists {
		return connInfo.DatabaseID
	}
	return ""
}

// ClearCache removes a database from cache, forcing re-query on next Connect
func (p *ConnectionPool) ClearCache(dbName string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.connections, dbName)
}

// ClearAllCache removes all databases from cache
func (p *ConnectionPool) ClearAllCache() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.connections = make(map[string]*ConnectionInfo)
	p.currentDB = ""
}

// SetCacheAge sets the maximum age for cached connections
// Default is 24 hours. Set to 0 for no caching.
func (p *ConnectionPool) SetCacheAge(duration time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.maxCacheAge = duration
}

// SetAutoReconnect enables/disables automatic reconnection on failure
func (p *ConnectionPool) SetAutoReconnect(enabled bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.autoReconnect = enabled
}

// ListCachedDatabases returns a list of all cached database names
func (p *ConnectionPool) ListCachedDatabases() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var dbNames []string
	for name := range p.connections {
		dbNames = append(dbNames, name)
	}
	return dbNames
}

// IsCached checks if a database connection is cached
func (p *ConnectionPool) IsCached(dbName string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if connInfo, exists := p.connections[dbName]; exists {
		return time.Since(connInfo.CachedAt) < p.maxCacheAge
	}
	return false
}

// GetCacheInfo returns information about a cached connection
func (p *ConnectionPool) GetCacheInfo(dbName string) *ConnectionInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if connInfo, exists := p.connections[dbName]; exists {
		// Return a copy to prevent external modification
		return &ConnectionInfo{
			DatabaseID: connInfo.DatabaseID,
			Name:       connInfo.Name,
			CachedAt:   connInfo.CachedAt,
		}
	}
	return nil
}
