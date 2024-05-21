package migrations

import "embed"

//go:embed *.sql
var FSMigrations embed.FS
