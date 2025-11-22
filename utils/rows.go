package utils

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// Rows simulates sql.Rows and sqlx.Rows behavior
type Rows struct {
	rows    []map[string]interface{}
	columns []string
	current int
	lastErr error
}

// NewRows creates a new Rows instance
func NewRows(rows []map[string]interface{}, columns []string) *Rows {
	// If columns are not provided, try to infer from the first row (unreliable order)
	// But for D1, we hope to get columns from the response.
	// If not, Scan() might be problematic without strict order.
	if len(columns) == 0 && len(rows) > 0 {
		for k := range rows[0] {
			columns = append(columns, k)
		}
	}
	return &Rows{
		rows:    rows,
		columns: columns,
		current: -1,
	}
}

// Next prepares the next result row for reading with the Scan method.
func (r *Rows) Next() bool {
	r.current++
	return r.current < len(r.rows)
}

// Err returns the error, if any, that was encountered during iteration.
func (r *Rows) Err() error {
	return r.lastErr
}

// Columns returns the column names.
func (r *Rows) Columns() ([]string, error) {
	return r.columns, nil
}

// Close closes the Rows, preventing further enumeration.
func (r *Rows) Close() error {
	r.rows = nil
	return nil
}

// Scan copies the columns in the current row into the values pointed at by dest.
// The number of values in dest must be the same as the number of columns in Rows.
func (r *Rows) Scan(dest ...interface{}) error {
	if r.current < 0 || r.current >= len(r.rows) {
		return errors.New("sql: Rows is closed")
	}

	row := r.rows[r.current]
	if len(dest) != len(r.columns) {
		return fmt.Errorf("sql: expected %d destination arguments in Scan, not %d", len(r.columns), len(dest))
	}

	for i, colName := range r.columns {
		val, ok := row[colName]
		if !ok {
			// If column not found in row (should not happen if consistent), treat as nil?
			val = nil
		}

		if err := convertAssign(dest[i], val); err != nil {
			return fmt.Errorf("sql: Scan error on column index %d, name %q: %v", i, colName, err)
		}
	}

	return nil
}

// StructScan scans the current row into a struct.
// It uses the "db" struct tag to map column names to fields.
func (r *Rows) StructScan(dest interface{}) error {
	if r.current < 0 || r.current >= len(r.rows) {
		return errors.New("sql: Rows is closed")
	}

	v := reflect.ValueOf(dest)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return errors.New("sql: StructScan requires a pointer to a struct")
	}

	v = v.Elem()
	t := v.Type()
	row := r.rows[r.current]

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("db")
		if tag == "" {
			tag = strings.ToLower(field.Name)
		}

		if val, ok := row[tag]; ok {
			if err := convertAssign(v.Field(i).Addr().Interface(), val); err != nil {
				return fmt.Errorf("sql: StructScan error on field %s: %v", field.Name, err)
			}
		}
	}

	return nil
}

// StructScanAll scans all remaining rows into a destination slice.
// dest must be a pointer to a slice, for example &[]User{}.
//
// Example:
//
//	var users []User
//	err := rows.StructScanAll(&users)
//
// The method will iterate through all rows starting from the current position
// and append each scanned struct to the destination slice.
func (r *Rows) StructScanAll(dest interface{}) error {
	// Validate that dest is a pointer
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr {
		return fmt.Errorf("dest must be a pointer, got %T", dest)
	}

	// Get the slice from the pointer
	sliceValue := destValue.Elem()
	if sliceValue.Kind() != reflect.Slice {
		return fmt.Errorf("dest must be a pointer to a slice, got %T", dest)
	}

	// Get the element type of the slice
	elemType := sliceValue.Type().Elem()

	// Iterate through all rows
	count := 0
	for r.Next() {
		// Create a new instance of the element type
		elemPtr := reflect.New(elemType)
		elemInterface := elemPtr.Interface()

		// Scan the current row into the element
		if err := r.StructScan(elemInterface); err != nil {
			return fmt.Errorf("StructScan failed at index %d: %w", count, err)
		}

		// Append the element to the slice
		sliceValue.Set(reflect.Append(sliceValue, elemPtr.Elem()))
		count++
	}

	return nil
}

// convertAssign copies to dest the value in src.
// This is a simplified version of database/sql/convert.go
func convertAssign(dest, src interface{}) error {
	// Common case: dest is *string, *int, etc.
	switch d := dest.(type) {
	case *string:
		if src == nil {
			*d = ""
			return nil
		}
		*d = fmt.Sprintf("%v", src)
		return nil
	case *int:
		if src == nil {
			*d = 0
			return nil
		}
		// JSON numbers are often float64
		if f, ok := src.(float64); ok {
			*d = int(f)
			return nil
		}
		// Or string
		if s, ok := src.(string); ok {
			var i int
			if _, err := fmt.Sscanf(s, "%d", &i); err == nil {
				*d = i
				return nil
			}
		}
		return fmt.Errorf("cannot convert %T to int", src)
	case *int64:
		if src == nil {
			*d = 0
			return nil
		}
		if f, ok := src.(float64); ok {
			*d = int64(f)
			return nil
		}
		return fmt.Errorf("cannot convert %T to int64", src)
	case *float64:
		if src == nil {
			*d = 0
			return nil
		}
		if f, ok := src.(float64); ok {
			*d = f
			return nil
		}
		return fmt.Errorf("cannot convert %T to float64", src)
	case *bool:
		if src == nil {
			*d = false
			return nil
		}
		if b, ok := src.(bool); ok {
			*d = b
			return nil
		}
		// 0/1
		if f, ok := src.(float64); ok {
			*d = f != 0
			return nil
		}
		return fmt.Errorf("cannot convert %T to bool", src)
	case *interface{}:
		*d = src
		return nil
	case sql.Scanner:
		return d.Scan(src)
	}

	// Reflection fallback for other types
	v := reflect.ValueOf(dest)
	if v.Kind() != reflect.Ptr {
		return errors.New("destination not a pointer")
	}

	// TODO: Add more robust conversion if needed
	return nil
}
