package goscanql

import (
	"fmt"

	"github.com/DATA-DOG/go-sqlmock"
)

const (
	exampleQuery = `
		SELECT
			user.id AS id,
			user.name AS name,
			user.date_of_birth AS date_of_birth,
			vehicle.medium AS vehicle_medium,
			vehicle.type AS vehicle_type,
			vehicle.colour AS vehicle_colour,
			vehicle.noise AS vehicle_noise
		FROM user LEFT JOIN vehicle ON user.id = vehicle.user_id;`
)

// User represents an example user struct that you might want to parse data into
type User struct {
	Id       int       `goscanql:"id"`
	Name     string    `goscanql:"name"`
	Vehicles []Vehicle `goscanql:"vehicle"`
}

// Vehicle represents an example vehicle struct that you might want to parse data into
type Vehicle struct {
	Medium string `goscanql:"medium"`
	Type   string `goscanql:"type"`
	Colour string `goscanql:"colour"`
	Noise  string `goscanql:"noise"`
}

func ExampleRowsToStructs() {

	// setup the example to allow with mock data
	db, mock, err := sqlmock.New()
	if err != nil {
		panic(err)
	}

	columns := []string{"id", "name", "vehicle_medium", "vehicle_type", "vehicle_colour", "vehicle_noise"}
	inputRows := sqlmock.NewRows(columns)

	inputRows.AddRow(1, "Stirling Archer", "land", "car", "black", "brum")
	inputRows.AddRow(2, "Cheryl Tunt", "air", "aeroplane", "white", "whoosh")
	inputRows.AddRow(3, "Algernop Krieger", "land", "van", "blue", "brum")
	inputRows.AddRow(3, "Algernop Krieger", "sea", "submarine", "black", "...")
	inputRows.AddRow(4, "Barry Dylan", "space", "spaceship", "grey", "RRRRRRRRRRRRRRRRRRGGHHHH")
	inputRows.AddRow(4, "Barry Dylan", "land", "motorbike", "black", "vroom")

	mock.ExpectQuery(exampleQuery).WillReturnRows(inputRows)

	rows, err := db.Query(exampleQuery)
	if err != nil {
		panic(err)
	}

	// Execute the RowsToStructs from goscanql
	result, err := RowsToStructs[User](rows)
	fmt.Println(fmt.Sprintf("FINAL %p", result[2].Vehicles))
	if err != nil {
		panic(err)
	}

	// Output: goscanql.User{Id:3, Name:"Algernop Krieger", Vehicles:[]goscanql.Vehicle{goscanql.Vehicle{Medium:"land", Type:"van", Colour:"blue", Noise:"brum"}}}
	fmt.Printf("%#v", result[2])
}
