package migrations

import (
	"regexp"
	"strconv"
)

type MigrationDirection int

const (
	Up MigrationDirection = iota
	Down
)

type Migration struct {
	Id   string
	Up   []string
	Down []string

	DisableTransactionUp   bool
	DisableTransactionDown bool
}

func (m Migration) Less(other *Migration) bool {
	switch {
	case m.isNumeric() && other.isNumeric() && m.VersionInt() != other.VersionInt():
		return m.VersionInt() < other.VersionInt()
	case m.isNumeric() && !other.isNumeric():
		return true
	case !m.isNumeric() && other.isNumeric():
		return false
	default:
		return m.Id < other.Id
	}
}

var numberPrefixRegex = regexp.MustCompile(`^(\d+).*$`)

func (m Migration) isNumeric() bool {
	return len(m.NumberPrefixMatches()) > 0
}

func (m Migration) NumberPrefixMatches() []string {
	return numberPrefixRegex.FindStringSubmatch(m.Id)
}

func (m Migration) VersionInt() int64 {
	v := m.NumberPrefixMatches()[1]
	value, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		// Should not happen if regex matched
		return 0
	}
	return value
}

type PlannedMigration struct {
	*Migration

	DisableTransaction bool
	Queries            []string
}

type MigrationSource interface {
	// Finds the migrations.
	//
	// The resulting slice of migrations should be sorted by Id.
	FindMigrations() ([]*Migration, error)
}
