package dbutil

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
