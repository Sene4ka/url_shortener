package shortener

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Sene4ka/url_shortener/configs"
	"github.com/Sene4ka/url_shortener/internal/models"
	"github.com/Sene4ka/url_shortener/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockService struct {
	createLinkFunc func(ctx context.Context, url string) (*models.Link, error)
	getByIdFunc    func(ctx context.Context, id string) (*models.Link, error)
}

func (m *mockService) CreateLink(ctx context.Context, url string) (*models.Link, error) {
	return m.createLinkFunc(ctx, url)
}

func (m *mockService) GetById(ctx context.Context, id string) (*models.Link, error) {
	return m.getByIdFunc(ctx, id)
}

func newTestHandler(svc Service) *LinkHandler {
	cfg := &configs.Config{
		Server: configs.ServerConfig{
			Host:         "localhost",
			Port:         "8080",
			ReadTimeout:  10,
			WriteTimeout: 10,
		},
	}
	return NewLinkHandler(svc, cfg)
}

func TestLinkHandler_Health(t *testing.T) {
	handler := newTestHandler(&mockService{})
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	handler.handleHealth(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "healthy", resp["status"])
}

func TestLinkHandler_CreateLink_Success(t *testing.T) {
	mockSvc := &mockService{
		createLinkFunc: func(ctx context.Context, url string) (*models.Link, error) {
			return &models.Link{Id: "abc1234567", Url: url}, nil
		},
	}
	handler := newTestHandler(mockSvc)

	body := `{"url":"https://example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/shortener/shorten/", strings.NewReader(body))
	rr := httptest.NewRecorder()

	handler.CreateLink(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var link models.Link
	err := json.Unmarshal(rr.Body.Bytes(), &link)
	require.NoError(t, err)
	assert.Equal(t, "abc1234567", link.Id)
	assert.Equal(t, "https://example.com", link.Url)
}

func TestLinkHandler_CreateLink_MethodNotAllowed(t *testing.T) {
	handler := newTestHandler(&mockService{})

	methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/shortener/shorten/", nil)
			rr := httptest.NewRecorder()
			handler.CreateLink(rr, req)

			assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
			var errResp map[string]string
			json.Unmarshal(rr.Body.Bytes(), &errResp)
			assert.Equal(t, "method not allowed", errResp["errors"])
		})
	}
}

func TestLinkHandler_CreateLink_InvalidJSON(t *testing.T) {
	handler := newTestHandler(&mockService{})

	req := httptest.NewRequest(http.MethodPost, "/shortener/shorten/", strings.NewReader(`{invalid json}`))
	rr := httptest.NewRecorder()
	handler.CreateLink(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	var errResp map[string]string
	json.Unmarshal(rr.Body.Bytes(), &errResp)
	assert.Equal(t, "invalid request body", errResp["errors"])
}

func TestLinkHandler_CreateLink_ValidationErrors(t *testing.T) {
	testCases := []struct {
		name         string
		err          error
		expectedCode int
	}{
		{"EmptyURL", util.ErrEmptyURL, http.StatusBadRequest},
		{"InvalidURL", util.ErrInvalidURL, http.StatusBadRequest},
		{"UnsupportedScheme", util.ErrUnsupportedScheme, http.StatusBadRequest},
		{"MissingHost", util.ErrMissingHost, http.StatusBadRequest},
		{"URLTooLong", util.ErrURLTooLong, http.StatusBadRequest},
		{"IDExists", util.ErrIDExists, http.StatusConflict},
		{"OtherError", errors.New("unexpected"), http.StatusInternalServerError},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := &mockService{
				createLinkFunc: func(ctx context.Context, url string) (*models.Link, error) {
					return nil, tc.err
				},
			}
			handler := newTestHandler(mockSvc)

			body := `{"url":"https://example.com"}`
			req := httptest.NewRequest(http.MethodPost, "/shortener/shorten/", strings.NewReader(body))
			rr := httptest.NewRecorder()
			handler.CreateLink(rr, req)

			assert.Equal(t, tc.expectedCode, rr.Code)
			var errResp map[string]string
			json.Unmarshal(rr.Body.Bytes(), &errResp)
			assert.Equal(t, tc.err.Error(), errResp["errors"])
		})
	}
}

func TestLinkHandler_GetById_Success(t *testing.T) {
	expectedLink := &models.Link{Id: "abc1234567", Url: "https://example.com"}
	mockSvc := &mockService{
		getByIdFunc: func(ctx context.Context, id string) (*models.Link, error) {
			assert.Equal(t, "abc1234567", id)
			return expectedLink, nil
		},
	}
	handler := newTestHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/shortener/url/abc1234567", nil)
	rr := httptest.NewRecorder()
	handler.GetById(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var link models.Link
	err := json.Unmarshal(rr.Body.Bytes(), &link)
	require.NoError(t, err)
	assert.Equal(t, expectedLink.Id, link.Id)
	assert.Equal(t, expectedLink.Url, link.Url)
}

func TestLinkHandler_GetById_NotFound(t *testing.T) {
	mockSvc := &mockService{
		getByIdFunc: func(ctx context.Context, id string) (*models.Link, error) {
			return nil, util.ErrLinkNotFound
		},
	}
	handler := newTestHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/shortener/url/nonexistent", nil)
	rr := httptest.NewRecorder()
	handler.GetById(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	var errResp map[string]string
	json.Unmarshal(rr.Body.Bytes(), &errResp)
	assert.Equal(t, util.ErrLinkNotFound.Error(), errResp["errors"])
}

func TestLinkHandler_GetById_MissingID(t *testing.T) {
	handler := newTestHandler(&mockService{})

	req := httptest.NewRequest(http.MethodGet, "/shortener/url/", nil)
	rr := httptest.NewRecorder()
	handler.GetById(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	var errResp map[string]string
	json.Unmarshal(rr.Body.Bytes(), &errResp)
	assert.Equal(t, "missing link id", errResp["errors"])
}

func TestLinkHandler_GetById_MethodNotAllowed(t *testing.T) {
	handler := newTestHandler(&mockService{})

	methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/shortener/url/abc123", nil)
			rr := httptest.NewRecorder()
			handler.GetById(rr, req)

			assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
			var errResp map[string]string
			json.Unmarshal(rr.Body.Bytes(), &errResp)
			assert.Equal(t, "method not allowed", errResp["errors"])
		})
	}
}

func TestLinkHandler_GetById_InternalError(t *testing.T) {
	mockSvc := &mockService{
		getByIdFunc: func(ctx context.Context, id string) (*models.Link, error) {
			return nil, errors.New("db connection error")
		},
	}
	handler := newTestHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/shortener/url/abc123", nil)
	rr := httptest.NewRecorder()
	handler.GetById(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	var errResp map[string]string
	json.Unmarshal(rr.Body.Bytes(), &errResp)
	assert.Equal(t, "db connection error", errResp["errors"])
}
