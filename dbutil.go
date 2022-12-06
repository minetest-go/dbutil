package dbutil

import (
	"database/sql"
	"fmt"
	"strings"
)

func Insert(d *sql.DB, entity Insertable, additionalStmts ...string) error {
	cols := entity.Columns(InsertAction)
	placeholders := make([]string, len(cols))
	for i := range cols {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	_, err := d.Exec(fmt.Sprintf(
		"insert into %s(%s) values(%s) %s",
		entity.Table(), strings.Join(cols, ","), strings.Join(placeholders, ","), strings.Join(additionalStmts, " ")),
		entity.Values(InsertAction)...,
	)

	return err
}

func InsertReturning(d *sql.DB, entity Insertable, retField string, retValue any) error {
	cols := entity.Columns(InsertAction)
	placeholders := make([]string, len(cols))
	for i := range cols {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	stmt, err := d.Prepare(fmt.Sprintf(
		"insert into %s(%s) values(%s) returning %s",
		entity.Table(), strings.Join(cols, ","), strings.Join(placeholders, ","), retField),
	)
	if err != nil {
		return err
	}

	row := stmt.QueryRow(entity.Values(InsertAction)...)
	err = row.Scan(retValue)

	return err
}

func Update(d *sql.DB, entity Insertable, constraints string, params ...any) error {
	cols := entity.Columns(UpdateAction)
	updates := make([]string, len(cols))
	values := entity.Values(UpdateAction)

	// start params-number after provided params
	pi := len(params) + 1
	for i := range cols {
		updates[i] = fmt.Sprintf("%s = $%d", cols[i], pi)
		params = append(params, values[i])
		pi++
	}

	_, err := d.Exec(fmt.Sprintf(
		"update %s set %s %s",
		entity.Table(), strings.Join(updates, ","), constraints),
		params...,
	)

	return err
}

func Select[E Selectable](d *sql.DB, entity E, constraints string, params ...any) (E, error) {
	row := d.QueryRow(fmt.Sprintf(
		"select %s from %s %s",
		strings.Join(entity.Columns(SelectAction), ","), entity.Table(), constraints),
		params...,
	)
	err := entity.Scan(SelectAction, row.Scan)
	return entity, err
}

func Count[E Selectable](d *sql.DB, entity E, constraints string, params ...any) (int, error) {
	row := d.QueryRow(fmt.Sprintf("select count(*) from %s %s", entity.Table(), constraints), params...)
	var count int
	return count, row.Scan(&count)
}

func Delete[E Selectable](d *sql.DB, entity E, constraints string, params ...any) error {
	_, err := d.Exec(
		fmt.Sprintf("delete from %s %s", entity.Table(), constraints),
		params...,
	)
	return err
}

func SelectMulti[E Selectable](d *sql.DB, p func() E, constraints string, params ...any) ([]E, error) {
	entity := p()
	rows, err := d.Query(fmt.Sprintf("select %s from %s %s", strings.Join(entity.Columns(SelectAction), ","), entity.Table(), constraints), params...)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	list := make([]E, 0)
	for rows.Next() {
		entry := p()
		err = entry.Scan(SelectAction, rows.Scan)
		if err != nil {
			return nil, err
		}

		list = append(list, entry)
	}

	return list, nil
}
