package genai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"github.intuit.com/dev-build/ibp-genai-service/cmd/ibp-genai-service/config"
	"github.intuit.com/dev-build/ibp-genai-service/internal/genai/express"
)

const (
	MessageRoleAssistant = "assistant"
	MessageRoleSystem    = "system"
	MessageRoleUser      = "user"

	ModelGTP4        = "gpt-4"              // More capable than any GPT-3.5 model, able to do more complex tasks, and optimized for chat. Will be updated with our latest model iteration. 8,192 tokens.
	ModelGTP4With32k = "gpt-4-32k"          // Same capabilities as the base gpt-4 mode but with 4x the context length. Will be updated with our latest model iteration.					32,768 tokens.
	ModelGTP35Turbo  = "gpt-35-turbo-v0301" // Most capable GPT-3.5 model and optimized for chat at 1/10th the cost of text-davinci-003. Will be updated with our latest model iteration. 	4,096 tokens.

	DefaultTimeOut = 2 * time.Minute

	IntuitOriginatingAssetAliasHeader = "intuit_originating_assetalias"
)

type contextKeyType string

const typeKey contextKeyType = "requestType"

// Client defines the interface for GenAI client methods
type Client interface {
	SubmitWithRetries(ctx context.Context, request express.ChatCompletionJSONRequestBody) (*express.ExpressModeResponse, error)
	Submit(ctx context.Context, request express.ChatCompletionJSONRequestBody) (*express.ExpressModeResponse, error)
}

type GenAI struct {
	client    express.ClientWithResponsesInterface
	sanitizer *Sanitizer
}

type GetSecret func() string

func NewClient(baseURL, identityServiceURL, identityAppID, identityJobID string, getSecret GetSecret) (*GenAI, error) {
	client, err := express.NewClientWithResponses(baseURL, express.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		// Get authorization header from identity service
		header, err := getAuthorizationHeaderFromIdentityService(identityServiceURL, identityAppID, identityJobID, getSecret)
		if err != nil {
			return fmt.Errorf("error getting authorization header: %w", err)
		}

		// Set authorization header
		req.Header.Set("Authorization", header)

		// Set originating asset alias header
		req.Header.Set(IntuitOriginatingAssetAliasHeader, config.AssetAlias)
		return nil
	}))

	if err != nil {
		return nil, err
	}

	return &GenAI{client, NewSanitizer()}, nil
}

func (g GenAI) SubmitWithRetries(ctx context.Context, request express.ChatCompletionJSONRequestBody) (*express.ExpressModeResponse, error) {
	// Initialize the backoff duration and the backoff factor
	backoff := 1 * time.Second
	const backoffFactor = 2

	// Create a timer with a very short duration that will fire immediately
	timer := time.NewTimer(0)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			// The context's deadline has been exceeded or it has been cancelled
			return nil, fmt.Errorf("operation cancelled or timed out: %w", ctx.Err())

		case <-timer.C:
			// Attempt the Submit operation
			response, err := g.Submit(ctx, request)
			if err == nil {
				return response, nil
			}

			// Log the error
			requestType, ok := ctx.Value(typeKey).(string)
			if !ok {
				requestType = "unknown" // Default or error handling
			}
			log.Error().Err(err).Str(string(typeKey), requestType).Msgf("error submitting request, retrying in %v", backoff)

			// Reset the timer for the backoff duration
			timer.Reset(backoff)

			// Increase the backoff for the next iteration
			backoff *= backoffFactor

			if backoff > DefaultTimeOut { // Ensure backoff doesn't exceed the default timeout
				backoff = DefaultTimeOut
			}
		}
	}
}

func (g GenAI) Submit(ctx context.Context, request express.ChatCompletionJSONRequestBody) (*express.ExpressModeResponse, error) {
	for i, message := range request.LlmParams.Messages {
		request.LlmParams.Messages[i].Content = stringPtr(g.sanitizer.Sanitize(*message.Content))
	}

	// Create chat completion params
	chatCompletionParams := &express.ChatCompletionParams{
		ExperienceId: stringPtr(config.ExperienceID),
	}

	// Send request
	chatCompletionResponse, err := g.client.ChatCompletionWithResponse(ctx, chatCompletionParams, request)
	if err != nil {
		return nil, fmt.Errorf("error submitting request: %w", err)
	}

	// Handle HTTP errors
	if chatCompletionResponse.HTTPResponse.StatusCode >= 400 {
		return nil, fmt.Errorf("error submitting request: %s", chatCompletionResponse.Status())
	}

	// Create an instance of the struct to hold the JSON data
	var response express.ExpressModeResponse

	// Unmarshal the JSON data into the struct
	if err := json.Unmarshal(chatCompletionResponse.Body, &response); err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON response: %w", err)
	}

	return &response, nil
}

type IdentityResponse struct {
	Data struct {
		IdentitySignInInternalApplicationWithPrivateAuth struct {
			AuthorizationHeader string `json:"authorizationHeader"`
		} `json:"identitySignInInternalApplicationWithPrivateAuth"`
	} `json:"data"`
}

func getAuthorizationHeaderFromIdentityService(serviceURL, appID, jobID string, getSecret GetSecret) (string, error) {
	// Get app secret
	appSecret := getSecret()

	// Create headers
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Intuit_IAM_Authentication intuit_appid=%s, intuit_app_secret=%s", appID, appSecret),
		"Content-Type":  "application/json",
	}

	// Create request body
	requestBody := map[string]interface{}{
		"query": "mutation identitySignInInternalApplicationWithPrivateAuth($input: Identity_SignInApplicationWithPrivateAuthInput!) {  identitySignInInternalApplicationWithPrivateAuth(input: $input) { authorizationHeader }}",
		"variables": map[string]interface{}{
			"input": map[string]interface{}{
				"profileId": jobID,
			},
		},
	}

	// Encode body to JSON
	jsonData, _ := json.Marshal(requestBody)

	// Create request
	req, err := http.NewRequest("POST", serviceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// Handle HTTP errors
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("error getting authorization header: %s", resp.Status)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	// Unmarshal response body
	var response IdentityResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling response body: %w", err)
	}

	// Create authorization header
	authorizationHeader := fmt.Sprintf("%s,intuit_appid=%s,intuit_app_secret=%s",
		response.Data.IdentitySignInInternalApplicationWithPrivateAuth.AuthorizationHeader,
		appID,
		appSecret)
	return authorizationHeader, nil
}
