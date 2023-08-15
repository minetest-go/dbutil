package dbutil

import "fmt"

type EntityProvider[E Selectable] func() E

type DBUtil[E Selectable] struct {
	db       DBTx
	dialect  SQLDialect
	provider EntityProvider[E]
}

func New[E Selectable](db DBTx, dialect SQLDialect, provider EntityProvider[E]) *DBUtil[E] {
	return &DBUtil[E]{
		db:       db,
		dialect:  dialect,
		provider: provider,
	}
}

func (dbu *DBUtil[E]) BindParam(i int) string {
	if dbu.dialect == DialectSQLite {
		// special case for sqlite
		return fmt.Sprintf("?%d", i)
	}
	return fmt.Sprintf("$%d", i)
}
