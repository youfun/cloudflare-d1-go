package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type APIResponse struct {
	Result  interface{} `json:"result"`
	Success bool        `json:"success"`
	Errors  []struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"errors"`
}

func DoRequest(method, url, payload, apiToken string) (*APIResponse, error) {
	req, err := http.NewRequest(method, url, strings.NewReader(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var apiRes APIResponse
	if err := json.Unmarshal(body, &apiRes); err != nil {
		return nil, err
	}

	return &apiRes, nil
}

// ToRows converts the APIResponse to a Rows object.
// It expects the result to contain "results" map with "rows" and optional "columns".
func (r *APIResponse) ToRows() (*Rows, error) {
	if !r.Success {
		if len(r.Errors) > 0 {
			return nil, fmt.Errorf("api error: %s", r.Errors[0].Message)
		}
		return nil, fmt.Errorf("api error: unknown")
	}

	// r.Result is usually []interface{} for queries
	results, ok := r.Result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected result format: not an array")
	}

	if len(results) == 0 {
		return NewRows(nil, nil), nil
	}

	// We take the first result set
	queryResult, ok := results[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected result item format")
	}

	// Check for "results" map
	resultsData, ok := queryResult["results"].(map[string]interface{})
	if !ok {
		// Maybe it's directly in queryResult?
		// But based on d1_test.go, it's in "results"
		return nil, fmt.Errorf("missing results map")
	}

	// Extract rows
	rowsRaw, ok := resultsData["rows"].([]interface{})
	if !ok {
		// If rows is not an array, return empty rows instead of error
		return NewRows(nil, nil), nil
	}

	// Extract columns if available
	var columns []string
	if colsRaw, ok := resultsData["columns"].([]interface{}); ok {
		for _, c := range colsRaw {
			if s, ok := c.(string); ok {
				columns = append(columns, s)
			}
		}
	}

	rows := make([]map[string]interface{}, len(rowsRaw))
	for i, row := range rowsRaw {
		rowMap := make(map[string]interface{})

		// Handle two cases: row is a map or row is an array
		switch v := row.(type) {
		case map[string]interface{}:
			// D1 sometimes returns objects
			rowMap = v
		case []interface{}:
			// D1 sometimes returns arrays, map them to columns
			if len(columns) == len(v) {
				for j, col := range columns {
					rowMap[col] = v[j]
				}
			} else {
				return nil, fmt.Errorf("row %d has %d values but expected %d columns", i, len(v), len(columns))
			}
		default:
			return nil, fmt.Errorf("row %d has unexpected type: %T", i, row)
		}

		rows[i] = rowMap
	}

	return NewRows(rows, columns), nil
}

// ToResult converts the APIResponse to a Result object.
// It expects the result to contain "meta" information.
func (r *APIResponse) ToResult() (*Result, error) {
	if !r.Success {
		if len(r.Errors) > 0 {
			return nil, fmt.Errorf("api error: %s", r.Errors[0].Message)
		}
		return nil, fmt.Errorf("api error: unknown")
	}

	// r.Result is usually []interface{} for queries
	results, ok := r.Result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected result format: not an array")
	}

	if len(results) == 0 {
		return NewResult(0, 0), nil
	}

	// We take the first result set
	queryResult, ok := results[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected result item format")
	}

	// Check for "meta" map
	metaData, ok := queryResult["meta"].(map[string]interface{})
	if !ok {
		// If no meta, return 0, 0
		return NewResult(0, 0), nil
	}

	var lastInsertId int64
	var rowsAffected int64

	if val, ok := metaData["last_row_id"]; ok {
		if f, ok := val.(float64); ok {
			lastInsertId = int64(f)
		}
	}

	if val, ok := metaData["rows_read"]; ok {
		// Note: D1 returns rows_read and rows_written.
		// For UPDATE/DELETE, rows_written might be more appropriate for RowsAffected?
		// Or maybe "changes"?
		// Let's check if there is "changes" or similar.
		// D1 API docs say: meta: { changed_db: bool, changes: int, duration: float, last_row_id: int, rows_read: int, rows_written: int, size_after: int }
		// So "changes" seems to be the one.
		if f, ok := val.(float64); ok {
			// Just reading for now, but let's look for "changes"
			_ = f
		}
	}

	if val, ok := metaData["changes"]; ok {
		if f, ok := val.(float64); ok {
			rowsAffected = int64(f)
		}
	} else if val, ok := metaData["rows_written"]; ok {
		// Fallback to rows_written if changes is missing
		if f, ok := val.(float64); ok {
			rowsAffected = int64(f)
		}
	}

	return NewResult(lastInsertId, rowsAffected), nil
}

// StructScanAll converts the APIResponse directly to a slice of structs.
// dest must be a pointer to a slice, for example &[]User{}.
// This is similar to sqlx.NamedQuery().StructScan() pattern but simpler.
//
// Example:
//
//	var users []User
//	err := res.StructScanAll(&users)
//
// The method internally converts the response to Rows and scans all rows
// into the destination slice using the "db" struct tags.
func (r *APIResponse) StructScanAll(dest interface{}) error {
	// Convert to Rows first
	rows, err := r.ToRows()
	if err != nil {
		return err
	}
	defer rows.Close()

	// Use Rows.StructScanAll to scan all rows into the destination slice
	return rows.StructScanAll(dest)
}

// Get converts the APIResponse to a single struct.
// dest must be a pointer to a struct, for example &User{}.
// This is similar to sqlx.DB.Get() and only returns the first row.
//
// Example:
//
//	var user User
//	err := res.Get(&user)
//
// If no rows are returned, it will return an error.
func (r *APIResponse) Get(dest interface{}) error {
	// Convert to Rows first
	rows, err := r.ToRows()
	if err != nil {
		return err
	}
	defer rows.Close()

	// Check if there's at least one row
	if !rows.Next() {
		return fmt.Errorf("sql: no rows in result set")
	}

	// Scan the first row into the destination struct
	return rows.StructScan(dest)
}
