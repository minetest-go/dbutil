package dbutil

import (
	"fmt"
	"strings"
)

func (dbu *DBUtil[E]) PrepareInsert(additionalStmts ...string) (func(E) error, error) {
	entity := dbu.provider()
	cols := entity.Columns(InsertAction)
	placeholders := make([]string, len(cols))
	for i := range cols {
		placeholders[i] = dbu.BindParam(i + 1)
	}

	stmt, err := dbu.db.Prepare(fmt.Sprintf(
		"insert into %s(%s) values(%s) %s",
		entity.Table(), strings.Join(cols, ","), strings.Join(placeholders, ","), strings.Join(additionalStmts, " ")),
	)
	if err != nil {
		return nil, err
	}

	return func(i E) error {
		_, err := stmt.Exec(i.Values(InsertAction)...)
		return err
	}, nil
}
