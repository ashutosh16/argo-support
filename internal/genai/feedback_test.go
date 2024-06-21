package genai

import (
	"errors"
	"github.intuit.com/dev-build/ibp-genai-service/internal/splunk"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSplunkClient is a mock of splunk.Client
type MockSplunkClient struct {
	mock.Mock
}

func (m *MockSplunkClient) Publish(data interface{}, sourceType string) error {
	args := m.Called(data, sourceType)
	return args.Error(0)
}

func TestStore_Success(t *testing.T) {
	mockClient := new(MockSplunkClient)
	userFeedback := NewUserFeedback(mockClient)
	feedbackData := "test feedback"

	mockClient.On("Publish", feedbackData, splunk.SplunkSourceTypeFeedback).
		Return(nil).
		Once()

	err := userFeedback.Store(feedbackData)

	assert.NoError(t, err)
}

func TestStore_ClientError(t *testing.T) {
	mockClient := new(MockSplunkClient)
	userFeedback := NewUserFeedback(mockClient)
	feedbackData := "test feedback"
	mockError := errors.New("client error")

	mockClient.On("Publish", feedbackData, splunk.SplunkSourceTypeFeedback).
		Return(mockError).
		Once()

	err := userFeedback.Store(feedbackData)

	assert.Error(t, err)
	assert.Equal(t, mockError.Error(), err.Error())
}
