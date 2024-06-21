package genai

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.intuit.com/dev-build/ibp-genai-service/internal/splunk"
)

func TestActivityStore_Success(t *testing.T) {
	mockClient := new(MockSplunkClient)
	userActivity := NewUserActivity(mockClient)
	activityData := "test activity"

	mockClient.On("Publish", activityData, splunk.SplunkSourceTypeActivity).
		Return(nil).
		Once()

	err := userActivity.Store(activityData)

	assert.NoError(t, err)
}

func TestActivityStore_ClientError(t *testing.T) {
	mockClient := new(MockSplunkClient)
	userActivity := NewUserActivity(mockClient)
	activityData := "test activity"
	mockError := errors.New("client error")

	mockClient.On("Publish", activityData, splunk.SplunkSourceTypeActivity).
		Return(mockError).
		Once()

	err := userActivity.Store(activityData)

	assert.Error(t, err)
	assert.Equal(t, mockError.Error(), err.Error())
}
