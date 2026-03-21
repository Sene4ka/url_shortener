package repository

import (
	"context"
	"sync"

	"github.com/Sene4ka/url_shortener/internal/models"
	"github.com/Sene4ka/url_shortener/internal/util"
)

type InMemoryRepository struct {
	mutex   sync.RWMutex
	links   map[string]string
	reverse map[string]string
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		mutex:   sync.RWMutex{},
		links:   make(map[string]string),
		reverse: make(map[string]string),
	}
}

func (i *InMemoryRepository) CreateLink(ctx context.Context, link *models.Link) (*models.Link, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	i.mutex.Lock()
	defer i.mutex.Unlock()

	if _, ok := i.links[link.Id]; ok {
		return nil, util.ErrIDExists
	}
	if _, ok := i.reverse[link.Url]; ok {
		return &models.Link{Id: i.reverse[link.Url], Url: link.Url}, nil
	}
	i.links[link.Id] = link.Url
	i.reverse[link.Url] = link.Id

	return link, nil
}

func (i *InMemoryRepository) GetById(ctx context.Context, id string) (*models.Link, error) {
	select {
	case <-ctx.Done():
		return &models.Link{}, ctx.Err()
	default:
	}

	i.mutex.RLock()
	defer i.mutex.RUnlock()

	url, ok := i.links[id]
	if !ok {
		return nil, util.ErrLinkNotFound
	}

	return &models.Link{Id: id, Url: url}, nil
}
