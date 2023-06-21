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
			user_alias.alias AS alias,
			vehicle.type AS vehicle_type,
			vehicle.colour AS vehicle_colour,
			vehicle.noise AS vehicle_noise,
			vehicle_medium.name AS vehicle_medium_name
		FROM user
		LEFT JOIN user_alias ON user.id=user_alias.user_id
		LEFT JOIN vehicle ON user.id = vehicle.user_id
        LEFT JOIN vehicle_medium ON vehicle.medium_id=vehicle_medium.id;`
)

// User represents an example user struct that you might want to parse data into
type User struct {
	Id       int       `goscanql:"id"`
	Name     string    `goscanql:"name"`
	Vehicles []Vehicle `goscanql:"vehicle"`
	Aliases  []string  `goscanql:"alias"`
}

// Vehicle represents an example vehicle struct that you might want to parse data into
type Vehicle struct {
	Type    string          `goscanql:"type"`
	Colour  string          `goscanql:"colour"`
	Noise   string          `goscanql:"noise"`
	Mediums []VehicleMedium `goscanql:"medium"`
}

// VehicleMedium represents the "medium" upon which a vehicle operates
type VehicleMedium struct {
	Name string `goscanql:"name"`
}

func ExampleRowsToStructs() {

	// setup the example to allow with mock data
	db, mock, err := sqlmock.New()
	if err != nil {
		panic(err)
	}

	columns := []string{"id", "name", "alias", "vehicle_type", "vehicle_colour", "vehicle_noise", "vehicle_medium_name"}
	inputRows := sqlmock.NewRows(columns)

	inputRows.AddRow(1, "Stirling Archer", "", "car", "black", "brum", "land")
	inputRows.AddRow(2, "Cheryl Tunt", "Chrystal", "aeroplane", "white", "whoosh", "air")
	inputRows.AddRow(2, "Cheryl Tunt", "Charlene", "aeroplane", "white", "whoosh", "air")
	inputRows.AddRow(3, "Algernop Krieger", "", "van", "blue", "brum", "land")
	inputRows.AddRow(3, "Algernop Krieger", "", "submarine", "black", "...", "sea")
	inputRows.AddRow(3, "Algernop Krieger", "", "submarine", "black", "...", "swimming pool")
	inputRows.AddRow(4, "Barry Dylan", "", "spaceship", "grey", "RRRRRRRRRRRRRRRRRRGGHHHH", "space")
	inputRows.AddRow(4, "Barry Dylan", "", "motorbike", "black", "vroom", "land")
	inputRows.AddRow(5, "Pam Poovey", nil, "motorbike", "black", "vroom", "land")
	inputRows.AddRow(5, "Pam Poovey", nil, nil, nil, nil, nil)

	mock.ExpectQuery(exampleQuery).WillReturnRows(inputRows)

	rows, err := db.Query(exampleQuery)
	if err != nil {
		panic(err)
	}

	// Execute the RowsToStructs from goscanql
	result, err := RowsToStructs[User](rows)
	if err != nil {
		panic(err)
	}

	// Output: goscanql.User{Id:3, Name:"Algernop Krieger", Vehicles:[]goscanql.Vehicle{goscanql.Vehicle{Medium:"land", Type:"van", Colour:"blue", Noise:"brum"}}}
	fmt.Printf("%#v", result[2])
}
