package dbutil_test

import (
	"database/sql"
	"os"
	"path"
	"testing"

	"github.com/minetest-go/dbutil"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func setupDB(t assert.TestingT) *sql.DB {
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

	// prepare insert
	ps, err := dbu.PrepareInsert()
	assert.NoError(t, err)
	assert.NotNil(t, ps)
	assert.NoError(t, ps(&MyTable{F1: 66}))

	// select prepared insert
	res, err = dbu.Select("where f1 = %s", 66)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, 66, res.F1)

}

func TestBindParams(t *testing.T) {
	dbu := dbutil.New[*MyTable](nil, dbutil.DialectSQLite, nil)
	str := dbu.FormatBindParams("where x = %s and y = %s", 2)
	assert.Equal(t, "where x = ?1 and y = ?2", str)

	dbu = dbutil.New[*MyTable](nil, dbutil.DialectPostgres, nil)
	str = dbu.FormatBindParams("where x = %s and y = %s", 2)
	assert.Equal(t, "where x = $1 and y = $2", str)
}

func BenchmarkInsert(b *testing.B) {
	// setup
	db := setupDB(b)
	provider := func() *MyTable { return &MyTable{} }
	dbu := dbutil.New(db, dbutil.DialectSQLite, provider)

	// insert
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		assert.NoError(b, dbu.Insert(&MyTable{F1: 1}))
	}
}

func BenchmarkInsertTx(b *testing.B) {
	// setup
	db := setupDB(b)
	provider := func() *MyTable { return &MyTable{} }
	tx, err := db.Begin()
	assert.NoError(b, err)
	assert.NotNil(b, tx)

	dbu := dbutil.New(tx, dbutil.DialectSQLite, provider)

	// insert
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		assert.NoError(b, dbu.Insert(&MyTable{F1: 1}))
	}

	assert.NoError(b, tx.Commit())
}

func BenchmarkInsertPrepared(b *testing.B) {
	// setup
	db := setupDB(b)
	provider := func() *MyTable { return &MyTable{} }
	dbu := dbutil.New(db, dbutil.DialectSQLite, provider)

	// prepare insert
	ps, err := dbu.PrepareInsert()
	assert.NoError(b, err)
	assert.NotNil(b, ps)
	assert.NoError(b, ps(&MyTable{F1: 66}))

	// insert
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		assert.NoError(b, ps(&MyTable{F1: 1}))
	}
}

func BenchmarkInsertPreparedTx(b *testing.B) {
	// setup
	db := setupDB(b)
	provider := func() *MyTable { return &MyTable{} }
	tx, err := db.Begin()
	assert.NoError(b, err)
	assert.NotNil(b, tx)

	dbu := dbutil.New(tx, dbutil.DialectSQLite, provider)

	// prepare insert
	ps, err := dbu.PrepareInsert()
	assert.NoError(b, err)
	assert.NotNil(b, ps)
	assert.NoError(b, ps(&MyTable{F1: 66}))

	// insert
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		assert.NoError(b, ps(&MyTable{F1: 1}))
	}

	assert.NoError(b, tx.Commit())
}
