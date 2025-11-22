package cloudflared1

import (
	"encoding/json"
	"fmt"

	"github.com/youfun/cloudflare-d1-go/utils"
)

type Client struct {
	AccountID  string
	APIToken   string
	DatabaseID string
}

func NewClient(accountID, apiToken string) *Client {
	if accountID == "" || apiToken == "" {
		return nil
	}
	return &Client{
		AccountID: accountID,
		APIToken:  apiToken,
	}
}

func (c *Client) ListDB() (*utils.APIResponse, error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/d1/database", c.AccountID)
	return utils.DoRequest("GET", url, "", c.APIToken)
}

func (c *Client) CreateDB(name string) (*utils.APIResponse, error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/d1/database", c.AccountID)
	body := fmt.Sprintf(`{"name":"%s"}`, name)
	return utils.DoRequest("POST", url, body, c.APIToken)
}

func (c *Client) DeleteDB(databaseID string) (*utils.APIResponse, error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/d1/database/%s", c.AccountID, databaseID)
	return utils.DoRequest("DELETE", url, "", c.APIToken)
}

// Runs SQL query on the D1 database with parameters
func (c *Client) QueryDB(databaseID string, query string, params []string) (*utils.APIResponse, error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/d1/database/%s/raw", c.AccountID, databaseID)

	// Build request body with proper JSON encoding
	requestBody := map[string]interface{}{
		"sql":    query,
		"params": params,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	return utils.DoRequest("POST", url, string(bodyBytes), c.APIToken)
}

func (c *Client) CreateTableWithID(databaseID, createQuery string) (*utils.APIResponse, error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/d1/database/%s/raw", c.AccountID, databaseID)

	requestBody := map[string]interface{}{
		"sql":    createQuery,
		"params": []string{},
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	return utils.DoRequest("POST", url, string(bodyBytes), c.APIToken)
}

func (c *Client) RemoveTableWithID(databaseID, tableName string) (*utils.APIResponse, error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/d1/database/%s/raw", c.AccountID, databaseID)
	query := fmt.Sprintf("DROP TABLE IF EXISTS %s;", tableName)

	requestBody := map[string]interface{}{
		"sql":    query,
		"params": []string{},
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	return utils.DoRequest("POST", url, string(bodyBytes), c.APIToken)
}

// ConnectDB finds and connects to a database by name, storing its ID for future operations
func (c *Client) ConnectDB(name string) error {
	resp, err := c.ListDB()
	if err != nil {
		return fmt.Errorf("failed to list databases: %w", err)
	}

	// Parse response to find database with matching name
	databases := resp.Result.([]interface{})
	for _, db := range databases {
		dbMap := db.(map[string]interface{})
		if dbMap["name"].(string) == name {
			c.DatabaseID = dbMap["uuid"].(string)
			return nil
		}
	}

	return fmt.Errorf("database with name %s not found", name)
}

// Query runs SQL query on the connected database
func (c *Client) Query(query string, params []string) (*utils.APIResponse, error) {
	if c.DatabaseID == "" {
		return nil, fmt.Errorf("no database connected, call ConnectDB first")
	}
	return c.QueryDB(c.DatabaseID, query, params)
}

// CreateTable creates a table in the connected database
func (c *Client) CreateTable(createQuery string) (*utils.APIResponse, error) {
	if c.DatabaseID == "" {
		return nil, fmt.Errorf("no database connected, call ConnectDB first")
	}
	return c.CreateTableWithID(c.DatabaseID, createQuery)
}

// RemoveTable removes a table from the connected database
func (c *Client) RemoveTable(tableName string) (*utils.APIResponse, error) {
	if c.DatabaseID == "" {
		return nil, fmt.Errorf("no database connected, call ConnectDB first")
	}
	return c.RemoveTableWithID(c.DatabaseID, tableName)
}

// Select executes a query and scans all results into a slice, similar to sqlx.Select
// Like sqlx: client.Select(&users, "SELECT * FROM users WHERE age > ?", 25)
func (c *Client) Select(dest interface{}, query string, args ...interface{}) error {
	params, err := utils.ConvertParams(args...)
	if err != nil {
		return err
	}
	res, err := c.Query(query, params)
	if err != nil {
		return err
	}
	return res.StructScanAll(dest)
}

// Get executes a query and scans the first result into a struct, similar to sqlx.Get
// Like sqlx: client.Get(&user, "SELECT * FROM users WHERE id = ?", 123)
func (c *Client) Get(dest interface{}, query string, args ...interface{}) error {
	params, err := utils.ConvertParams(args...)
	if err != nil {
		return err
	}
	res, err := c.Query(query, params)
	if err != nil {
		return err
	}
	return res.Get(dest)
}

// Exec executes a query and returns the number of rows affected, similar to sqlx.Exec
// Like sqlx: rowsAffected, err := client.Exec("UPDATE users SET age = ? WHERE id = ?", 30, 123)
func (c *Client) Exec(query string, args ...interface{}) (int64, error) {
	params, err := utils.ConvertParams(args...)
	if err != nil {
		return 0, err
	}

	res, err := c.Query(query, params)
	if err != nil {
		return 0, err
	}

	result, err := res.ToResult()
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}
