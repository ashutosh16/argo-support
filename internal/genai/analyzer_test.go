package genai

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.intuit.com/dev-build/ibp-genai-service/internal/genai/express"
)

type MockGenAIClientForAnalyzer struct {
	mock.Mock
}

func (m *MockGenAIClientForAnalyzer) SubmitWithRetries(ctx context.Context, request express.ChatCompletionJSONRequestBody) (*express.ExpressModeResponse, error) {
	args := m.Called(ctx, request)

	// Get the response
	var response *express.ExpressModeResponse
	if args.Get(0) != nil {
		response = args.Get(0).(*express.ExpressModeResponse)
	}

	return response, args.Error(1)
}

func (m *MockGenAIClientForAnalyzer) Submit(ctx context.Context, request express.ChatCompletionJSONRequestBody) (*express.ExpressModeResponse, error) {
	//TODO implement me
	panic("implement me")
}

func TestAnalyze_Success(t *testing.T) {
	mockClient := new(MockGenAIClientForAnalyzer)
	analyzer := NewFailureAnalyzer(mockClient)
	content := "build failure context"

	mockResponse := &express.ExpressModeResponse{
		Answer: express.SuccessAnswerInExpressMode{
			Content: "analysis result"},
	}

	mockClient.On("SubmitWithRetries", mock.Anything, mock.Anything).
		Return(mockResponse, nil).
		Once()

	ctx := context.Background()
	response, err := analyzer.Analyze(ctx, content)

	assert.NoError(t, err)
	assert.Equal(t, "analysis result", response)
}

func TestAnalyze_ClientError(t *testing.T) {
	mockClient := new(MockGenAIClientForAnalyzer)
	analyzer := NewFailureAnalyzer(mockClient)
	content := "build failure context"

	mockError := errors.New("client error")

	mockClient.On("SubmitWithRetries", mock.Anything, mock.Anything).
		Return(nil, mockError).
		Once()

	ctx := context.Background()
	response, err := analyzer.Analyze(ctx, content)

	assert.Error(t, err)
	assert.Equal(t, mockError.Error(), err.Error())
	assert.Equal(t, "", response)
}
