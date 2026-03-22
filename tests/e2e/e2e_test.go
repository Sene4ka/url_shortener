//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Sene4ka/url_shortener/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	appImage = "url_shortener-shortener:latest"
	appPort  = "8080"
)

func startContainer(t *testing.T) (baseURL string, cleanup func()) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        appImage,
		ExposedPorts: []string{fmt.Sprintf("%s/tcp", appPort)},
		Env: map[string]string{
			"DB_USE_IN_MEMORY": "true",
		},
		WaitingFor: wait.ForHTTP("/health").WithPort(appPort).WithStartupTimeout(30 * time.Second),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	mappedPort, err := container.MappedPort(ctx, appPort)
	require.NoError(t, err)
	host, err := container.Host(ctx)
	require.NoError(t, err)
	baseURL = fmt.Sprintf("http://%s:%s", host, mappedPort.Port())

	cleanup = func() {
		_ = container.Terminate(ctx)
	}
	return baseURL, cleanup
}

func isAllowedChar(c rune) bool {
	return (c >= '0' && c <= '9') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= 'a' && c <= 'z') ||
		c == '_'
}

func TestE2E_CreateAndGetLink(t *testing.T) {
	baseURL, cleanup := startContainer(t)
	defer cleanup()

	url := "https://example.com"
	reqBody := map[string]string{"url": url}
	jsonBody, _ := json.Marshal(reqBody)

	resp, err := http.Post(baseURL+"/shortener/shorten/", "application/json", bytes.NewReader(jsonBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var link models.Link
	err = json.NewDecoder(resp.Body).Decode(&link)
	require.NoError(t, err)
	assert.NotEmpty(t, link.Id)
	assert.Equal(t, url, link.Url)

	for _, ch := range link.Id {
		assert.True(t, isAllowedChar(ch), "недопустимый символ в ID: %c", ch)
	}

	getResp, err := http.Get(baseURL + "/shortener/url/" + link.Id)
	require.NoError(t, err)
	defer getResp.Body.Close()

	assert.Equal(t, http.StatusOK, getResp.StatusCode)

	var gotLink models.Link
	err = json.NewDecoder(getResp.Body).Decode(&gotLink)
	require.NoError(t, err)
	assert.Equal(t, link, gotLink)
}

func TestE2E_CreateLink_DuplicateURL(t *testing.T) {
	baseURL, cleanup := startContainer(t)
	defer cleanup()

	url := "https://duplicate.com"

	reqBody := map[string]string{"url": url}
	jsonBody, _ := json.Marshal(reqBody)

	resp1, err := http.Post(baseURL+"/shortener/shorten/", "application/json", bytes.NewReader(jsonBody))
	require.NoError(t, err)
	defer resp1.Body.Close()
	assert.Equal(t, http.StatusOK, resp1.StatusCode)

	var link1 models.Link
	err = json.NewDecoder(resp1.Body).Decode(&link1)
	require.NoError(t, err)

	resp2, err := http.Post(baseURL+"/shortener/shorten/", "application/json", bytes.NewReader(jsonBody))
	require.NoError(t, err)
	defer resp2.Body.Close()
	assert.Equal(t, http.StatusOK, resp2.StatusCode)

	var link2 models.Link
	err = json.NewDecoder(resp2.Body).Decode(&link2)
	require.NoError(t, err)

	assert.Equal(t, link1.Id, link2.Id)
	assert.Equal(t, link1.Url, link2.Url)
}

func TestE2E_CreateLink_ValidationErrors(t *testing.T) {
	baseURL, cleanup := startContainer(t)
	defer cleanup()

	tests := []struct {
		name     string
		url      string
		wantCode int
		wantErr  string
	}{
		{"Empty URL", "", http.StatusBadRequest, "url field is empty"},
		{"Invalid URL", "://example.com", http.StatusBadRequest, "invalid url"},
		{"Unsupported scheme", "ftp://example.com", http.StatusBadRequest, "unsupported url scheme"},
		{"Missing host", "https://", http.StatusBadRequest, "url is missing host"},
		{"Too long", "https://example.com/" + strings.Repeat("a", 1005), http.StatusBadRequest, "url is too long"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := map[string]string{"url": tt.url}
			jsonBody, _ := json.Marshal(reqBody)

			resp, err := http.Post(baseURL+"/shortener/shorten/", "application/json", bytes.NewReader(jsonBody))
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.wantCode, resp.StatusCode)

			var errResp map[string]string
			err = json.NewDecoder(resp.Body).Decode(&errResp)
			require.NoError(t, err)
			assert.Contains(t, errResp["errors"], tt.wantErr)
		})
	}
}

func TestE2E_GetById_NotFound(t *testing.T) {
	baseURL, cleanup := startContainer(t)
	defer cleanup()

	resp, err := http.Get(baseURL + "/shortener/url/nonexistent")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	var errResp map[string]string
	err = json.NewDecoder(resp.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.Equal(t, "link not found", errResp["errors"])
}

func TestE2E_GetById_MissingID(t *testing.T) {
	baseURL, cleanup := startContainer(t)
	defer cleanup()

	resp, err := http.Get(baseURL + "/shortener/url/")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var errResp map[string]string
	err = json.NewDecoder(resp.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.Equal(t, "missing link id", errResp["errors"])
}

func TestE2E_MethodNotAllowed_CreateLink(t *testing.T) {
	baseURL, cleanup := startContainer(t)
	defer cleanup()

	resp, err := http.Get(baseURL + "/shortener/shorten/")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}

func TestE2E_MethodNotAllowed_GetById(t *testing.T) {
	baseURL, cleanup := startContainer(t)
	defer cleanup()

	reqBody := map[string]string{"url": "https://example.com"}
	jsonBody, _ := json.Marshal(reqBody)

	resp, err := http.Post(baseURL+"/shortener/url/abc123", "application/json", bytes.NewReader(jsonBody))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}

func TestE2E_Health(t *testing.T) {
	baseURL, cleanup := startContainer(t)
	defer cleanup()

	resp, err := http.Get(baseURL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var health map[string]string
	err = json.NewDecoder(resp.Body).Decode(&health)
	require.NoError(t, err)
	assert.Equal(t, "healthy", health["status"])
}
