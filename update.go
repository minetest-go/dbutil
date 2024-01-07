package dbutil

import (
	"fmt"
	"strings"
)

func (dbu *DBUtil[E]) Update(entity E, constraints string, params ...any) error {
	cols := entity.Columns(UpdateAction)
	updates := make([]string, len(cols))
	values := entity.Values(UpdateAction)
	constraints = dbu.FormatBindParams(constraints, len(params))

	// start params-number after provided params
	pi := len(params) + 1
	for i := range cols {
		updates[i] = fmt.Sprintf("%s = %s", cols[i], dbu.BindParam(pi))
		params = append(params, values[i])
		pi++
	}

	sql := fmt.Sprintf(
		"update %s set %s %s",
		entity.Table(), strings.Join(updates, ","), constraints)

	_, err := dbu.db.Exec(sql, params...)

	return err
}
