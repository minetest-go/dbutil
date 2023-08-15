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
	provider := func() *MyTable { return &MyTable{} }
	dbu := dbutil.New(db, dbutil.DialectSQLite, provider)

	// insert
	assert.NoError(t, dbu.Insert(&MyTable{F1: 1}))

	// insert with return value (sqlite specific: "INSERT RETURNING")
	var retVal int64 = -1
	assert.NoError(t, dbu.InsertReturning(&MyTable{F1: 2}, "pk", &retVal))
	assert.True(t, retVal >= 0)

	// insert or replace (sqlite flavor)
	assert.NoError(t, dbu.InsertOrReplace(&MyTable{F1: 2, PK: &retVal}))

	// select single
	res, err := dbu.Select("where f1 = $1", 1)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, 1, res.F1)

	// select single non-existent
	_, err = dbu.Select("where f1 = $1", 0)
	assert.Error(t, err)
	assert.ErrorIs(t, sql.ErrNoRows, err)

	// select multi
	list, err := dbu.SelectMulti("where true=true")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(list))

	// count all
	count, err := dbu.Count("where true=true")
	assert.NoError(t, err)
	assert.Equal(t, 2, count)

	// count all (in tx)
	tx, err := db.Begin()
	dbutx := dbutil.New(tx, dbutil.DialectSQLite, provider)

	assert.NoError(t, err)
	count, err = dbutx.Count("where true=true")
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
	assert.NoError(t, tx.Commit())

	// update (where f2 = 2)
	tbl, err := dbu.Select("where f1 = $1", 2)
	assert.NoError(t, err)
	assert.NotNil(t, tbl)
	tbl.F1 = 3
	err = dbu.Update(tbl, "where f1 = $1", 2)
	assert.NoError(t, err)

	// count specific
	count, err = dbu.Count("where f1 = $1", 3)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	// delete
	err = dbu.Delete("where f1 = $1", 3)
	assert.NoError(t, err)

	// count specific (after delete)
	count, err = dbu.Count("where f1 = $1", 3)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)

}
