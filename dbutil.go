package dbutil

import "fmt"

type EntityProvider[E Entity] func() E

type DBUtil[E Entity] struct {
	db       DBTx
	dialect  SQLDialect
	provider EntityProvider[E]
}

func New[E Entity](db DBTx, dialect SQLDialect, provider EntityProvider[E]) *DBUtil[E] {
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

func (dbu *DBUtil[E]) FormatBindParams(str string, param_count int) string {
	params := []any{}
	for i := 1; i <= param_count; i++ {
		params = append(params, dbu.BindParam(i))
	}
	return fmt.Sprintf(str, params...)
}
