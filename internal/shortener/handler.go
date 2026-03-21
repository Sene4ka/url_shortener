package shortener

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/Sene4ka/url_shortener/configs"
	"github.com/Sene4ka/url_shortener/internal/models"
	"github.com/Sene4ka/url_shortener/internal/util"
)

type Service interface {
	CreateLink(ctx context.Context, url string) (*models.Link, error)
	GetById(ctx context.Context, id string) (*models.Link, error)
}

type LinkHandler struct {
	linkService Service
	httpServer  *http.Server
	config      *configs.Config
}

func NewLinkHandler(linkService Service, config *configs.Config) *LinkHandler {
	handler := &LinkHandler{
		config:      config,
		linkService: linkService,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/health", handler.handleHealth)

	mux.HandleFunc("/shortener/shorten/", handler.CreateLink)

	mux.HandleFunc("/shortener/url/", handler.GetById)

	handler.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%s", config.Server.Host, config.Server.Port),
		Handler:      CORS(mux),
		ReadTimeout:  config.Server.ReadTimeout,
		WriteTimeout: config.Server.WriteTimeout,
	}

	return handler
}

func (l *LinkHandler) Start() error {
	return l.httpServer.ListenAndServe()
}

func (l *LinkHandler) Shutdown(ctx context.Context) error {
	return l.httpServer.Shutdown(ctx)
}

func (l *LinkHandler) handleHealth(w http.ResponseWriter, r *http.Request) {
	JSONResponse(w, http.StatusOK, map[string]string{"status": "healthy"})
}

func (l *LinkHandler) CreateLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req UrlInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := l.linkService.CreateLink(r.Context(), req.Url)
	if err != nil {
		message := err.Error()
		if errors.Is(err, util.ErrIDExists) {
			JSONError(w, http.StatusConflict, message)
			return
		}

		if errors.Is(err, util.ErrEmptyURL) ||
			errors.Is(err, util.ErrInvalidURL) || errors.Is(err, util.ErrUnsupportedScheme) ||
			errors.Is(err, util.ErrMissingHost) || errors.Is(err, util.ErrURLTooLong) {
			JSONError(w, http.StatusBadRequest, message)
			return
		}

		JSONError(w, http.StatusInternalServerError, message)
		return
	}

	JSONResponse(w, http.StatusOK, resp)
}

func (l *LinkHandler) GetById(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	linkId := strings.TrimPrefix(r.URL.Path, "/shortener/url/")
	if linkId == "" {
		JSONError(w, http.StatusBadRequest, "missing link id")
		return
	}

	link, err := l.linkService.GetById(r.Context(), linkId)
	if err != nil {
		message := err.Error()
		if errors.Is(err, util.ErrLinkNotFound) {
			JSONError(w, http.StatusNotFound, message)
			return
		}

		JSONError(w, http.StatusInternalServerError, message)
		return
	}

	JSONResponse(w, http.StatusOK, link)
}

func JSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Printf("failed to encode JSON response: %v", err)
	}
}

func JSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(map[string]string{"errors": message})
	if err != nil {
		log.Printf("failed to encode JSON errors: %v", err)
	}
}
