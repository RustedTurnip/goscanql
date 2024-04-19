package goscanql

import (
	"fmt"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

const (
	exampleQuery = `
		SELECT
			user.id AS id,
			user.name AS name,
			user.date_of_birth AS date_of_birth,
			user.nemesis AS nemesis,
			user.catchphrase AS catchphrase,
			vehicle.medium AS vehicle_medium,
			vehicle.type AS vehicle_type,
			vehicle.colour AS vehicle_colour,
			vehicle.noise AS vehicle_noise
		FROM user LEFT JOIN vehicle ON user.id = vehicle.user_id;`
)

// User represents an example user struct that you might want to parse data into
type User struct {
	Id          int         `sql:"id"`
	Name        string      `sql:"name"`
	DateOfBirth time.Time   `sql:"date_of_birth"`
	Nemesis     NullString  `sql:"nemesis"`
	Catchphrase interface{} `sql:"catchphrase"`
	Vehicles    []Vehicle   `sql:"vehicle"`
}

// Vehicle represents an example vehicle struct that you might want to parse data into
type Vehicle struct {
	Medium string `sql:"medium"`
	Type   string `sql:"type"`
	Colour string `sql:"colour"`
	Noise  string `sql:"noise"`
}

func ExampleRowsToStructs() {
	// setup the example to allow with mock data
	db, mock, err := sqlmock.New()
	if err != nil {
		panic(err)
	}

	columns := []string{"id", "name", "date_of_birth", "nemesis", "catchphrase", "vehicle_medium", "vehicle_type", "vehicle_colour", "vehicle_noise"}
	inputRows := sqlmock.NewRows(columns)

	inputRows.AddRow(1, "Stirling Archer", time.Date(1978, 12, 30, 0, 0, 0, 0, time.UTC), "Barry Dylan", "Danger Zone.", "land", "car", "black", "brum")
	inputRows.AddRow(2, "Cheryl Tunt", time.Date(1987, 4, 24, 0, 0, 0, 0, time.UTC), "Cecil Tunt", nil, "air", "aeroplane", "white", "whoosh")
	inputRows.AddRow(3, "Algernop Krieger", time.Date(1977, 9, 24, 0, 0, 0, 0, time.UTC), nil, "Yep Yep Yep!", "land", "van", "blue", "brum")
	inputRows.AddRow(3, "Algernop Krieger", time.Date(1977, 9, 24, 0, 0, 0, 0, time.UTC), nil, "Yep Yep Yep!", "sea", "submarine", "black", "...")
	inputRows.AddRow(4, "Barry Dylan", time.Date(1984, 6, 19, 0, 0, 0, 0, time.UTC), "Stirling Archer", nil, "space", "spaceship", "grey", "RRRRRRRRRRRRRRRRRRGGHHHH")
	inputRows.AddRow(4, "Barry Dylan", time.Date(1984, 6, 19, 0, 0, 0, 0, time.UTC), "Stirling Archer", nil, "land", "motorbike", "black", "vroom")

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

	// Output: goscanql.User{Id:3, Name:"Algernop Krieger", DateOfBirth:time.Date(1977, time.September, 24, 0, 0, 0, 0, time.UTC), Nemesis:goscanql.NullString{String:"", Valid:false}, Catchphrase:"Yep Yep Yep!", Vehicles:[]goscanql.Vehicle{goscanql.Vehicle{Medium:"land", Type:"van", Colour:"blue", Noise:"brum"}, goscanql.Vehicle{Medium:"sea", Type:"submarine", Colour:"black", Noise:"..."}}}
	fmt.Printf("%#v", result[2])
}
