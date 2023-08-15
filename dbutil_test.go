package dbutil_test

import (
	"database/sql"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/minetest-go/dbutil"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func setupDB(t *testing.T) *sql.DB {
	tmpdir, err := os.MkdirTemp(os.TempDir(), "dbutil")
	assert.NoError(t, err)
	db_, err := sql.Open("sqlite3", path.Join(tmpdir, "dbutil.sqlite"))
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

func ExampleDBUtil() {
	// setup
	db, err := sql.Open("sqlite3", "mydb")
	if err != nil {
		panic(err)
	}

	provider := func() *MyTable { return &MyTable{} }
	dbu := dbutil.New(db, dbutil.DialectSQLite, provider)

	// insert
	err = dbu.Insert(&MyTable{F1: 1})
	if err != nil {
		panic(err)
	}

	// select single
	// NOTE: %s is getting replaced by the db-native bind placeholder ($1 for postgres or ?1 for sqlite)
	res, err := dbu.Select("where f1 = %s", 1)
	if err != nil {
		panic(err)
	}
	fmt.Println(res)

	// count all
	count, err := dbu.Count("where true=true")
	if err != nil {
		panic(err)
	}
	fmt.Println(count)

	// update
	res.F1 = 3
	err = dbu.Update(res, "where f1 = %s", 2)
	if err != nil {
		panic(err)
	}

	// delete
	err = dbu.Delete("where f1 = %s", 3)
	if err != nil {
		panic(err)
	}
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
	res, err := dbu.Select("where f1 = %s", 1)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, 1, res.F1)

	// select single non-existent
	_, err = dbu.Select("where f1 = %s", 0)
	assert.Error(t, err)
	assert.ErrorIs(t, sql.ErrNoRows, err)

	// select multi
	list, err := dbu.SelectMulti("where true=true")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(list))

	// select multi (no results)
	list, err = dbu.SelectMulti("where f1 = %s", -1)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(list))

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
	tbl, err := dbu.Select("where f1 = %s", 2)
	assert.NoError(t, err)
	assert.NotNil(t, tbl)
	tbl.F1 = 3
	err = dbu.Update(tbl, "where f1 = %s", 2)
	assert.NoError(t, err)

	// count specific
	count, err = dbu.Count("where f1 = %s", 3)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	// delete
	err = dbu.Delete("where f1 = %s", 3)
	assert.NoError(t, err)

	// count specific (after delete)
	count, err = dbu.Count("where f1 = %s", 3)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)

}

func TestBindParams(t *testing.T) {
	dbu := dbutil.New[*MyTable](nil, dbutil.DialectSQLite, nil)
	str := dbu.FormatBindParams("where x = %s and y = %s", 2)
	assert.Equal(t, "where x = ?1 and y = ?2", str)

	dbu = dbutil.New[*MyTable](nil, dbutil.DialectPostgres, nil)
	str = dbu.FormatBindParams("where x = %s and y = %s", 2)
	assert.Equal(t, "where x = $1 and y = $2", str)
}
