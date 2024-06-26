package goscanql

import (
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

const (
	scanTestQuery = `
		SELECT
			user.id AS id,
			user.name AS name,
			user.office_access_pin AS office_access_pin,
			user.characteristics AS characteristics,
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

// TestUserCharacteristics represents a Scanner type that has custom "Scan" behaviour.
// In this instance, it demonstrates how you might scan a string and parse it into a
// slice, which goscanql couldn't do on its own.
type TestUserCharacteristics []string

func (c *TestUserCharacteristics) Scan(b interface{}) error {
	if b == nil {
		return nil
	}

	*c = strings.Split(b.(string), ",")
	return nil
}

func (c *TestUserCharacteristics) ID() []byte {
	return []byte(strings.Join(*c, ","))
}

// User represents an example user struct that you might want to parse data into
type TestUser struct {
	Id              int                     `sql:"id"`
	Name            string                  `sql:"name"`
	OfficeAccessPin ByteSlice               `sql:"office_access_pin"`
	Characteristics TestUserCharacteristics `sql:"characteristics"`
	DateOfBirth     NullTime                `sql:"date_of_birth"`
	Vehicles        []TestVehicle           `sql:"vehicle"`
	Aliases         []string                `sql:"alias"`
	Role            *TestRole               `sql:"role"`
}

// Role represents the User's position in their organisation, carrying with it any
// relevant attributes
type TestRole struct {
	Title      string `sql:"title"`
	Department string `sql:"department"`
}

// Vehicle represents an example vehicle struct that you might want to parse data into
type TestVehicle struct {
	Type    string              `sql:"type"`
	Colour  string              `sql:"colour"`
	Noise   string              `sql:"noise"`
	Mediums []TestVehicleMedium `sql:"medium"`
}

// VehicleMedium represents the "medium" upon which a vehicle operates
type TestVehicleMedium struct {
	Name string `sql:"name"`
}

func Test_ExampleRowsToStructs(t *testing.T) {
	// Arrange
	// setup the example to allow with mock data
	db, mock, err := sqlmock.New()
	if err != nil {
		panic(err)
	}

	columns := []string{"id", "name", "office_access_pin", "characteristics", "date_of_birth", "role_title", "role_department", "alias", "vehicle_type", "vehicle_colour", "vehicle_noise", "vehicle_medium_name"}
	inputRows := sqlmock.NewRows(columns)

	inputRows.AddRow(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	inputRows.AddRow(1, "Stirling Archer", []byte("1234"), "narcissistic,arrogant,selfish,insensitive,self-absorbed,sex-crazed", time.Date(1978, 12, 30, 0, 0, 0, 0, time.UTC), "field agent", "field operations", "", "car", "black", "brum", "land")
	inputRows.AddRow(2, "Cheryl Tunt", []byte("9876"), "crazy", time.Date(1987, 4, 24, 0, 0, 0, 0, time.UTC), "secretary", "", "Chrystal", "aeroplane", "white", "whoosh", "air")
	inputRows.AddRow(2, "Cheryl Tunt", []byte("9876"), "crazy", time.Date(1987, 4, 24, 0, 0, 0, 0, time.UTC), "secretary", "", "Charlene", "aeroplane", "white", "whoosh", "air")
	inputRows.AddRow(3, "Algernop Krieger", []byte("3141"), nil, time.Date(1977, 9, 24, 0, 0, 0, 0, time.UTC), "lab geek", "research & development", "", "van", "blue", "brum", "land")
	inputRows.AddRow(3, "Algernop Krieger", []byte("3141"), nil, time.Date(1977, 9, 24, 0, 0, 0, 0, time.UTC), "lab geek", "research & development", "", "submarine", "black", "...", "sea")
	inputRows.AddRow(3, "Algernop Krieger", []byte("3141"), nil, time.Date(1977, 9, 24, 0, 0, 0, 0, time.UTC), "lab geek", "research & development", "", "submarine", "black", "...", "swimming pool")
	inputRows.AddRow(4, "Barry Dylan", nil, "bipolar", nil, nil, nil, "", "spaceship", "grey", "RRRRRRRRRRRRRRRRRRGGHHHH", "space")
	inputRows.AddRow(4, "Barry Dylan", nil, "bipolar", nil, nil, nil, nil, "motorbike", "black", "vroom", "land")
	inputRows.AddRow(5, "Pam Poovey", []byte{}, "inappropriate", nil, "hr manager", "human resources", nil, "motorbike", "black", "vroom", "land")
	inputRows.AddRow(5, "Pam Poovey", []byte{}, "inappropriate", nil, "hr manager", "human resources", nil, nil, nil, nil, nil)

	mock.ExpectQuery(scanTestQuery).WillReturnRows(inputRows)

	rows, err := db.Query(scanTestQuery)
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
			Id:              1,
			Name:            "Stirling Archer",
			OfficeAccessPin: ByteSlice{'1', '2', '3', '4'},
			Characteristics: TestUserCharacteristics{
				"narcissistic",
				"arrogant",
				"selfish",
				"insensitive",
				"self-absorbed",
				"sex-crazed",
			},
			DateOfBirth: NullTime{
				Time:  time.Date(1978, 12, 30, 0, 0, 0, 0, time.UTC),
				Valid: true,
			},
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

			OfficeAccessPin: ByteSlice{'9', '8', '7', '6'},
			Characteristics: TestUserCharacteristics{
				"crazy",
			},
			DateOfBirth: NullTime{
				Time:  time.Date(1987, 4, 24, 0, 0, 0, 0, time.UTC),
				Valid: true,
			},
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
			Id:              3,
			Name:            "Algernop Krieger",
			OfficeAccessPin: ByteSlice{'3', '1', '4', '1'},
			Characteristics: nil,
			DateOfBirth: NullTime{
				Time:  time.Date(1977, 9, 24, 0, 0, 0, 0, time.UTC),
				Valid: true,
			},
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
			Id:              4,
			Name:            "Barry Dylan",
			OfficeAccessPin: nil,
			Characteristics: TestUserCharacteristics{
				"bipolar",
			},
			DateOfBirth: NullTime{
				Time:  time.Time{},
				Valid: false,
			},
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
			Id:              5,
			Name:            "Pam Poovey",
			OfficeAccessPin: ByteSlice{},
			Characteristics: TestUserCharacteristics{
				"inappropriate",
			},
			DateOfBirth: NullTime{
				Time:  time.Time{},
				Valid: false,
			},
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

func Test_RecordListNilMapAssignment(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		panic(err)
	}

	columns := []string{"id", "name", "office_access_pin", "characteristics", "date_of_birth", "role_title", "role_department", "alias", "vehicle_type", "vehicle_colour", "vehicle_noise", "vehicle_medium_name"}
	inputRows := sqlmock.NewRows(columns)

	inputRows.AddRow(1, "Stirling Archer", []byte("1234"), "narcissistic,arrogant,selfish,insensitive,self-absorbed,sex-crazed", time.Date(1978, 12, 30, 0, 0, 0, 0, time.UTC), "field agent", "field operations", "", nil, nil, nil, nil)
	inputRows.AddRow(1, "Stirling Archer", []byte("1234"), "narcissistic,arrogant,selfish,insensitive,self-absorbed,sex-crazed", time.Date(1978, 12, 30, 0, 0, 0, 0, time.UTC), "field agent", "field operations", "", "car", "black", "brum", "land")

	mock.ExpectQuery(scanTestQuery).WillReturnRows(inputRows)

	rows, err := db.Query(scanTestQuery)
	if err != nil {
		panic(err)
	}

	expected := []TestUser{
		{
			Id:              1,
			Name:            "Stirling Archer",
			OfficeAccessPin: ByteSlice{'1', '2', '3', '4'},
			Characteristics: TestUserCharacteristics{"narcissistic", "arrogant", "selfish", "insensitive", "self-absorbed", "sex-crazed"},
			DateOfBirth: NullTime{
				Time:  time.Date(1978, 12, 30, 0, 0, 0, 0, time.UTC),
				Valid: true,
			},
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
	}

	// Act
	result, err := RowsToStructs[TestUser](rows)

	// Assert
	assert.Nil(t, err)
	assert.Equal(t, expected, result)
}
