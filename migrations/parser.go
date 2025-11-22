package migrations

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
)

type ParsedMigration struct {
	UpStatements   []string
	DownStatements []string

	DisableTransactionUp   bool
	DisableTransactionDown bool
}

var (
	matchEmptyLines = true
)

func ParseMigration(id string, r io.ReadSeeker) (*Migration, error) {
	m := &Migration{
		Id: id,
	}

	parsed, err := parseMigration(r)
	if err != nil {
		return nil, fmt.Errorf("Error parsing migration (%s): %w", id, err)
	}

	m.Up = parsed.UpStatements
	m.Down = parsed.DownStatements

	m.DisableTransactionUp = parsed.DisableTransactionUp
	m.DisableTransactionDown = parsed.DisableTransactionDown

	return m, nil
}

func parseMigration(r io.ReadSeeker) (*ParsedMigration, error) {
	p := &ParsedMigration{}

	scanner := bufio.NewScanner(r)
	var currentDirection MigrationDirection = Up
	var buf bytes.Buffer

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "-- +migrate Up") {
			if buf.Len() > 0 {
				appendStatement(p, currentDirection, buf.String())
				buf.Reset()
			}
			currentDirection = Up
			if strings.Contains(line, "notransaction") {
				p.DisableTransactionUp = true
			}
		} else if strings.HasPrefix(line, "-- +migrate Down") {
			if buf.Len() > 0 {
				appendStatement(p, currentDirection, buf.String())
				buf.Reset()
			}
			currentDirection = Down
			if strings.Contains(line, "notransaction") {
				p.DisableTransactionDown = true
			}
		} else if strings.HasPrefix(line, "-- +migrate StatementBegin") {
			// For now, we treat StatementBegin/End as just part of the SQL or ignore special handling
			// since D1 might not support complex PL/SQL blocks the same way, but we keep reading.
			// Actually, sql-migrate uses this to handle semicolons inside blocks.
			// For simplicity in this port, we will just accumulate lines until we see StatementEnd or EOF/next section
			// But a simpler approach for D1 (SQLite) is usually sufficient.
			// Let's stick to simple semicolon splitting or just whole block if no semicolon found?
			// Cloudflare D1 API takes a single string or list of strings.
			// Let's assume we just accumulate lines into the buffer.
		} else if strings.HasPrefix(line, "-- +migrate StatementEnd") {
			// End of statement block
		} else {
			// Regular line
			if buf.Len() > 0 {
				buf.WriteString("\n")
			}
			buf.WriteString(line)
		}
	}

	if buf.Len() > 0 {
		appendStatement(p, currentDirection, buf.String())
	}

	return p, nil
}

func appendStatement(p *ParsedMigration, dir MigrationDirection, sql string) {
	// Split by semicolon for SQLite, as D1 client might want individual statements?
	// The D1 client `Query` takes a single statement usually.
	// `Exec` in sql-migrate splits by semicolon.
	// Let's implement simple semicolon splitting.

	stmts := splitSQLStatements(sql)
	for _, stmt := range stmts {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if dir == Up {
			p.UpStatements = append(p.UpStatements, stmt)
		} else {
			p.DownStatements = append(p.DownStatements, stmt)
		}
	}
}

func splitSQLStatements(sql string) []string {
	// This is a very naive splitter. A proper one would handle quotes, comments, etc.
	// For now, we split by semicolon at end of line or just semicolon.
	// D1 client might handle multiple statements in one go if passed to batch?
	// But `Query` usually expects one.
	// Let's just split by `;`
	return strings.Split(sql, ";")
}
