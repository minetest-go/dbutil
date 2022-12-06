
Golang database utilities and helpers

![](https://github.com/minetest-go/dbutil/workflows/test/badge.svg)
[![Coverage Status](https://coveralls.io/repos/github/minetest-go/dbutil/badge.svg)](https://coveralls.io/github/minetest-go/dbutil)

# Example


Type definition
```go
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
```

Queries
```go
// insert
dbutil.Insert(db, &MyTable{F1: 1})

// insert with return value
var retVal int64 = -1
dbutil.InsertReturning(db, &MyTable{F1: 2}, "pk", &retVal)

// select single
res, err := dbutil.Select(db, &MyTable{}, "where f1 = $1", 1)
//NOTE: err may be of type "sql.ErrNoRows" if no rows were found

// select multi
list, err := dbutil.SelectMulti(db, func() *MyTable { return &MyTable{} }, "where true=true")

// count
count, err := dbutil.Count(db, &MyTable{}, "where true=true")

// update (where f2 = 2)
err := dbutil.Update(db, &MyTable{F1: 3}, map[string]any{"f1": 2})

// delete
err := dbutil.Delete(db, &MyTable{}, "where f1 = $1", 3)
```

# License

Code: **MIT**