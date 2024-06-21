package genai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.intuit.com/dev-build/ibp-genai-service/internal/genai/express"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockGenOSClient is a mock for ClientWithResponsesInterface
type MockGenOSClient struct {
	mock.Mock
}

func (m *MockGenOSClient) ChatCompletionWithBodyWithResponse(ctx context.Context, params *express.ChatCompletionParams, contentType string, body io.Reader, reqEditors ...express.RequestEditorFn) (*express.ChatCompletionResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockGenOSClient) ChatCompletionWithResponse(ctx context.Context, params *express.ChatCompletionParams, body express.ChatCompletionJSONRequestBody, reqEditors ...express.RequestEditorFn) (*express.ChatCompletionResponse, error) {
	args := m.Called(ctx, body, reqEditors)

	// Get the response
	var response *express.ChatCompletionResponse
	if args.Get(0) != nil {
		response = args.Get(0).(*express.ChatCompletionResponse)
	}

	return response, args.Error(1)
}

func (m *MockGenOSClient) CodeGenerationWithBodyWithResponse(ctx context.Context, params *express.CodeGenerationParams, contentType string, body io.Reader, reqEditors ...express.RequestEditorFn) (*express.CodeGenerationResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockGenOSClient) CodeGenerationWithResponse(ctx context.Context, params *express.CodeGenerationParams, body express.CodeGenerationJSONRequestBody, reqEditors ...express.RequestEditorFn) (*express.CodeGenerationResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockGenOSClient) TextGenerationWithBodyWithResponse(ctx context.Context, params *express.TextGenerationParams, contentType string, body io.Reader, reqEditors ...express.RequestEditorFn) (*express.TextGenerationResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockGenOSClient) TextGenerationWithResponse(ctx context.Context, params *express.TextGenerationParams, body express.TextGenerationJSONRequestBody, reqEditors ...express.RequestEditorFn) (*express.TextGenerationResponse, error) {
	//TODO implement me
	panic("implement me")
}

func newMockGenOSClient(mockClient *MockGenOSClient) *GenAI {
	return &GenAI{client: mockClient, sanitizer: NewSanitizer()}
}

func getTestAppSecret() string {
	return "appSecret"
}

func TestNewClient_Success(t *testing.T) {
	// Call NewClient
	client, err := NewClient("baseURL", "identityServiceURL", "appID", "jobID", getTestAppSecret)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, client)
}

func TestSubmit_Success(t *testing.T) {
	// Set up the mock client
	mockClient := new(MockGenOSClient)

	// Create an instance of GenAI with the mock client
	genAI := newMockGenOSClient(mockClient)

	// Prepare the test data
	content := "Example user message for debugging"

	requestBody := express.ChatCompletionJSONRequestBody{
		ConversationId: stringPtr("NEW"),
		LlmParams: express.LLMParamsMessage{
			LlmConfiguration: express.LLMConfiguration{
				MaxTokens:   intPtr(512),
				Model:       ModelGTP4,
				Temperature: float64Ptr(0.7),
				TopP:        float64Ptr(1),
			},
			Messages: []express.Message{
				{
					Role: stringPtr(MessageRoleSystem),
					Content: stringPtr("You are an expert at debugging failures. Please help with this error and follow this structure: " +
						"A first part where you summarize the error. " +
						"A second part where you explain how to fix the error with step-by-step instructions. " +
						"Finally, separate the 2 parts with this text: `-/-/-/-` as a delimiter for parsing."),
				},
				{
					Role:    stringPtr(MessageRoleUser),
					Content: stringPtr(content),
				},
			},
		},
	}

	// Define the expected response
	expectedResponse := &express.ExpressModeResponse{
		Answer: express.SuccessAnswerInExpressMode{
			Content: "Error Summary: The issue is related to a null pointer exception.\n" +
				"Fix Instructions:\n" +
				"1. Check if the object is initialized before use.\n" +
				"2. Ensure all required fields of the object are set.\n" +
				"-/-/-/-\n" +
				"Remember to test the solution thoroughly."},
	}

	// Marshal the expected response to JSON
	expectedJSON, err := json.Marshal(expectedResponse)
	require.NoError(t, err, "Failed to marshal expected response")

	// Mock the success response
	mockResponse := &express.ChatCompletionResponse{
		HTTPResponse: &http.Response{StatusCode: 200},
		Body:         expectedJSON,
	}

	// Test the successful submission
	mockClient.On("ChatCompletionWithResponse", mock.Anything, requestBody, mock.Anything).Return(mockResponse, nil)

	// Call the method
	ctx := context.Background()
	response, err := genAI.Submit(ctx, requestBody)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, response)

	// Verify that the mock was called
	responseJSON, err := json.Marshal(response)
	assert.NoError(t, err)
	assert.JSONEq(t, string(expectedJSON), string(responseJSON))

	// Verify that the mock was called
	mockClient.AssertExpectations(t)
}

