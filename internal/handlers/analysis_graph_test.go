package handlers

import (
	"crypto-analytics/internal/models"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockAnalysisService struct {
	Response *models.AnalysisData
	Error    error
}

func (m *MockAnalysisService) GetPairInfo(pair, timeframe string) (*models.AnalysisData, error) {
	return m.Response, m.Error
}

func TestHandler_GetPairInfo_Simple(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		mockResponse   *models.AnalysisData
		mockError      error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "successful request",
			queryParams:    "?pair=BTCUSDT&timeframe=1h",
			mockResponse:   &models.AnalysisData{Pair: "BTCUSDT", Timeframe: "1h"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing pair parameter",
			queryParams:    "?timeframe=1h",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Параметры pair и timeframe обязательны\n",
		},
		{
			name:           "missing timeframe parameter",
			queryParams:    "?pair=BTCUSDT",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Параметры pair и timeframe обязательны\n",
		},
		{
			name:           "service returns error",
			queryParams:    "?pair=UNKNOWN&timeframe=1h",
			mockError:      fmt.Errorf("данные для пары UNKNOWN и таймфрейма 1h не найдены"),
			expectedStatus: http.StatusNotFound,
			expectedBody:   "данные для пары UNKNOWN и таймфрейма 1h не найдены\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockAnalysisService{
				Response: tt.mockResponse,
				Error:    tt.mockError,
			}

			h := &Handler{Analysis: mockService}
			req := httptest.NewRequest("GET", "/pair-info"+tt.queryParams, nil)
			rr := httptest.NewRecorder()

			h.GetPairInfo(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			if tt.expectedBody != "" {
				if rr.Body.String() != tt.expectedBody {
					t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), tt.expectedBody)
				}
			}
		})
	}
}
