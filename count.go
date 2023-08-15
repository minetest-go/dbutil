package dbutil

import "fmt"

func (dbu *DBUtil[E]) Count(constraints string, params ...any) (int, error) {
	entity := dbu.provider()
	row := dbu.db.QueryRow(
		fmt.Sprintf("select count(*) from %s %s",
			entity.Table(), dbu.FormatBindParams(constraints, len(params))), params...)
	var count int
	return count, row.Scan(&count)
}
