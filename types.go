package dbutil

const (
	InsertAction = "insert"
	UpdateAction = "update"
	SelectAction = "select"
)

type Entity interface {
	Columns(action string) []string
	Table() string
}

type Selectable interface {
	Entity
	Scan(action string, r func(dest ...any) error) error
}

type Insertable interface {
	Entity
	Values(action string) []any
}
