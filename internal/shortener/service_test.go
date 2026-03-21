package shortener

import (
	"context"
	"strings"
	"testing"

	"github.com/Sene4ka/url_shortener/internal/repository"
	"github.com/Sene4ka/url_shortener/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockIDGenerator struct {
	id  string
	err error
}

func (m *mockIDGenerator) GenerateId() (string, error) {
	return m.id, m.err
}

func TestLinkService_CreateLink_Success(t *testing.T) {
	repo := repository.NewInMemoryRepository()
	gen := &mockIDGenerator{id: "abcdefghij", err: nil}
	svc := NewLinkService(gen, repo)

	ctx := context.Background()
	link, err := svc.CreateLink(ctx, "https://example.com")

	require.NoError(t, err)
	assert.NotNil(t, link)
	assert.Equal(t, "abcdefghij", link.Id)
	assert.Equal(t, "https://example.com", link.Url)

	saved, err := repo.GetById(ctx, link.Id)
	require.NoError(t, err)
	assert.Equal(t, link, saved)
}

func TestLinkService_CreateLink_ValidationErrors(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want error
	}{
		{"empty", "", util.ErrEmptyURL},
		{"invalid url", "://example.com", util.ErrInvalidURL},
		{"unsupported scheme", "ftp://example.com", util.ErrUnsupportedScheme},
		{"missing host 1", "https://", util.ErrMissingHost},
		{"missing host 2", "https:/abc", util.ErrMissingHost},
		{"too long", "https://example.com/" + strings.Repeat("a", 1005), util.ErrURLTooLong},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repository.NewInMemoryRepository()
			gen := &mockIDGenerator{id: "abc", err: nil}
			svc := NewLinkService(gen, repo)

			_, err := svc.CreateLink(context.Background(), tt.url)
			assert.ErrorIs(t, err, tt.want)
		})
	}
}

func TestLinkService_CreateLink_GenError(t *testing.T) {
	repo := repository.NewInMemoryRepository()
	gen := &mockIDGenerator{err: util.ErrFailedToGenerateID}
	svc := NewLinkService(gen, repo)

	_, err := svc.CreateLink(context.Background(), "https://example.com")
	assert.ErrorIs(t, err, util.ErrFailedToGenerateID)
}

func TestLinkService_CreateLink_InvalidIDLength(t *testing.T) {
	repo := repository.NewInMemoryRepository()
	gen := &mockIDGenerator{id: "short"}
	svc := NewLinkService(gen, repo)

	_, err := svc.CreateLink(context.Background(), "https://example.com")
	assert.ErrorIs(t, err, util.ErrInvalidIDLength)
}

func TestLinkService_CreateLink_IDConflict(t *testing.T) {
	repo := repository.NewInMemoryRepository()
	gen1 := &mockIDGenerator{id: "sameid1234"}
	svc1 := NewLinkService(gen1, repo)
	_, err := svc1.CreateLink(context.Background(), "https://first.com")
	require.NoError(t, err)

	gen2 := &mockIDGenerator{id: "sameid1234"}
	svc2 := NewLinkService(gen2, repo)
	_, err = svc2.CreateLink(context.Background(), "https://second.com")
	assert.ErrorIs(t, err, util.ErrIDExists)
}

func TestLinkService_CreateLink_DuplicateURL(t *testing.T) {
	repo := repository.NewInMemoryRepository()
	gen1 := &mockIDGenerator{id: "abc1234567"}
	svc1 := NewLinkService(gen1, repo)
	link1, err := svc1.CreateLink(context.Background(), "https://duplicate.com")
	require.NoError(t, err)

	gen2 := &mockIDGenerator{id: "xyz9876543"}
	svc2 := NewLinkService(gen2, repo)
	link2, err := svc2.CreateLink(context.Background(), "https://duplicate.com")
	require.NoError(t, err)

	assert.Equal(t, link1.Id, link2.Id)
	assert.Equal(t, link1.Url, link2.Url)
}

func TestLinkService_GetById_Success(t *testing.T) {
	repo := repository.NewInMemoryRepository()
	gen := &mockIDGenerator{id: "test123456"}
	svc := NewLinkService(gen, repo)

	ctx := context.Background()
	created, err := svc.CreateLink(ctx, "https://example.com")
	require.NoError(t, err)

	found, err := svc.GetById(ctx, created.Id)
	require.NoError(t, err)
	assert.Equal(t, created, found)
}

func TestLinkService_GetById_NotFound(t *testing.T) {
	repo := repository.NewInMemoryRepository()
	svc := NewLinkService(&mockIDGenerator{}, repo)

	_, err := svc.GetById(context.Background(), "nonexistent")
	assert.ErrorIs(t, err, util.ErrLinkNotFound)
}
