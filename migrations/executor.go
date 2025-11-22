package migrations

import (
	"fmt"
	"time"

	cloudflare_d1_go "github.com/youfun/cloudflare-d1-go/client"
)

type MigrationSet struct {
	TableName string
}

var migSet = MigrationSet{}

func (ms MigrationSet) getTableName() string {
	if ms.TableName == "" {
		return "d1_migrations"
	}
	return ms.TableName
}

// SetTable sets the name of the table used to store migration info.
func SetTable(name string) {
	migSet.TableName = name
}

type MigrationRecord struct {
	Id        string    `json:"id"`
	AppliedAt time.Time `json:"applied_at"`
}

// Exec executes a set of migrations
func Exec(client *cloudflare_d1_go.Client, m MigrationSource, dir MigrationDirection) (int, error) {
	return ExecMax(client, m, dir, 0)
}

// ExecMax executes a set of migrations with a limit
func ExecMax(client *cloudflare_d1_go.Client, m MigrationSource, dir MigrationDirection, max int) (int, error) {
	return migSet.ExecMax(client, m, dir, max)
}

func (ms MigrationSet) ExecMax(client *cloudflare_d1_go.Client, m MigrationSource, dir MigrationDirection, max int) (int, error) {
	// 1. Ensure migration table exists
	err := ms.ensureTable(client)
	if err != nil {
		return 0, fmt.Errorf("failed to ensure migration table: %w", err)
	}

	// 2. Get applied migrations
	applied, err := ms.getAppliedMigrations(client)
	if err != nil {
		return 0, fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// 3. Get all available migrations
	allMigrations, err := m.FindMigrations()
	if err != nil {
		return 0, fmt.Errorf("failed to find migrations: %w", err)
	}

	// 4. Plan migrations
	toApply := ms.planMigrations(allMigrations, applied, dir, max)

	// 5. Apply migrations
	count := 0
	for _, migration := range toApply {
		err := ms.applyMigration(client, migration, dir)
		if err != nil {
			return count, fmt.Errorf("failed to apply migration %s: %w", migration.Id, err)
		}
		count++
	}

	return count, nil
}

func (ms MigrationSet) ensureTable(client *cloudflare_d1_go.Client) error {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		id TEXT PRIMARY KEY,
		applied_at DATETIME
	);`, ms.getTableName())

	_, err := client.CreateTable(query)
	return err
}

func (ms MigrationSet) getAppliedMigrations(client *cloudflare_d1_go.Client) ([]string, error) {
	query := fmt.Sprintf("SELECT id FROM %s ORDER BY id ASC;", ms.getTableName())
	res, err := client.Query(query, nil)
	if err != nil {
		// If table doesn't exist yet (should be handled by ensureTable, but just in case)
		return nil, err
	}

	// Use ToRows to iterate
	rows, err := res.ToRows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		// We need to scan into a struct or map usually with this client,
		// but let's see if we can scan into a simple string if it's one column?
		// The client's StructScan expects a struct.
		var record struct {
			Id string `db:"id"`
		}
		if err := rows.StructScan(&record); err != nil {
			return nil, err
		}
		ids = append(ids, record.Id)
	}
	return ids, nil
}

func (ms MigrationSet) planMigrations(all []*Migration, applied []string, dir MigrationDirection, max int) []*Migration {
	appliedMap := make(map[string]bool)
	for _, id := range applied {
		appliedMap[id] = true
	}

	var toApply []*Migration

	if dir == Up {
		for _, m := range all {
			if !appliedMap[m.Id] {
				toApply = append(toApply, m)
			}
		}
	} else {
		// Down: find applied migrations in reverse order
		// We need to map applied IDs back to Migration objects
		migrationMap := make(map[string]*Migration)
		for _, m := range all {
			migrationMap[m.Id] = m
		}

		// Iterate applied in reverse
		for i := len(applied) - 1; i >= 0; i-- {
			id := applied[i]
			if m, ok := migrationMap[id]; ok {
				toApply = append(toApply, m)
			}
		}
	}

	if max > 0 && len(toApply) > max {
		toApply = toApply[:max]
	}

	return toApply
}

func (ms MigrationSet) applyMigration(client *cloudflare_d1_go.Client, m *Migration, dir MigrationDirection) error {
	queries := m.Up
	if dir == Down {
		queries = m.Down
	}

	// Execute queries
	// TODO: Transaction support if D1 supports it via batch?
	// For now, execute sequentially.
	for _, q := range queries {
		_, err := client.Query(q, nil)
		if err != nil {
			return err
		}
	}

	// Record migration
	if dir == Up {
		query := fmt.Sprintf("INSERT INTO %s (id, applied_at) VALUES (?, ?);", ms.getTableName())
		_, err := client.Query(query, []string{m.Id, time.Now().Format(time.RFC3339)})
		if err != nil {
			return err
		}
	} else {
		query := fmt.Sprintf("DELETE FROM %s WHERE id = ?;", ms.getTableName())
		_, err := client.Query(query, []string{m.Id})
		if err != nil {
			return err
		}
	}

	return nil
}
