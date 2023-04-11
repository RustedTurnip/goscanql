package main

import "time"

// Account represents a type of parent entity you might expect to use when
// using goscanql.
type Account struct {
	ID          int64     `goscanql:"id"`
	Email       string    `goscanql:"email"`
	DateOfBirth time.Time `goscanql:"date_of_birth"`
	CreatedAt   time.Time `goscanql:"created_at"`
	Pets        []*Pet    `goscanql:"pets"`
}

// Colour represents a type of parent (and child) entity you might expect
// to use when using goscanql.
type Colour struct {
	Name  string `goscanql:"name"`
	Red   int64  `goscanql:"red"`
	Green int64  `goscanql:"green"`
	Blue  int64  `goscanql:"blue"`
}

// Pet represents a type of (child) entity you might expect to use when
// using goscanql.
type Pet struct {
	ID     int64   `goscanql:"id"`
	Name   string  `goscanql:"name"`
	Animal string  `goscanql:"animal"`
	Breed  string  `goscanql:"breed"`
	Colour *Colour `goscanql:"colour"`
}
