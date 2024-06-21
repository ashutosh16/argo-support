package genai

import (
	"github.intuit.com/dev-build/ibp-genai-service/internal/splunk"
)

type Feedback interface {
	Store(feedbackData interface{}) error
}

type UserFeedback struct {
	splunkClient splunk.Client
}

func NewUserFeedback(splunkClient splunk.Client) *UserFeedback {
	return &UserFeedback{
		splunkClient: splunkClient,
	}
}

func (f *UserFeedback) Store(feedbackData interface{}) error {
	// Publish event
	return f.splunkClient.Publish(feedbackData, splunk.SplunkSourceTypeFeedback)
}
