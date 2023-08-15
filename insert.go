package dbutil

import (
	"fmt"
	"strings"
)

func (dbu *DBUtil[E]) Insert(entity Insertable, additionalStmts ...string) error {
	cols := entity.Columns(InsertAction)
	placeholders := make([]string, len(cols))
	for i := range cols {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	_, err := dbu.db.Exec(fmt.Sprintf(
		"insert into %s(%s) values(%s) %s",
		entity.Table(), strings.Join(cols, ","), strings.Join(placeholders, ","), strings.Join(additionalStmts, " ")),
		entity.Values(InsertAction)...,
	)

	return err
}

func (dbu *DBUtil[E]) InsertOrReplace(entity Insertable, additionalStmts ...string) error {
	cols := entity.Columns(InsertAction)
	placeholders := make([]string, len(cols))
	for i := range cols {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	_, err := dbu.db.Exec(fmt.Sprintf(
		"insert or replace into %s(%s) values(%s) %s",
		entity.Table(), strings.Join(cols, ","), strings.Join(placeholders, ","), strings.Join(additionalStmts, " ")),
		entity.Values(InsertAction)...,
	)

	return err
}

func (dbu *DBUtil[E]) InsertReturning(entity Insertable, retField string, retValue any) error {
	cols := entity.Columns(InsertAction)
	placeholders := make([]string, len(cols))
	for i := range cols {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	stmt, err := dbu.db.Prepare(fmt.Sprintf(
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
