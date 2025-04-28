package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"time-guard-bot/internal/models"
)

func TestSendJSON(t *testing.T) {
	t.Run("Successful JSON Response", func(t *testing.T) {
		testData := struct {
			Name  string `json:"name"`
			Value int    `json:"value"`
		}{
			Name:  "test",
			Value: 123,
		}

		rec := httptest.NewRecorder()

		sendJSON(rec, testData)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
		}

		contentType := rec.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type %s, got %s", "application/json", contentType)
		}

		var response struct {
			Name  string `json:"name"`
			Value int    `json:"value"`
		}

		if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.Name != testData.Name {
			t.Errorf("Expected name %s, got %s", testData.Name, response.Name)
		}

		if response.Value != testData.Value {
			t.Errorf("Expected value %d, got %d", testData.Value, response.Value)
		}
	})
}

func TestSendJSONError(t *testing.T) {
	testCases := []struct {
		name       string
		message    string
		statusCode int
	}{
		{
			name:       "Not Found Error",
			message:    "Resource not found",
			statusCode: http.StatusNotFound,
		},
		{
			name:       "Bad Request Error",
			message:    "Invalid parameters",
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "Internal Server Error",
			message:    "Internal server error",
			statusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			sendJSONError(rec, tc.message, tc.statusCode)

			if rec.Code != tc.statusCode {
				t.Errorf("Expected status code %d, got %d", tc.statusCode, rec.Code)
			}

			contentType := rec.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Expected Content-Type %s, got %s", "application/json", contentType)
			}

			var response models.ErrorResponse
			if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if response.Error != tc.message {
				t.Errorf("Expected error message %s, got %s", tc.message, response.Error)
			}
		})
	}
}