func TestSubmit_NetworkError(t *testing.T) {
	ctx := context.Background()
	requestBody := express.ChatCompletionJSONRequestBody{}

	mockClient := new(MockGenOSClient)
	genAI := newMockGenOSClient(mockClient)

	mockClient.On("ChatCompletionWithResponse", mock.Anything, requestBody, mock.Anything).Return(nil, fmt.Errorf("network error"))
	_, err := genAI.Submit(ctx, requestBody)
	require.Error(t, err, "Expected an error for network issues")
}

func TestSubmit_HTTPError(t *testing.T) {
	ctx := context.Background()
	requestBody := express.ChatCompletionJSONRequestBody{}

	mockClient := new(MockGenOSClient)
	genAI := newMockGenOSClient(mockClient)

	mockResponse := &express.ChatCompletionResponse{
		HTTPResponse: &http.Response{StatusCode: 400},
		Body:         []byte{}, // Mock body
	}
	mockClient.On("ChatCompletionWithResponse", mock.Anything, requestBody, mock.Anything).Return(mockResponse, nil)
	_, err := genAI.Submit(ctx, requestBody)
	require.Error(t, err, "Expected an error for HTTP response error")
}

func TestSubmit_JSONUnmarshalError(t *testing.T) {
	ctx := context.Background()
	requestBody := express.ChatCompletionJSONRequestBody{}

	mockClient := new(MockGenOSClient)
	genAI := newMockGenOSClient(mockClient)

	invalidJSON := []byte("invalid json")
	mockResponse := &express.ChatCompletionResponse{
		HTTPResponse: &http.Response{StatusCode: 200},
		Body:         invalidJSON,
	}
	mockClient.On("ChatCompletionWithResponse", mock.Anything, requestBody, mock.Anything).Return(mockResponse, nil)
	_, err := genAI.Submit(ctx, requestBody)
	require.Error(t, err, "Expected an error for JSON unmarshalling")
}

func TestSubmitWithRetries_Success(t *testing.T) {
	requestBody := express.ChatCompletionJSONRequestBody{}

	mockClient := new(MockGenOSClient)
	genAI := newMockGenOSClient(mockClient)

	// Define the expected response
	expectedResponse := &express.ExpressModeResponse{
		Answer: express.SuccessAnswerInExpressMode{
			Content: "Error Summary: The issue is related to a null pointer exception.\n" +
				"Fix Instructions:\n" +
				"1. Check if the object is initialized before use.\n" +
				"2. Ensure all required fields of the object are set.\n" +
				"-/-/-/-\n" +
				"Remember to test the solution thoroughly."},
	}

	// Marshal the expected response to JSON
	expectedJSON, err := json.Marshal(expectedResponse)
	require.NoError(t, err, "Failed to marshal expected response")

	// Mock the success response
	mockResponse := &express.ChatCompletionResponse{
		HTTPResponse: &http.Response{StatusCode: 200},
		Body:         expectedJSON,
	}

	// Test the successful submission
	mockClient.On("ChatCompletionWithResponse", mock.Anything, requestBody, mock.Anything).Return(mockResponse, nil)

	// Call the method
	ctx := context.Background()
	response, err := genAI.SubmitWithRetries(ctx, requestBody)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, response)

	// Verify that the mock was called
	responseJSON, err := json.Marshal(response)
	assert.NoError(t, err)
	assert.JSONEq(t, string(expectedJSON), string(responseJSON))

	// Verify that the mock was called
	mockClient.AssertExpectations(t)
}

func TestSubmitWithRetries_ContextCancelled(t *testing.T) {
	// Create a context and cancel it immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	requestBody := express.ChatCompletionJSONRequestBody{ /* ... */ }

	mockClient := new(MockGenOSClient)
	genAI := newMockGenOSClient(mockClient)

	// Expect the Submit method to be called, but it won't complete due to the context timeout
	mockClient.On("ChatCompletionWithResponse", mock.Anything, requestBody, mock.Anything).Return(nil, fmt.Errorf("this error should not be seen due to context timeout"))

	// Call the method
	_, err := genAI.SubmitWithRetries(ctx, requestBody)

	// Assertions
	require.Error(t, err)
	assert.Contains(t, err.Error(), "operation cancelled or timed out")

	// Since context is cancelled, SubmitWithRetries should not successfully call the underlying method
	mockClient.AssertNotCalled(t, "ChatCompletionWithResponse", mock.Anything, requestBody, mock.Anything)
}

