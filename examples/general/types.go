package main

import "time"

// Account represents a type of parent entity you might expect to use when
// using goscanql.
type Account struct {
	ID          int64     `sql:"id"`
	Email       string    `sql:"email"`
	DateOfBirth time.Time `sql:"date_of_birth"`
	CreatedAt   time.Time `sql:"created_at"`
	Pets        []*Pet    `sql:"pets"`
}

// Colour represents a type of parent (and child) entity you might expect
// to use when using goscanql.
type Colour struct {
	Name  string `sql:"name"`
	Red   int64  `sql:"red"`
	Green int64  `sql:"green"`
	Blue  int64  `sql:"blue"`
}

// Pet represents a type of (child) entity you might expect to use when
// using goscanql.
type Pet struct {
	ID     int64   `sql:"id"`
	Name   string  `sql:"name"`
	Animal string  `sql:"animal"`
	Breed  string  `sql:"breed"`
	Colour *Colour `sql:"colour"`
}
