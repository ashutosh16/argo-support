package genai

import (
	"github.intuit.com/dev-build/ibp-genai-service/internal/splunk"
)

type Activity interface {
	Store(data interface{}) error
}

type UserActivity struct {
	splunkClient splunk.Client
}

func NewUserActivity(splunkClient splunk.Client) *UserActivity {
	return &UserActivity{
		splunkClient: splunkClient,
	}
}

func (f *UserActivity) Store(data interface{}) error {
	return f.splunkClient.Publish(data, splunk.SplunkSourceTypeActivity)
}
