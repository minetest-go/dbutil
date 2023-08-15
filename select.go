package dbutil

import (
	"database/sql"
	"fmt"
	"strings"
)

func (dbu *DBUtil[E]) Select(constraints string, params ...any) (E, error) {
	entity := dbu.provider()
	row := dbu.db.QueryRow(fmt.Sprintf(
		"select %s from %s %s",
		strings.Join(entity.Columns(SelectAction), ","), entity.Table(), constraints),
		params...,
	)
	err := entity.Scan(SelectAction, row.Scan)
	return entity, err
}

func (dbu *DBUtil[E]) SelectMulti(constraints string, params ...any) ([]E, error) {
	entity := dbu.provider()
	rows, err := dbu.db.Query(fmt.Sprintf("select %s from %s %s", strings.Join(entity.Columns(SelectAction), ","), entity.Table(), constraints), params...)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	list := make([]E, 0)
	for rows.Next() {
		entry := dbu.provider()
		err = entry.Scan(SelectAction, rows.Scan)
		if err != nil {
			return nil, err
		}

		list = append(list, entry)
	}

	return list, nil
}
