package goscanql

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

const (
	scanTestQuery = `
		SELECT
			user.id AS id,
			user.name AS name,
			user.date_of_birth AS date_of_birth,
			user_alias.alias AS alias,
			user_role.title AS role_title,
			user_role.department AS role_department,
			vehicle.type AS vehicle_type,
			vehicle.colour AS vehicle_colour,
			vehicle.noise AS vehicle_noise,
			vehicle_medium.name AS vehicle_medium_name
		FROM user
		LEFT JOIN user_alias ON user.id=user_alias.user_id
		LEFT JOIN user_role ON user.role_id=user_role.id
		LEFT JOIN vehicle ON user.id = vehicle.user_id
        LEFT JOIN vehicle_medium ON vehicle.medium_id=vehicle_medium.id;`
)

// User represents an example user struct that you might want to parse data into
type TestUser struct {
	Id       int           `goscanql:"id"`
	Name     string        `goscanql:"name"`
	Vehicles []TestVehicle `goscanql:"vehicle"`
	Aliases  []string      `goscanql:"alias"`
	Role     *TestRole     `goscanql:"role"`
}

// Role represents the User's position in their organisation, carrying with it any
// relevant attributes
type TestRole struct {
	Title      string `goscanql:"title"`
	Department string `goscanql:"department"`
}

// Vehicle represents an example vehicle struct that you might want to parse data into
type TestVehicle struct {
	Type    string              `goscanql:"type"`
	Colour  string              `goscanql:"colour"`
	Noise   string              `goscanql:"noise"`
	Mediums []TestVehicleMedium `goscanql:"medium"`
}

// VehicleMedium represents the "medium" upon which a vehicle operates
type TestVehicleMedium struct {
	Name string `goscanql:"name"`
}

func Test_ExampleRowsToStructs(t *testing.T) {

	// Arrange
	// setup the example to allow with mock data
	db, mock, err := sqlmock.New()
	if err != nil {
		panic(err)
	}

	columns := []string{"id", "name", "role_title", "role_department", "alias", "vehicle_type", "vehicle_colour", "vehicle_noise", "vehicle_medium_name"}
	inputRows := sqlmock.NewRows(columns)

	inputRows.AddRow(1, "Stirling Archer", "field agent", "field operations", "", "car", "black", "brum", "land")
	inputRows.AddRow(2, "Cheryl Tunt", "secretary", "", "Chrystal", "aeroplane", "white", "whoosh", "air")
	inputRows.AddRow(2, "Cheryl Tunt", "secretary", "", "Charlene", "aeroplane", "white", "whoosh", "air")
	inputRows.AddRow(3, "Algernop Krieger", "lab geek", "research & development", "", "van", "blue", "brum", "land")
	inputRows.AddRow(3, "Algernop Krieger", "lab geek", "research & development", "", "submarine", "black", "...", "sea")
	inputRows.AddRow(3, "Algernop Krieger", "lab geek", "research & development", "", "submarine", "black", "...", "swimming pool")
	inputRows.AddRow(4, "Barry Dylan", nil, nil, "", "spaceship", "grey", "RRRRRRRRRRRRRRRRRRGGHHHH", "space")
	inputRows.AddRow(4, "Barry Dylan", nil, nil, "", "motorbike", "black", "vroom", "land")
	inputRows.AddRow(5, "Pam Poovey", "hr manager", "human resources", nil, "motorbike", "black", "vroom", "land")
	inputRows.AddRow(5, "Pam Poovey", "hr manager", "human resources", nil, nil, nil, nil, nil)

	mock.ExpectQuery(exampleQuery).WillReturnRows(inputRows)

	rows, err := db.Query(exampleQuery)
	if err != nil {
		panic(err)
	}

	// Act
	// Execute the RowsToStructs from goscanql
	result, err := RowsToStructs[TestUser](rows)

	// Assert
	assert.Nil(t, err)
	assert.Equal(t, expectedUsers, result)
}

var (
	expectedUsers = []TestUser{
		{
			Id:   1,
			Name: "Stirling Archer",
			Vehicles: []TestVehicle{
				{
					Type:   "car",
					Colour: "black",
					Noise:  "brum",
					Mediums: []TestVehicleMedium{
						{
							Name: "land",
						},
					},
				},
			},
			Aliases: []string{
				"",
			},
			Role: &TestRole{
				Title:      "field agent",
				Department: "field operations",
			},
		},
		{
			Id:   2,
			Name: "Cheryl Tunt",
			Vehicles: []TestVehicle{
				{
					Type:   "aeroplane",
					Colour: "white",
					Noise:  "whoosh",
					Mediums: []TestVehicleMedium{
						{
							Name: "air",
						},
					},
				},
			},
			Aliases: []string{
				"Chrystal",
				"Charlene",
			},
			Role: &TestRole{
				Title:      "secretary",
				Department: "",
			},
		},
		{
			Id:   3,
			Name: "Algernop Krieger",
			Vehicles: []TestVehicle{
				{
					Type:   "van",
					Colour: "blue",
					Noise:  "brum",
					Mediums: []TestVehicleMedium{
						{
							Name: "land",
						},
					},
				},
				{
					Type:   "submarine",
					Colour: "black",
					Noise:  "...",
					Mediums: []TestVehicleMedium{
						{
							Name: "sea",
						},
						{
							Name: "swimming pool",
						},
					},
				},
			},
			Aliases: []string{
				"",
			},
			Role: &TestRole{
				Title:      "lab geek",
				Department: "research & development",
			},
		},
		{
			Id:   4,
			Name: "Barry Dylan",
			Vehicles: []TestVehicle{
				{
					Type:   "spaceship",
					Colour: "grey",
					Noise:  "RRRRRRRRRRRRRRRRRRGGHHHH",
					Mediums: []TestVehicleMedium{
						{
							Name: "space",
						},
					},
				},
				{
					Type:   "motorbike",
					Colour: "black",
					Noise:  "vroom",
					Mediums: []TestVehicleMedium{
						{
							Name: "land",
						},
					},
				},
			},
			Aliases: []string{
				"",
			},
			Role: nil,
		},
		{
			Id:   5,
			Name: "Pam Poovey",
			Vehicles: []TestVehicle{
				{
					Type:   "motorbike",
					Colour: "black",
					Noise:  "vroom",
					Mediums: []TestVehicleMedium{
						{
							Name: "land",
						},
					},
				},
			},
			Aliases: nil,
			Role: &TestRole{
				Title:      "hr manager",
				Department: "human resources",
			},
		},
	}
)
