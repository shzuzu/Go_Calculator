package calc_test

import (
	"testing"

	"github.com/shzuzu/Go_Calculator/pkg/calc"
)

func TestCals(t *testing.T) {
	testSucces := []struct {
		name           string
		expression     string
		expectedResult float64
	}{
		{
			name:           "simple",
			expression:     "1+1",
			expectedResult: 2,
		},
		{
			name:           "priority",
			expression:     "(2+2)*2",
			expectedResult: 8,
		},
		{
			name:           "priority",
			expression:     "2+2*2",
			expectedResult: 6,
		},
		{
			name:           "/",
			expression:     "1/2",
			expectedResult: 0.5,
		},
	}
	for _, tc := range testSucces {
		t.Run(tc.name, func(t *testing.T) {
			value, err := calc.Calc(tc.expression)
			if err != nil {
				t.Fatalf("something went wrong with succesful case: %s", tc.expression)
			}
			if value != tc.expectedResult {
				t.Fatalf("Expected %f, but got %f", tc.expectedResult, value)
			}
		})
	}
	testFail := []struct {
		name        string
		expression  string
		expectedErr error
	}{
		{
			name:       "simple",
			expression: "1+1*",
		},
		{
			name:       "priority",
			expression: "2+2**2",
		},
		{
			name:       "priority",
			expression: "((2+2-*(2",
		},
		{
			name:       "/",
			expression: "",
		},
	}
	for _, tc := range testFail {
		t.Run(tc.name, func(t *testing.T) {
			value, err := calc.Calc(tc.expression)
			if err == nil {
				t.Fatalf("expression %s is invalid, but got result %f", tc.expression, value)
			}
		})
	}
}
