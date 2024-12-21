package application_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	"net/http/httptest"

	"github.com/shzuzu/Go_Calculator/internal/application"
)

func TestCalcHandler(t *testing.T) {
	tt := []struct {
		name           string
		result         any
		expectedStatus int
		expectedBody   []byte
		expression     map[string]string
		responseType   string
	}{
		{
			name:           "OK",
			result:         6.0,
			expectedStatus: http.StatusOK,
			expectedBody:   []byte(`{"result":6}`),
			expression:     map[string]string{"expression": "2+2*2"},
			responseType:   "result",
		},
		{
			name:           "Expression error",
			result:         "Expression is not valid",
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   []byte(`{"error":"Expression is not valid"}`),
			expression:     map[string]string{"expression": "2(+()"},
			responseType:   "error",
		},
		{
			name:           "Expression error",
			result:         "Expression is not valid",
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   []byte(`{"error":"Expression is not valid"}`),
			expression:     map[string]string{"expression": "2(2*3)"},
			responseType:   "error",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			requestBodyBytes, _ := json.Marshal(tc.expression)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate",
				bytes.NewReader(requestBodyBytes))

			rr := httptest.NewRecorder()
			req.Header.Set("Content-Type", "application/json")

			application.CalcHandler(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("Expected status code %d, but got %d", tc.expectedStatus, rr.Code)
			}

			if !JSONBytesEqual(rr.Body.Bytes(), tc.expectedBody) {
				t.Errorf("Expected body %v, but got %v", string(tc.expectedBody), rr.Body.String())
			}
			var responseBody map[string]any
			err := json.Unmarshal(rr.Body.Bytes(), &responseBody)
			if err != nil {
				t.Fatal(err)
			}

			if responseBody[tc.responseType] != tc.result {
				t.Fatalf("Ожидаемый результат: %v, получено %v", tc.result, responseBody["result"])
			}

		})
	}

}

func JSONBytesEqual(a, b []byte) bool {
	var j, j2 interface{}
	if err := json.Unmarshal(a, &j); err != nil {
		return false
	}
	if err := json.Unmarshal(b, &j2); err != nil {
		return false
	}
	return reflect.DeepEqual(j2, j)
}
