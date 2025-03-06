package application_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/shzuzu/Go_Calculator/internal/application"
)

func TestCreateExpressionHandler(t *testing.T) {
	orchestrator := application.NewOrchestrator(4)

	tt := []struct {
		name           string
		expression     string
		expectedStatus int
		expectedBody   []byte
	}{
		{
			name:           "Valid Expression",
			expression:     "2+2*2",
			expectedStatus: http.StatusCreated,
			expectedBody:   []byte(`{"id":"1"}`),
		},
		{
			name:           "Invalid Expression",
			expression:     "2(+()",
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   []byte(`{"error":"Expression is not valid"}`),
		},
		{
			name:           "Division by Zero",
			expression:     "1/0",
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   []byte(`{"error":"Division by zero"}`),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			requestBody, _ := json.Marshal(map[string]string{"expression": tc.expression})
			req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", bytes.NewReader(requestBody))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			orchestrator.CreateExpressionHandler(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("Expected status code %d, but got %d", tc.expectedStatus, rr.Code)
			}

			if !JSONBytesEqual(rr.Body.Bytes(), tc.expectedBody) {
				t.Errorf("Expected body %s, but got %s", string(tc.expectedBody), rr.Body.String())
			}
		})
	}
}

func TestGetExpressionsHandler(t *testing.T) {
	orchestrator := application.NewOrchestrator(4)
	orchestrator.Exprs = []application.Expression{
		{Id: "1", Status: "done", Result: func() *float64 { f := 6.0; return &f }()},
		{Id: "2", Status: "pending", Result: nil},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/expressions", nil)
	rr := httptest.NewRecorder()

	orchestrator.GetExpressionsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, rr.Code)
	}

	expectedBody := `{"expressions":[{"id":"1","status":"done","result":6},{"id":"2","status":"pending","result":null}]}`
	if !JSONBytesEqual(rr.Body.Bytes(), []byte(expectedBody)) {
		t.Errorf("Expected body %s, but got %s", expectedBody, rr.Body.String())
	}
}

func TestExpressionFromID(t *testing.T) {
	orchestrator := application.NewOrchestrator(4)
	orchestrator.Exprs = []application.Expression{
		{Id: "1", Status: "done", Result: func() *float64 { f := 10.0; return &f }()},
		{Id: "2", Status: "pending", Result: nil},
	}

	tt := []struct {
		name           string
		id             string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Valid ID",
			id:             "1",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":"1","status":"done","result":10}`,
		},
		{
			name:           "Invalid ID",
			id:             "999",
			expectedStatus: http.StatusNotFound,
			expectedBody:   `Expression with ID 999 not found`,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/expressions/"+tc.id, nil)
			req.SetPathValue("id", tc.id)

			rr := httptest.NewRecorder()

			orchestrator.ExpressionFromID(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("Expected status code %d, but got %d", tc.expectedStatus, rr.Code)
			}

			// убираю лишние пробелы и переносы строк
			body := strings.TrimSpace(rr.Body.String())
			expectedBody := strings.TrimSpace(tc.expectedBody)

			if body != expectedBody {
				t.Errorf("Expected body `%s`, but got `%s`", expectedBody, body)
			}
		})
	}
}

func TestExpressionFromID_InternalServerError(t *testing.T) {
	orchestrator := application.NewOrchestrator(4)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/expressions/1", nil)
	req.SetPathValue("id", "1")

	rr := httptest.NewRecorder()

	orchestrator.ExpressionFromID(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, but got %d", http.StatusNotFound, rr.Code)
	}

	body := strings.TrimSpace(rr.Body.String())
	expectedBody := strings.TrimSpace(`Expression with ID 1 not found`)

	if body != expectedBody {
		t.Errorf("Expected body `%s`, but got `%s`", expectedBody, body)
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
