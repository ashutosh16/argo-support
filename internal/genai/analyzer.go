package genai

import (
	"context"

	"github.intuit.com/dev-build/ibp-genai-service/internal/genai/express"
)

type Analyzer interface {
	Analyze(ctx context.Context, content string) (string, error)
}

type FailureAnalyzer struct {
	client Client
}

func NewFailureAnalyzer(client Client) *FailureAnalyzer {
	return &FailureAnalyzer{
		client: client,
	}
}

func (a *FailureAnalyzer) Analyze(ctx context.Context, content string) (string, error) {
	// Build prompt
	prompt := "Follow the instructions provide in the context, provide summary in markdown format. "

	// Build request
	request := express.ChatCompletionJSONRequestBody{
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
					Role:    stringPtr(MessageRoleSystem),
					Content: stringPtr(prompt),
				},
				{
					Role:    stringPtr(MessageRoleUser),
					Content: stringPtr(content),
				},
			},
		},
	}

	// Set timeout for request
	timeoutCtx, cancel := context.WithTimeout(ctx, DefaultTimeOut) // TODO: Read timeout from config
	defer cancel()

	// Set type for context
	ctxWithType := context.WithValue(timeoutCtx, typeKey, "Analyzer")

	// Submit request
	response, err := a.client.SubmitWithRetries(ctxWithType, request)
	if err != nil {
		return "", err
	}

	return response.Answer.Content, nil
}

func float64Ptr(v float64) *float64 {
	return &v
}

func intPtr(v int) *int {
	return &v
}

func stringPtr(v string) *string {
	return &v
}
