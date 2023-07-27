package migrations

import (
	"embed"

	migrate "github.com/rubenv/sql-migrate"
)

//go:embed *.sql
var fileSystem embed.FS

var Source = migrate.EmbedFileSystemMigrationSource{
	FileSystem: fileSystem,
	Root:       ".",
}
