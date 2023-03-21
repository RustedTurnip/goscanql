# goscanql

***Note:** This project is subject to change. Until v1, changes may be backwards incompatible.*

`goscanql` is a library to supplement sql operations in Go. It allows you to layout a struct (using tags)
that an `sql.Rows` response from querying a database can be mapped to.

## Example

```go
type User struct {
	Id       int64  `goscanql:"id"`
	Name     string `goscanql:"name"`
	Username string `goscanql:"username"`
}

rows, err := db.Query('SELECT * FROM users')
if err != nil {
	panic(err)
}

users, err := goscanql.RowsToStructs[*User](rows)
...
```


