# goscanql

[![Go Reference](https://pkg.go.dev/badge/github.com/rustedturnip/goscanql.svg)](https://pkg.go.dev/github.com/rustedturnip/goscanql)

`goscanql` is a library to supplement sql operations in Go. It allows you to layout a struct (using tags) that an 
`sql.Rows` response from querying a database can be mapped to.



## Example

```go
type User struct {
	Id       int64  `sql:"id"`
	Name     string `sql:"name"`
	Username string `sql:"username"`
}

rows, err := db.Query('SELECT * FROM users')
if err != nil {
	panic(err)
}

users, err := goscanql.RowsToStructs[*User](rows)
...
```



## Scanner Interface

If a field implements the `goscanql.Scanner` interface, then the SQL value will be passed directly into the field
(whether it is a primitive type or not). For example:

```go
type User struct {
	Id       goscanql.NullInt64 `sql:"id"`
	Name     string             `sql:"name"`
	Username string             `sql:"username"`
}

rows, err := db.Query('SELECT * FROM users')
if err != nil {
	panic(err)
}

users, err := goscanql.RowsToStructs[*User](rows)
...
```

As `goscanql.NullInt64` implements the scanner interface, the value of the sql query under the column `id` will be
passed directly into the `sql.NullInt64` struct (whereas otherwise, the `sql.NullInt` struct would have been analysed
for sub-fields that have `sql` tags).



## SQL Joins

This library is particularly useful in aggregating data resulting from SQL joins as it can aggregate parents by 
common elements and append children with differing elements to child slices, for example:

```go
type User struct {
	Id       int64    `sql:"id"`
	Name     string   `sql:"name"`
	Username string   `sql:"username"`
	Aliases  []string `sql:"aliases"`
	Pets     []Pet    `sql:"pets"`
}

type Pet struct {
	Animal string  `sql:"animal"`
	Name   string  `sql:"name"`
	Colour *Colour `sql:"colour"`
}

type Colour struct {
	Red   int8 `sql:"red"`
	Green int8 `sql:"green"`
	Blue  int8 `sql:"blue"`
}

rows, err := db.Query('
    SELECT
        user.id          AS id, 
		user.name        AS name,
		user.username    AS username,
		user_alias.alias AS aliases,
		pet.animal       AS pets_anmial,
		pet.name         AS pets_name,
		colour.red       AS pets_colour_red,
		colour.green     AS pets_colour_green,
		colour.blue      AS pets_colour_blue
    FROM users
    LEFT JOIN user_alias ON user.id = user_alias.user_id
    LEFT JOIN pet        ON user.id = pet.user_id
    LEFT JOIN colour     ON pet.id = colour.pet_id')
if err != nil {
	panic(err)
}

users, err := goscanql.RowsToStructs[*User](rows)
...
```

In the example above, you can see how the gocanql package handles composite structs where slices are used to hold 
children of a struct, (e.g. how `[]Pets` represents entries of the pet table, and how a user can have multiple pets).

When working with nested structs, the field in the SQL query must be aliased to show that it belongs as a child of 
that struct by prefixing the alias with the `sql` tag of the parent, for instance in the example above, every 
field that belongs to the `Pets` struct is prefixed with:

- `pets_`

And every instance that belongs to a `Colour` struct (that in turn belongs to the `Pets` struct) is prefixed with:

- `pets_colour_`


### Aggregation

#### One-to-Many

A `one-to-many` relationship is indicated by using a slice as a field type. 

Currently, aggregation works by merging all fields that match in a `one-to-many` relationship. For example, in the 
example at the top of this "SQL Joins" section, all Users that have the same:

- `id`
- `name`
- `username`

Will be treated as the same user, but when any of these fields differ, a new `*User` will be appended to the slice 
of users. The same can be said for any children of the `User` struct also.

Where two `aliases` for the user match, the will be treated as the same and will only be added to the `Aliases` 
of `User` field once.

#### One-to-One

Where a one-to-one relationship exists, the fields of the sub-struct will be treated as an extension of the parent. 
For example, in the `Pet` to `Colour` relationship (where one pet can have one colour), if all of the `Pet` fields 
match, but any of the `Colour` fields differ, they will be treated as two different pets.



## ByteSlice

If you have a column in your database with a type that effectively translates to a byte slice in go (`[]byte`) then
by having that type directly in a struct, you may notice the undesirable behaviour of `goscanql` treating that type
instead as a one-to-many relationship of single bytes, for example:

```go
type User struct {
	ID  int    `sql:"id"`
	Pin []byte `sql:"pin"`
}
```

`goscanql` would expect in this case, that the column `pin` is actually of type `byte` and expect that a single user
could have many `pins`, each being a single `byte`, when actually a `pin` is a single value consisting of multiple
`bytes`.

To work around this issue, `goscanql` provides a type called `ByteSlice` that overrides the default "scan" behaviour
and treats those `bytes` as the single value they were intended to be. To use this, you would simply change the type
of the field `Pin` in the user example above to use `ByteSlice` like so:

```go
type User struct {
	ID  int       `sql:"id"`
	Pin ByteSlice `sql:"pin"`
}
```

`ByteSlice` has a base type of `[]byte`, meaning that it can be used in the same way.



## Limitations

### Unsupported fields

The following field types are not supported:
- Arrays
- Maps
- Multi-dimensional slices

### Cyclic Structs

It should also be noted that currently, cyclic structs are also not supported as they will cause an infinite loop, 
for example:

```go
type User struct {
	Id       int64   `sql:"id"`
	Name     string  `sql:"name"`
	Username string  `sql:"username"`
	Friends  []*User `sql:"friends"`
}
```

Will not work as the package will recursively add `friends` of a `User` to each user and will not know when to stop. 
A more elegant workaround will be implemented in future versions of `goscanql`.

