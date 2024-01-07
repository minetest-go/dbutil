package dbutil_test

import (
	"database/sql"
	"fmt"

	"github.com/minetest-go/dbutil"
)

// database struct
type MyTable struct {
	PK *int64 // optional primary key (autoincremented)
	F1 int    // simple field
}

// entity
func (t *MyTable) Table() string {
	return "mytable"
}

// all column names
func (t *MyTable) Columns(action string) []string {
	return []string{"pk", "f1"}
}

// database -> struct
func (t *MyTable) Scan(action string, r func(dest ...any) error) error {
	return r(&t.PK, &t.F1)
}

// struct -> database
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
