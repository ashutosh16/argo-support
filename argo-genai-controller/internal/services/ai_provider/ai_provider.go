package ai_provider

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/argoproj-labs/argo-support/api/v1alpha1"
	"github.com/argoproj-labs/argo-support/internal/utils"
	"github.com/argoproj-labs/argo-support/internal/wf_operations/common"
	"io"

	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"time"
)

const (
	AppSecretKey = "app.secret"
)

type ArgoCDClient struct {
	BaseURL   string
	AppSecret string
}

type AIProvider interface {
	SubmitTokensToGenAI(ctx context.Context, tokens string, endpointSuffix string) (interface{}, error)
}

func (client *ArgoCDClient) GetArgoApp(fullUrl string) (*common.Application, error) {
	req, err := http.NewRequest("GET", fullUrl, nil)
	if err != nil {
		return nil, err
	}

	// Adding headers as per the curl command
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Cookie", "argocd.token="+client.AppSecret+"; Secure; HttpOnly")
	httpClient := &http.Client{

		Timeout: time.Minute * 5,
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned non-OK status: %d %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var app common.Application
	if err := json.Unmarshal(body, &app); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %v", err)
	}

	return &app, nil
}

func GetArgoCDClienWithSecret(ctx context.Context, k8sClient client.Client, authProvider *v1alpha1.AuthProvider, namespace string) (*ArgoCDClient, error) {
	logger := log.FromContext(ctx)

	secret, err := utils.GetSecret(ctx, k8sClient, authProvider)
	if err != nil {
		logger.Error(err, "failed to get Secret from AuthProvider", "namespace", namespace)
		return nil, err
	}

	if secret == nil {
		return nil, fmt.Errorf("secret is missing")
	}

	return &ArgoCDClient{
		BaseURL:   authProvider.Spec.Auth.BaseURL,
		AppSecret: string(secret.Data[AppSecretKey]),
	}, nil
	return nil, nil
}
