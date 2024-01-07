package dbutil

import "database/sql"

type SQLDialect int

const (
	DialectSQLite SQLDialect = iota
	DialectPostgres
)

const (
	InsertAction = "insert"
	UpdateAction = "update"
	SelectAction = "select"
)

type DBTx interface {
	Exec(query string, args ...any) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
	QueryRow(query string, args ...any) *sql.Row
	Query(query string, args ...any) (*sql.Rows, error)
}

type Entity interface {
	Columns(action string) []string
	Table() string
	Scan(action string, r func(dest ...any) error) error
	Values(action string) []any
}
