package dbutil_test

import (
	"database/sql"
	"os"
	"path"
	"testing"

	"github.com/minetest-go/dbutil"

	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

func setupDB(t *testing.T) *sql.DB {
	tmpdir, err := os.MkdirTemp(os.TempDir(), "dbutil")
	assert.NoError(t, err)
	db_, err := sql.Open("sqlite", path.Join(tmpdir, "dbutil.sqlite"))
	assert.NoError(t, err)
	db_.SetMaxOpenConns(1)

	_, err = db_.Exec(`
		create table mytable(
			pk integer primary key autoincrement,
			f1 int
		);
	`)
	assert.NoError(t, err)
	return db_
}

// database struct
type MyTable struct {
	PK *int64 // optional primary key (autoincremented)
	F1 int    // simple field
}

// entity
func (t *MyTable) Table() string {
	return "mytable"
}
func (t *MyTable) Columns(action string) []string {
	return []string{"pk", "f1"}
}

// selectable (implies deletable)
func (t *MyTable) Scan(action string, r func(dest ...any) error) error {
	return r(&t.PK, &t.F1)
}

// insertable
func (t *MyTable) Values(action string) []any {
	return []any{t.PK, t.F1}
}

// testing below here

func Test(t *testing.T) {
	// setup
	db := setupDB(t)

	// insert
	assert.NoError(t, dbutil.Insert(db, &MyTable{F1: 1}))

	// inser with return value (sqlite specific: "INSERT RETURNING")
	var retVal int64 = -1
	assert.NoError(t, dbutil.InsertReturning(db, &MyTable{F1: 2}, "pk", &retVal))
	assert.True(t, retVal >= 0)

	// select single
	res, err := dbutil.Select(db, &MyTable{}, "where f1 = $1", 1)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, 1, res.F1)

	// select single non-existent
	_, err = dbutil.Select(db, &MyTable{}, "where f1 = $1", 0)
	assert.Error(t, err)
	assert.ErrorIs(t, sql.ErrNoRows, err)

	// select multi
	list, err := dbutil.SelectMulti(db, func() *MyTable { return &MyTable{} }, "where true=true")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(list))

	// count all
	count, err := dbutil.Count(db, &MyTable{}, "where true=true")
	assert.NoError(t, err)
	assert.Equal(t, 2, count)

	// count all (in tx)
	tx, err := db.Begin()
	assert.NoError(t, err)
	count, err = dbutil.Count(tx, &MyTable{}, "where true=true")
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
	assert.NoError(t, tx.Commit())

	// update (where f2 = 2)
	tbl, err := dbutil.Select(db, &MyTable{}, "where f1 = $1", 2)
	assert.NoError(t, err)
	assert.NotNil(t, tbl)
	tbl.F1 = 3
	err = dbutil.Update(db, tbl, "where f1 = $1", 2)
	assert.NoError(t, err)

	// count specific
	count, err = dbutil.Count(db, &MyTable{}, "where f1 = $1", 3)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	// delete
	err = dbutil.Delete(db, &MyTable{}, "where f1 = $1", 3)
	assert.NoError(t, err)

	// count specific (after delete)
	count, err = dbutil.Count(db, &MyTable{}, "where f1 = $1", 3)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)

}
