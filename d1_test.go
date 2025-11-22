package cloudflared1_test

import (
	"testing"

	cloudflare_d1_go "github.com/ashayas/cloudflare-d1-go/client"
	"github.com/ashayas/cloudflare-d1-go/utils"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name      string
		accountID string
		apiToken  string
		wantErr   bool
	}{
		{
			name:      "valid credentials",
			accountID: "1234567890",
			apiToken:  "1234567890",
			wantErr:   false,
		},
		{
			name:      "empty account ID",
			accountID: "",
			apiToken:  "1234567890",
			wantErr:   true,
		},
		{
			name:      "empty API token",
			accountID: "1234567890",
			apiToken:  "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := cloudflare_d1_go.NewClient(tt.accountID, tt.apiToken)

			if tt.wantErr {
				if client != nil {
					t.Errorf("NewClient() = %v, want nil for invalid inputs", client)
				}
				return
			}

			if client == nil {
				t.Fatal("NewClient() returned nil for valid inputs")
			}

			if client.AccountID != tt.accountID {
				t.Errorf("NewClient().AccountID = %v, want %v", client.AccountID, tt.accountID)
			}

			if client.APIToken != tt.apiToken {
				t.Errorf("NewClient().APIToken = %v, want %v", client.APIToken, tt.apiToken)
			}
		})
	}
}

// TestListDB lists the databases
func TestListDB(t *testing.T) {
	client := cloudflare_d1_go.NewClient("account_id", "api_token")
	res, err := client.ListDB()
	if err != nil {
		t.Errorf("ListDB failed: %v", err)
	}
	t.Logf("ListDB response: %+v", res)

	if res == nil {
		t.Error("Expected non-nil response from ListDB")
	}
}

// TestCreateAndDeleteDB creates a database, then deletes it
func TestCreateAndDeleteDB(t *testing.T) {
	client := cloudflare_d1_go.NewClient("account_id", "api_token")
	res, err := client.CreateDB("test-db-2")
	if err != nil {
		t.Errorf("CreateDB failed: %v", err)
	}
	t.Logf("CreateDB response: %+v", res)

	if res == nil {
		t.Error("Expected non-nil response from CreateDB")
	}

	// Only do this if the database was created successfully
	if res != nil && res.Success {
		res, err = client.DeleteDB(res.Result.(map[string]interface{})["uuid"].(string))
		if err != nil {
			t.Errorf("DeleteDB failed: %v", err)
		}
		t.Logf("DeleteDB response: %+v", res)

		if res == nil {
			t.Error("Expected non-nil response from DeleteDB")
		}
	}

}

// TestCreateAndRemoveTable creates a table, then removes it
func TestCreateAndRemoveTable(t *testing.T) {
	client := cloudflare_d1_go.NewClient("account_id", "api_token")

	// Create a test database
	res, err := client.CreateDB("test_db_3")
	if err != nil {
		t.Errorf("CreateDB failed: %v", err)
		return
	}

	if !res.Success {
		t.Errorf("CreateDB was not successful: %v", res.Errors)
		return
	}

	dbID, ok := res.Result.(map[string]interface{})["uuid"].(string)
	if !ok {
		t.Error("Failed to get database UUID from response")
		return
	}

	// Create a test table
	createQuery := "CREATE TABLE IF NOT EXISTS test_table (id INTEGER PRIMARY KEY, name TEXT);"
	res, err = client.CreateTableWithID(dbID, createQuery)
	if err != nil {
		t.Errorf("CreateTable failed: %v", err)
		return
	}

	if !res.Success {
		t.Errorf("CreateTable was not successful: %v", res.Errors)
		return
	}

	t.Logf("CreateTable response: %+v", res)

	// Only attempt to remove if table was created successfully
	res, err = client.RemoveTableWithID(dbID, "test_table")
	if err != nil {
		t.Errorf("RemoveTable failed: %v", err)
	}
	t.Logf("RemoveTable response: %+v", res)

	if !res.Success {
		t.Errorf("RemoveTable was not successful: %v", res.Errors)
	}
}

