package migrations

import (
	"embed"
	"fmt"
	"net/http"
	"os"
	"path"
	"sort"
	"strings"
)

type byId []*Migration

func (b byId) Len() int           { return len(b) }
func (b byId) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b byId) Less(i, j int) bool { return b[i].Less(b[j]) }

// A hardcoded set of migrations, in-memory.
type MemoryMigrationSource struct {
	Migrations []*Migration
}

var _ MigrationSource = (*MemoryMigrationSource)(nil)

func (m MemoryMigrationSource) FindMigrations() ([]*Migration, error) {
	migrations := make([]*Migration, len(m.Migrations))
	copy(migrations, m.Migrations)
	sort.Sort(byId(migrations))
	return migrations, nil
}

// A set of migrations loaded from a directory.
type FileMigrationSource struct {
	Dir string
}

var _ MigrationSource = (*FileMigrationSource)(nil)

func (f FileMigrationSource) FindMigrations() ([]*Migration, error) {
	filesystem := http.Dir(f.Dir)
	return findMigrations(filesystem, "/")
}

// A set of migrations loaded from an go1.16 embed.FS
type EmbedFileSystemMigrationSource struct {
	FileSystem embed.FS
	Root       string
}

var _ MigrationSource = (*EmbedFileSystemMigrationSource)(nil)

func (f EmbedFileSystemMigrationSource) FindMigrations() ([]*Migration, error) {
	return findMigrations(http.FS(f.FileSystem), f.Root)
}

func findMigrations(dir http.FileSystem, root string) ([]*Migration, error) {
	migrations := make([]*Migration, 0)

	file, err := dir.Open(root)
	if err != nil {
		return nil, err
	}

	files, err := file.Readdir(0)
	if err != nil {
		return nil, err
	}

	for _, info := range files {
		if strings.HasSuffix(info.Name(), ".sql") {
			migration, err := migrationFromFile(dir, root, info)
			if err != nil {
				return nil, err
			}

			migrations = append(migrations, migration)
		}
	}

	// Make sure migrations are sorted
	sort.Sort(byId(migrations))

	return migrations, nil
}

func migrationFromFile(dir http.FileSystem, root string, info os.FileInfo) (*Migration, error) {
	path := path.Join(root, info.Name())
	file, err := dir.Open(path)
	if err != nil {
		return nil, fmt.Errorf("Error while opening %s: %w", info.Name(), err)
	}
	defer func() { _ = file.Close() }()

	migration, err := ParseMigration(info.Name(), file)
	if err != nil {
		return nil, fmt.Errorf("Error while parsing %s: %w", info.Name(), err)
	}
	return migration, nil
}
