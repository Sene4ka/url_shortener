package shortener

import (
	"context"
	"net/url"

	"github.com/Sene4ka/url_shortener/internal/models"
	"github.com/Sene4ka/url_shortener/internal/util"
)

type IdGenerator interface {
	GenerateId() (string, error)
}

type LinkRepository interface {
	CreateLink(ctx context.Context, link *models.Link) (*models.Link, error)
	GetById(ctx context.Context, id string) (*models.Link, error)
}

type LinkService struct {
	gen  IdGenerator
	repo LinkRepository
}

func NewLinkService(gen IdGenerator, repo LinkRepository) *LinkService {
	return &LinkService{gen: gen, repo: repo}
}

func (s *LinkService) CreateLink(ctx context.Context, url string) (*models.Link, error) {

	err := validateURL(url)
	if err != nil {
		return nil, err
	}

	id, err := s.gen.GenerateId()
	if err != nil {
		return nil, util.ErrFailedToGenerateID
	}
	if len(id) != 10 {
		return nil, util.ErrInvalidIDLength
	}

	link := &models.Link{
		Id:  id,
		Url: url,
	}

	link, err = s.repo.CreateLink(ctx, link)
	if err != nil {
		return nil, err
	}

	return link, nil
}

func (s *LinkService) GetById(ctx context.Context, id string) (*models.Link, error) {
	link, err := s.repo.GetById(ctx, id)
	if err != nil {
		return nil, err
	}

	return link, nil
}

func validateURL(rawURL string) error {
	if rawURL == "" {
		return util.ErrEmptyURL
	}

	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return util.ErrInvalidURL
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return util.ErrUnsupportedScheme
	}

	if u.Host == "" {
		return util.ErrMissingHost
	}

	if len(rawURL) > 1024 {
		return util.ErrURLTooLong
	}

	return nil
}
