package splunk

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPublish_Success(t *testing.T) {
	// Create a new httptest server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Respond with a dummy response or the expected response
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{ "status": "success" }`))
	}))
	defer server.Close()

	// Create your Splunk client with the mock server URL
	client := NewSplunk(server.URL, "dummy-token", "dev")

	// Test the Publish method
	err := client.Publish("test data", SplunkSourceTypeFeedback)
	assert.NoError(t, err)
}

func TestPublish_Error(t *testing.T) {
	// Create a new httptest server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Respond with a dummy response or the expected response
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{ "status": "error" }`))
	}))
	defer server.Close()

	// Create your Splunk client with the mock server URL
	client := NewSplunk(server.URL, "dummy-token", "dev")

	// Test the Publish method
	err := client.Publish("test data", SplunkSourceTypeFeedback)
	assert.Error(t, err)
}
