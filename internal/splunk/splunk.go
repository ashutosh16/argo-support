package splunk

import (
	"time"

	"github.com/ZachtimusPrime/Go-Splunk-HTTP/splunk/v2"
	"github.com/rs/zerolog/log"
)

const (
	SplunkIndex              = "ibp_genai"
	SplunkSourceGenAIService = "ibp-genai-service"
	SplunkSourceTypeFeedback = "user-feedback"
	SplunkSourceTypeActivity = "activity"
)

// Client defines the interface for GenAI client methods
type Client interface {
	Publish(data interface{}, sourceType string) error
}

type Splunk struct {
	client *splunk.Client
	env    string
}

func NewSplunk(url, token, env string) *Splunk {
	return &Splunk{
		client: splunk.NewClient(nil, url, token, SplunkSourceGenAIService, "", SplunkIndex),
		env:    env,
	}
}

func (s *Splunk) Publish(data interface{}, sourceType string) error {
	// Set source and source type
	source := SplunkSourceGenAIService

	// If we're in dev, publish to other source and sourceType in order to avoid polluting prod
	if s.env == "dev" {
		source = source + "-dev"
		sourceType = sourceType + "-dev"
	}

	// Create event
	event := s.client.NewEventWithTime(time.Now(), data, source, sourceType, SplunkIndex)

	log.Info().Msgf("Publishing event to Splunk with source: %s, sourceType: %s, index: %s", source, sourceType, SplunkIndex)

	// Log event
	return s.client.LogEvent(event)
}
