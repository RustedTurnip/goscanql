package main

import "time"

type Account struct {
	Id          int64     `goscanql:"id"`
	Email       string    `goscanql:"email"`
	DateOfBirth time.Time `goscanql:"date_of_birth"`
	CreatedAt   time.Time `goscanql:"created_at"`
	Pets        []*Pet    `goscanql:"pets"`
}

type Colour struct {
	Name  string `goscanql:"name"`
	Red   int64  `goscanql:"red"`
	Green int64  `goscanql:"green"`
	Blue  int64  `goscanql:"blue"`
}

type Pet struct {
	Id     int64   `goscanql:"id"`
	Name   string  `goscanql:"name"`
	Animal string  `goscanql:"animal"`
	Breed  string  `goscanql:"breed"`
	Colour *Colour `goscanql:"colour"`
}
