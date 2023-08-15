package dbutil

import "fmt"

func (dbu *DBUtil[E]) Delete(constraints string, params ...any) error {
	entity := dbu.provider()
	_, err := dbu.db.Exec(
		fmt.Sprintf("delete from %s %s", entity.Table(), dbu.FormatBindParams(constraints, len(params))),
		params...,
	)
	return err
}
