package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-bot/internal/api/apierror"
	"go-bot/internal/database"
	"go-bot/internal/services/mocks"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAdminHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)
	silentLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	jwtSecret := "test-secret-for-jwt-token-generation"

	testCases := []struct {
		name         string
		payload      gin.H
		mockSetup    func(*mocks.AdminServiceInterface)
		expectedCode int
		checkToken   bool
	}{
		{
			name:    "Successful Login",
			payload: gin.H{"login": "admin", "password": "password"},
			mockSetup: func(mockService *mocks.AdminServiceInterface) {
				mockService.On("Authenticate", mock.Anything, "admin", "password").
					Return(&database.Admin{ID: 1, Login: "admin"}, nil).Once()
			},
			expectedCode: http.StatusOK,
			checkToken:   true,
		},
		{
			name:    "Invalid Credentials",
			payload: gin.H{"login": "admin", "password": "wrongpassword"},
			mockSetup: func(mockService *mocks.AdminServiceInterface) {
				mockService.On("Authenticate", mock.Anything, "admin", "wrongpassword").
					Return(nil, errors.New("invalid credentials")).Once()
			},
			expectedCode: http.StatusUnauthorized,
			checkToken:   false,
		},
		{
			name:    "Validation Error - Short Password",
			payload: gin.H{"login": "admin", "password": "123"},
			mockSetup: func(mockService *mocks.AdminServiceInterface) {
				// Authenticate should not be called
			},
			expectedCode: http.StatusBadRequest,
			checkToken:   false,
		},
		{
			name:         "Invalid JSON",
			payload:      nil, // Will be replaced by invalid body
			mockSetup:    func(mockService *mocks.AdminServiceInterface) {},
			expectedCode: http.StatusBadRequest,
			checkToken:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := new(mocks.AdminServiceInterface)
			tc.mockSetup(mockService)

			h := NewAdminHandler(mockService, silentLogger, jwtSecret)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			var reqBody io.Reader
			if tc.name == "Invalid JSON" {
				reqBody = bytes.NewBufferString(`{"login":"admin"`) // Malformed JSON
			} else {
				jsonPayload, _ := json.Marshal(tc.payload)
				reqBody = bytes.NewBuffer(jsonPayload)
			}

			c.Request, _ = http.NewRequest(http.MethodPost, "/login", reqBody)
			c.Request.Header.Set("Content-Type", "application/json")

			apierror.ErrorWrapper(h.Login)(c)

			assert.Equal(t, tc.expectedCode, w.Code)

			if tc.checkToken {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response["token"])
			}

			mockService.AssertExpectations(t)
		})
	}
}
