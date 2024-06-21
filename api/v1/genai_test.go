package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock dependencies
type MockAnalyzer struct {
	mock.Mock
}

func (m *MockAnalyzer) Analyze(ctx context.Context, content string) (string, error) {
	args := m.Called(ctx, content)
	return args.String(0), args.Error(1)
}

type MockFeedback struct {
	mock.Mock
}

func (m *MockFeedback) Store(feedback interface{}) error {
	args := m.Called(feedback)
	return args.Error(0)
}

func TestAnalyzeFailures_Success(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	// Initialize the service with mock dependencies
	mockAnalyzer := new(MockAnalyzer)

	// Initialize the service
	service := NewGenAIService(mockAnalyzer, nil)
	service.RegisterHandlers(router)

	// Define the expected behavior of the mock
	mockAnalyzer.On("Analyze", mock.Anything, mock.Anything).Return("mock analysis", nil)

	// Prepare the request
	requestBody, err := json.Marshal(AnalyzeFailuresRequest{
		Failures: []*Failure{{Context: "test failure"}},
	})
	require.NoError(t, err, "Failed to marshal request body")

	req, err := http.NewRequest(http.MethodPost, "/analyze", bytes.NewBuffer(requestBody))
	require.NoError(t, err, "Failed to create request")

	w := httptest.NewRecorder()

	// Invoke the handler
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAnalyzeFailures_ParseError(t *testing.T) {
	// Set Gin to Test Mode
	gin.SetMode(gin.TestMode)

	// Create a Gin router
	router := gin.Default()

	// Initialize the service with mock dependencies
	mockAnalyzer := new(MockAnalyzer)

	// Initialize the service
	service := NewGenAIService(mockAnalyzer, nil)

	// Register the handlers for the service
	service.RegisterHandlers(router)

	// Prepare an invalid request (e.g., an improperly formatted JSON)
	invalidRequestBody := bytes.NewBufferString(`{"failures": "invalid"}`)
	req, _ := http.NewRequest(http.MethodPost, "/analyze", invalidRequestBody)
	w := httptest.NewRecorder()

	// Invoke the handler
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAnalyzeFailures_AnalysisError(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	// Initialize the service with mock dependencies
	mockAnalyzer := new(MockAnalyzer)

	// Initialize the service
	service := NewGenAIService(mockAnalyzer, nil)
	service.RegisterHandlers(router)

	mockAnalyzer.On("Analyze", mock.Anything, mock.Anything).Return("", errors.New("analysis error"))

	// Prepare a valid request that should trigger an analysis error
	validRequestBody, _ := json.Marshal(AnalyzeFailuresRequest{
		Failures: []*Failure{{Context: "test failure"}},
	})
	req, _ := http.NewRequest(http.MethodPost, "/analyze", bytes.NewBuffer(validRequestBody))
	w := httptest.NewRecorder()

	// Invoke the handler
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestFeedback_SuccessFeedbackForFailure(t *testing.T) {
	// Set Gin to Test Mode
	gin.SetMode(gin.TestMode)

	// Create a Gin router
	router := gin.Default()

	// Initialize the mock dependencies
	mockFeedback := new(MockFeedback)

	// Create an instance of the service with the mocks
	service := NewGenAIService(nil, mockFeedback)
	service.RegisterHandlers(router)

	// Define the expected behavior of the mock
	mockFeedback.On("Store", mock.Anything).Return(nil)

	// Prepare the request
	requestBody, err := json.Marshal(UserFeedbackRequestFailure{
		UserFeedbackRequest: UserFeedbackRequest{
			FeedbackFor: FeedbackForFailure,
			Positive:    true,
		},
	})
	require.NoError(t, err, "Failed to marshal request body")

	req, err := http.NewRequest(http.MethodPost, "/feedback", bytes.NewBuffer(requestBody))
	require.NoError(t, err, "Failed to create request")

	w := httptest.NewRecorder()

	// Invoke the handler
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestFeedback_ShouldBindBodyWithError(t *testing.T) {
	// Set Gin to Test Mode
	gin.SetMode(gin.TestMode)

	// Create a Gin router
	router := gin.Default()

	// Initialize the mock dependencies
	mockFeedback := new(MockFeedback)

	// Create an instance of the service with the mocks
	service := NewGenAIService(nil, mockFeedback)
	service.RegisterHandlers(router)

	// Define the expected behavior of the mock
	mockFeedback.On("Store", mock.Anything).Return(nil)

	// Prepare the request
	req, err := http.NewRequest(http.MethodPost, "/feedback", strings.NewReader("not a json body"))
	require.NoError(t, err, "Failed to create request")

	w := httptest.NewRecorder()

	// Invoke the handler
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFeedback_UnsupportedFeedbackFor(t *testing.T) {
	// Set Gin to Test Mode
	gin.SetMode(gin.TestMode)

	// Create a Gin router
	router := gin.Default()

	// Initialize the mock dependencies
	mockFeedback := new(MockFeedback)

	// Create an instance of the service with the mocks
	service := NewGenAIService(nil, mockFeedback)
	service.RegisterHandlers(router)

	// Define the expected behavior of the mock
	mockFeedback.On("Store", mock.Anything).Return(nil)

	// Prepare the request
	requestBody, err := json.Marshal(UserFeedbackRequest{
		FeedbackFor: "unsupported",
	})
	require.NoError(t, err, "Failed to marshal request body")

	req, err := http.NewRequest(http.MethodPost, "/feedback", bytes.NewBuffer(requestBody))
	require.NoError(t, err, "Failed to create request")

	w := httptest.NewRecorder()

	// Invoke the handler
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFeedback_StoreError(t *testing.T) {
	// Set Gin to Test Mode
	gin.SetMode(gin.TestMode)

	// Create a Gin router
	router := gin.Default()

	// Initialize the mock dependencies
	mockFeedback := new(MockFeedback)

	// Create an instance of the service with the mocks
	service := NewGenAIService(nil, mockFeedback)
	service.RegisterHandlers(router)

	// Define the expected behavior of the mock
	mockFeedback.On("Store", mock.Anything).Return(errors.New("store error"))

	// Prepare the request
	requestBody, err := json.Marshal(UserFeedbackRequest{
		FeedbackFor: FeedbackForFailure,
	})
	require.NoError(t, err, "Failed to marshal request body")

	req, err := http.NewRequest(http.MethodPost, "/feedback", bytes.NewBuffer(requestBody))
	require.NoError(t, err, "Failed to create request")

	w := httptest.NewRecorder()

	// Invoke the handler
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestFeedback_ShouldNotAcceptV2Body(t *testing.T) {
	// Set Gin to Test Mode
	gin.SetMode(gin.TestMode)

	// Create a Gin router
	router := gin.Default()

	// Initialize the mock dependencies
	mockFeedback := new(MockFeedback)

	// Create an instance of the service with the mocks
	service := NewGenAIService(nil, mockFeedback)
	service.RegisterHandlers(router)

	// Define the expected behavior of the mock
	mockFeedback.On("Store", mock.Anything).Return(nil)

	// Prepare the request
	req, err := http.NewRequest(http.MethodPost, "/feedback", strings.NewReader(
		`{
			"buildUrl": "http://localhost:8000/jenkins/job/ibp-metrics-agent/view/change-requests/job/PR-13/30/",
			"comments": "",
			"failureAnalysis": "analysis",
			"failureContext": "context",
			"failureResolution": "resolution",
			"feedbackFor": "failure",
			"pipelineStepsNodeUrl": "http://localhost:8000/jenkins/job/ibp-metrics-agent/job/PR-13/30/execution/node/15/log/",
			"positive": "true"",
			"userId": "admin"
    }`))
	require.NoError(t, err, "Failed to create request")

	w := httptest.NewRecorder()

	// Invoke the handler
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