// TestQueryDB creates a table, inserts a row, then selects it and deletes the table and database
func TestQueryDB(t *testing.T) {
	client := cloudflare_d1_go.NewClient("account_id", "api_token")

	// Create a test database
	res, err := client.CreateDB("test_db_6")
	if err != nil {
		t.Errorf("CreateDB failed: %v", err)
		return
	}

	if !res.Success {
		t.Errorf("CreateDB was not successful: %v", res.Errors)
		return
	}

	dbID, ok := res.Result.(map[string]interface{})["uuid"].(string)
	if !ok {
		t.Error("Failed to get database UUID from response")
		return
	}

	// Create a test table
	createQuery := "CREATE TABLE IF NOT EXISTS test_table (id INTEGER PRIMARY KEY, name TEXT);"
	res, err = client.CreateTableWithID(dbID, createQuery)
	if err != nil || !res.Success {
		t.Errorf("CreateTable failed: %v, errors: %v", err, res.Errors)
		return
	}

	t.Logf("CreateTable response: %+v", res)

	// Insert test data
	insertQuery := "INSERT INTO test_table (name) VALUES (?);"
	params := []string{"test_name"}
	res, err = client.QueryDB(dbID, insertQuery, params)
	if err != nil || !res.Success {
		t.Errorf("Insert query failed: %v, errors: %v", err, res.Errors)
		return
	}

	// Select the data
	selectQuery := "SELECT * FROM test_table WHERE name = ?;"
	res, err = client.QueryDB(dbID, selectQuery, params)
	if err != nil {
		t.Errorf("Select query failed: %v", err)
		return
	}

	if !res.Success {
		t.Errorf("Select query was not successful: %v", res.Errors)
		return
	}

	// Parse the response correctly
	results, ok := res.Result.([]interface{})
	if !ok {
		t.Error("Failed to parse Result as array")
		return
	}

	if len(results) == 0 {
		t.Error("No results returned from query")
		return
	}

	// First element contains the query results
	queryResult, ok := results[0].(map[string]interface{})
	if !ok {
		t.Error("Failed to parse query result")
		return
	}

	// Access the actual rows from the results
	resultsMap, ok := queryResult["results"].(map[string]interface{})
	if !ok {
		t.Error("Failed to parse results map")
		return
	}

	rows, ok := resultsMap["rows"].([]interface{})
	if !ok {
		t.Error("Failed to parse rows")
		return
	}

	// Delete the table
	res, err = client.RemoveTableWithID(dbID, "test_table")
	if err != nil || !res.Success {
		t.Errorf("RemoveTable failed: %v, errors: %v", err, res.Errors)
		return
	}

	// Delete the database
	res, err = client.DeleteDB(dbID)
	if err != nil || !res.Success {
		t.Errorf("DeleteDB failed: %v, errors: %v", err, res.Errors)
		return
	}

	t.Logf("Query returned %d rows", len(rows))
	t.Logf("Full response: %+v", res)
}

// TestRowsScan tests the Rows.Scan and StructScan functionality using mock data
func TestRowsScan(t *testing.T) {
	// Mock data simulating a D1 response
	mockRows := []map[string]interface{}{
		{"name": "Alice", "age": float64(30)}, // JSON numbers are float64
		{"name": "Bob", "age": float64(25)},
	}
	mockColumns := []string{"name", "age"}

	// Create Rows directly
	rows := utils.NewRows(mockRows, mockColumns)
	defer rows.Close()

	// Test StructScan
	type User struct {
		Name string `db:"name"`
		Age  int    `db:"age"`
	}

	var users []User
	for rows.Next() {
		var u User
		if err := rows.StructScan(&u); err != nil {
			t.Errorf("StructScan failed: %v", err)
		}
		users = append(users, u)
	}

	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}

	if users[0].Name != "Alice" || users[0].Age != 30 {
		t.Errorf("Expected Alice (30), got %v", users[0])
	}
	if users[1].Name != "Bob" || users[1].Age != 25 {
		t.Errorf("Expected Bob (25), got %v", users[1])
	}

	// Test Scan
	// Reset rows (NewRows again because we can't reset cursor easily without seeking support)
	rows = utils.NewRows(mockRows, mockColumns)

	var name string
	var age int

	if !rows.Next() {
		t.Fatal("Expected row")
	}
	if err := rows.Scan(&name, &age); err != nil {
		t.Errorf("Scan failed: %v", err)
	}
	if name != "Alice" || age != 30 {
		t.Errorf("Scan mismatch: %s, %d", name, age)
	}
}

// TestExecResult tests the Result functionality using mock data
func TestExecResult(t *testing.T) {
	// Mock data simulating a D1 response for INSERT/UPDATE
	mockMeta := map[string]interface{}{
		"last_row_id":  float64(123),
		"changes":      float64(1),
		"rows_written": float64(1),
	}

	mockResult := map[string]interface{}{
		"meta":    mockMeta,
		"results": map[string]interface{}{}, // Empty results for Exec usually
	}

	apiResponse := &utils.APIResponse{
		Success: true,
		Result:  []interface{}{mockResult},
	}

	result, err := apiResponse.ToResult()
	if err != nil {
		t.Fatalf("ToResult failed: %v", err)
	}

	lastInsertId, err := result.LastInsertId()
	if err != nil {
		t.Errorf("LastInsertId failed: %v", err)
	}
	if lastInsertId != 123 {
		t.Errorf("Expected LastInsertId 123, got %d", lastInsertId)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		t.Errorf("RowsAffected failed: %v", err)
	}
	if rowsAffected != 1 {
		t.Errorf("Expected RowsAffected 1, got %d", rowsAffected)
	}
}
