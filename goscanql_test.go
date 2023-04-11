package goscanql

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidateType(t *testing.T) {

	tests := []struct {
		name        string
		run         func() error
		expectedErr error
	}{
		{
			name: "Validate Struct",
			run: func() error {
				return validateType[struct{}]()
			},
			expectedErr: nil,
		},
		{
			name: "Validate Struct Pointer",
			run: func() error {
				return validateType[*struct{}]()
			},
			expectedErr: nil,
		},
		{
			name: "Validate Struct Pointer Pointer",
			run: func() error {
				return validateType[**struct{}]()
			},
			expectedErr: nil,
		},
		{
			name: "Validate Non-Struct",
			run: func() error {
				return validateType[string]()
			},
			expectedErr: fmt.Errorf("unexpected type of (string) - type should be a struct or pointer to a struct"),
		},
		{
			name: "Validate Non-Struct Pointer",
			run: func() error {
				return validateType[*int]()
			},
			expectedErr: fmt.Errorf("unexpected type of (*int) - type should be a struct or pointer to a struct"),
		},
		{
			name: "Validate Non-Struct Pointer Pointer",
			run: func() error {
				return validateType[**float64]()
			},
			expectedErr: fmt.Errorf("unexpected type of (**float64) - type should be a struct or pointer to a struct"),
		},
	}

	for _, test := range tests {
		msg := fmt.Sprintf("%s: failed", test.name)
		assert.Equalf(t, test.expectedErr, test.run(), msg)
	}
}
