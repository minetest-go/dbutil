package dbutil

import (
	"fmt"
	"strings"
)

func (dbu *DBUtil[E]) Update(entity Insertable, constraints string, params ...any) error {
	cols := entity.Columns(UpdateAction)
	updates := make([]string, len(cols))
	values := entity.Values(UpdateAction)

	// start params-number after provided params
	pi := len(params) + 1
	for i := range cols {
		updates[i] = fmt.Sprintf("%s = ?%d", cols[i], pi)
		params = append(params, values[i])
		pi++
	}

	sql := fmt.Sprintf(
		"update %s set %s %s",
		entity.Table(), strings.Join(updates, ","), constraints)

	_, err := dbu.db.Exec(sql, params...)

	return err
}