func TestSubmitWithRetries_SubmitError(t *testing.T) {
	// Create a context with a very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel() // Ensure resources are cleaned up

	requestBody := express.ChatCompletionJSONRequestBody{}

	mockClient := new(MockGenOSClient)
	genAI := newMockGenOSClient(mockClient)

	mockClient.On("ChatCompletionWithResponse", mock.Anything, requestBody, mock.Anything).Return(nil, fmt.Errorf("network error"))

	_, err := genAI.SubmitWithRetries(ctx, requestBody)
	require.Error(t, err, "Expected an error for network issues")
}

func TestGetAuthorizationHeader_Success(t *testing.T) {
	// Setup
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	identityServiceURL := "http://example.com/identity"
	appID := "testAppID"
	jobID := "testJobID"

	// Mock the expected request and response
	expectedRequestBody := `{"query":"mutation identitySignInInternalApplicationWithPrivateAuth($input: Identity_SignInApplicationWithPrivateAuthInput!) {  identitySignInInternalApplicationWithPrivateAuth(input: $input) { authorizationHeader }}","variables":{"input":{"profileId":"testJobID"}}}`
	expectedAuthorizationHeader := "testAuthHeader"
	httpmock.RegisterResponder("POST", identityServiceURL,
		func(req *http.Request) (*http.Response, error) {
			// Check request headers
			authHeader := req.Header.Get("Authorization")
			contentType := req.Header.Get("Content-Type")
			assert.Equal(t, "Intuit_IAM_Authentication intuit_appid=testAppID, intuit_app_secret="+getTestAppSecret(), authHeader)
			assert.Equal(t, "application/json", contentType)

			// Read and assert the body of the request
			body, _ := io.ReadAll(req.Body)
			assert.Equal(t, expectedRequestBody, string(body))

			// Return a mock response
			resp := `{"data":{"identitySignInInternalApplicationWithPrivateAuth":{"authorizationHeader":"` + expectedAuthorizationHeader + `"}}}`
			return httpmock.NewStringResponse(200, resp), nil
		},
	)

	// Execute the method to be tested
	actualAuthorizationHeader, err := getAuthorizationHeaderFromIdentityService(identityServiceURL, appID, jobID, getTestAppSecret)

	// Asserts
	assert.NoError(t, err)
	assert.Equal(t, "testAuthHeader,intuit_appid=testAppID,intuit_app_secret="+getTestAppSecret(), actualAuthorizationHeader)
}

func TestGetAuthorizationHeader_RequestCreationFailure(t *testing.T) {
	_, err := getAuthorizationHeaderFromIdentityService("://bad-url", "appID", "jobID", getTestAppSecret)
	require.Error(t, err)
}

func TestGetAuthorizationHeader_ClientDoError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Register a responder that returns an error
	httpmock.RegisterResponder("POST", "http://example.com",
		func(req *http.Request) (*http.Response, error) {
			return nil, fmt.Errorf("forced error in client.Do")
		},
	)
	_, err := getAuthorizationHeaderFromIdentityService("http://example.com", "appID", "jobID", getTestAppSecret)
	require.Error(t, err)
}

func TestGetAuthorizationHeader_HTTPErrorStatus(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "http://example.com", httpmock.NewStringResponder(400, ""))

	_, err := getAuthorizationHeaderFromIdentityService("http://example.com", "appID", "jobID", getTestAppSecret)
	require.Error(t, err)
}

func TestGetAuthorizationHeader_JSONUnmarshalError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "http://example.com", httpmock.NewStringResponder(200, "invalid JSON"))

	_, err := getAuthorizationHeaderFromIdentityService("http://example.com", "appID", "jobID", getTestAppSecret)
	require.Error(t, err)
}

func TestGetAuthorizationHeader_EmptyResponse(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "http://example.com",
		httpmock.NewStringResponder(200, ""))

	_, err := getAuthorizationHeaderFromIdentityService("http://example.com", "appID", "jobID", getTestAppSecret)
	require.Error(t, err)
}
