package genstudio

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/argoproj-labs/argo-support/api/v1alpha1"
	"github.com/argoproj-labs/argo-support/internal/services/ai_provider"
	"github.com/argoproj-labs/argo-support/internal/utils"
	"io"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"time"
)

type GenStudioClient struct {
	BaseURL          string
	AppID            string
	AppSecret        string
	IdentityEndpoint string
	IdentityJobID    string
	APIVersion       string
}

var (
	_ ai_provider.AIProvider = &GenStudioClient{}
)

type IdentityResponse struct {
	Data struct {
		IdentitySignInInternalApplicationWithPrivateAuth struct {
			AuthorizationHeader string `json:"authorizationHeader"`
		} `json:"identitySignInInternalApplicationWithPrivateAuth"`
	} `json:"data"`
}

func (client *GenStudioClient) GetAuthorizationHeaderFromIdentityService() (string, error) {
	headers := make(map[string]string)
	headers["Authorization"] = fmt.Sprintf("Intuit_IAM_Authentication intuit_appid=%s, intuit_app_secret=%s", client.AppID, client.AppSecret)
	headers["Content-Type"] = "application/json"

	requestBody := fmt.Sprintf(`{"query":"mutation identitySignInInternalApplicationWithPrivateAuth($input: Identity_SignInApplicationWithPrivateAuthInput!) { identitySignInInternalApplicationWithPrivateAuth(input: $input) { authorizationHeader }}","variables":{"input":{"profileId":%s}}}`, client.IdentityJobID)

	req, err := http.NewRequest("POST", client.IdentityEndpoint+"/v1/graphql", bytes.NewBufferString(requestBody))
	if err != nil {
		return "", err
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("error getting authorization header: %v", resp.Status)
	}

	var identityResponse IdentityResponse
	if err := json.NewDecoder(resp.Body).Decode(&identityResponse); err != nil {
		return "", err
	}

	authorizationHeader := fmt.Sprintf("%s,intuit_appid=%s,intuit_app_secret=%s", identityResponse.Data.IdentitySignInInternalApplicationWithPrivateAuth.AuthorizationHeader, client.AppID, client.AppSecret)

	return authorizationHeader, nil
}

func (client *GenStudioClient) SubmitTokensToGenAI(ctx context.Context, tokens string, endpointSuffix string) (interface{}, error) {
	logger := log.FromContext(ctx)
	if len(tokens) == 0 {
		return nil, fmt.Errorf("skipping the genai request due to no tokens for genai")
	}

	if !json.Valid([]byte(tokens)) {
		return nil, fmt.Errorf("Unable to generate token for GenAI")
	}
	body := []byte(tokens)

	authorizationHeader, err := client.GetAuthorizationHeaderFromIdentityService()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", client.BaseURL+"/"+client.APIVersion+endpointSuffix, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", authorizationHeader)
	req.Header.Add("Content-Type", "application/json")

	httpClient := &http.Client{
		Timeout: time.Minute * 5,
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Error(err, "received response error from genai", "status-code", resp.StatusCode)
		return nil, err
	}

	if err != nil {
		return nil, err
	}
	resDataBytes, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		return nil, err
	}
	var resData interface{}
	err = json.Unmarshal(resDataBytes, &resData)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("unexpected response status: %v", resp.Status)
	}

	return resData, nil
}

func GenAIClientWithSecret(ctx context.Context, k8sClient client.Client, authProvider *v1alpha1.AuthProvider, namespace string) ai_provider.AIProvider {
	logger := log.FromContext(ctx)

	secret, err := utils.GetSecret(ctx, k8sClient, authProvider)
	if err != nil {
		logger.Error(err, "failed to get Secret from AuthProvider", "namespace", namespace, "name", authProvider.Name)
		return nil
	}

	if secret == nil {
		return nil
	}

	return &GenStudioClient{
		BaseURL:          authProvider.Spec.Auth.BaseURL,
		AppID:            authProvider.Spec.Auth.AppID,
		IdentityEndpoint: authProvider.Spec.Auth.IdentityEndpoint,
		IdentityJobID:    authProvider.Spec.Auth.IdentityJobID,
		APIVersion:       authProvider.Spec.Auth.APIVersion,
		AppSecret:        string(secret.Data[ai_provider.AppSecretKey]),
	}
	return nil
}
