package repository

import (
	"context"
	"testing"

	"github.com/Sene4ka/url_shortener/internal/models"
	"github.com/Sene4ka/url_shortener/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemoryRepository_CreateLink_Success(t *testing.T) {
	repo := NewInMemoryRepository()
	ctx := context.Background()

	link := &models.Link{Id: "abc123", Url: "https://example.com"}

	created, err := repo.CreateLink(ctx, link)
	require.NoError(t, err)
	assert.Equal(t, link, created)

	saved, err := repo.GetById(ctx, link.Id)
	require.NoError(t, err)
	assert.Equal(t, link, saved)
}

func TestInMemoryRepository_CreateLink_IDAlreadyExists(t *testing.T) {
	repo := NewInMemoryRepository()
	ctx := context.Background()

	link1 := &models.Link{Id: "abc123", Url: "https://example1.com"}
	_, err := repo.CreateLink(ctx, link1)
	require.NoError(t, err)

	link2 := &models.Link{Id: "abc123", Url: "https://example2.com"}
	_, err = repo.CreateLink(ctx, link2)
	assert.ErrorIs(t, err, util.ErrIDExists)
}

func TestInMemoryRepository_CreateLink_DuplicateURL(t *testing.T) {
	repo := NewInMemoryRepository()
	ctx := context.Background()

	link1 := &models.Link{Id: "abc123", Url: "https://example.com"}
	created1, err := repo.CreateLink(ctx, link1)
	require.NoError(t, err)
	assert.Equal(t, link1, created1)

	link2 := &models.Link{Id: "xyz789", Url: "https://example.com"}
	created2, err := repo.CreateLink(ctx, link2)
	require.NoError(t, err)

	assert.Equal(t, created1.Id, created2.Id)
	assert.Equal(t, created1.Url, created2.Url)
}

func TestInMemoryRepository_GetById_Success(t *testing.T) {
	repo := NewInMemoryRepository()
	ctx := context.Background()

	link := &models.Link{Id: "abc123", Url: "https://example.com"}
	_, err := repo.CreateLink(ctx, link)
	require.NoError(t, err)

	found, err := repo.GetById(ctx, link.Id)
	require.NoError(t, err)
	assert.Equal(t, link, found)
}

func TestInMemoryRepository_GetById_NotFound(t *testing.T) {
	repo := NewInMemoryRepository()
	ctx := context.Background()

	_, err := repo.GetById(ctx, "nonexistent")
	assert.ErrorIs(t, err, util.ErrLinkNotFound)
}

func TestInMemoryRepository_CreateLink_ContextCanceled(t *testing.T) {
	repo := NewInMemoryRepository()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	link := &models.Link{Id: "abc123", Url: "https://example.com"}
	_, err := repo.CreateLink(ctx, link)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestInMemoryRepository_GetById_ContextCanceled(t *testing.T) {
	repo := NewInMemoryRepository()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := repo.GetById(ctx, "any")
	assert.ErrorIs(t, err, context.Canceled)
}

func TestInMemoryRepository_Concurrency(t *testing.T) {
	repo := NewInMemoryRepository()
	ctx := context.Background()

	const goroutines = 100
	const iterations = 100

	done := make(chan bool)
	for i := 0; i < goroutines; i++ {
		go func(i int) {
			for j := 0; j < iterations; j++ {
				id := "id"
				url := "https://example.com"
				link := &models.Link{Id: id, Url: url}
				_, _ = repo.CreateLink(ctx, link)
				_, _ = repo.GetById(ctx, id)
			}
			done <- true
		}(i)
	}

	for i := 0; i < goroutines; i++ {
		<-done
	}
}
