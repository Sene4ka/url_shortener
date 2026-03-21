package repository

import (
	"context"
	"os"
	"testing"

	"github.com/Sene4ka/url_shortener/internal/models"
	"github.com/Sene4ka/url_shortener/internal/util"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func setupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("Skipping PostgreSQL tests: TEST_POSTGRES_DSN not set")
	}

	ctx := context.Background()

	postgresContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
	)
	require.NoError(t, err)

	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)

	_, err = db.Exec(ctx, `
        CREATE TABLE links (
            id VARCHAR(10) PRIMARY KEY,
            url TEXT UNIQUE NOT NULL
        );
    `)
	require.NoError(t, err)

	cleanup := func() {
		db.Close()
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate container: %v", err)
		}
	}
	return db, cleanup
}

func TestPostgresRepository_CreateLink_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	repo := NewPostgresRepository(db)

	ctx := context.Background()
	link := &models.Link{Id: "abc123", Url: "https://example.com"}

	created, err := repo.CreateLink(ctx, link)
	require.NoError(t, err)
	assert.Equal(t, link.Id, created.Id)
	assert.Equal(t, link.Url, created.Url)

	found, err := repo.GetById(ctx, link.Id)
	require.NoError(t, err)
	assert.Equal(t, link, found)
}

func TestPostgresRepository_CreateLink_IDConflict(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	repo := NewPostgresRepository(db)

	ctx := context.Background()
	link1 := &models.Link{Id: "abc123", Url: "https://example1.com"}
	_, err := repo.CreateLink(ctx, link1)
	require.NoError(t, err)

	link2 := &models.Link{Id: "abc123", Url: "https://example2.com"}
	_, err = repo.CreateLink(ctx, link2)
	assert.ErrorIs(t, err, util.ErrIDExists)
}

func TestPostgresRepository_CreateLink_DuplicateURL(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	repo := NewPostgresRepository(db)

	ctx := context.Background()
	link1 := &models.Link{Id: "abc123", Url: "https://example.com"}
	created1, err := repo.CreateLink(ctx, link1)
	require.NoError(t, err)

	link2 := &models.Link{Id: "xyz789", Url: "https://example.com"}
	created2, err := repo.CreateLink(ctx, link2)
	require.NoError(t, err)

	assert.Equal(t, created1.Id, created2.Id)
	assert.Equal(t, created1.Url, created2.Url)
}

func TestPostgresRepository_GetById_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	repo := NewPostgresRepository(db)

	ctx := context.Background()
	link := &models.Link{Id: "abc123", Url: "https://example.com"}
	_, err := repo.CreateLink(ctx, link)
	require.NoError(t, err)

	found, err := repo.GetById(ctx, link.Id)
	require.NoError(t, err)
	assert.Equal(t, link, found)
}

func TestPostgresRepository_GetById_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	repo := NewPostgresRepository(db)

	ctx := context.Background()
	_, err := repo.GetById(ctx, "nonexistent")
	assert.ErrorIs(t, err, util.ErrLinkNotFound)
}

func TestPostgresRepository_GetByUrl_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	repo := NewPostgresRepository(db)

	ctx := context.Background()
	link := &models.Link{Id: "abc123", Url: "https://example.com"}
	_, err := repo.CreateLink(ctx, link)
	require.NoError(t, err)

	found, err := repo.GetByUrl(ctx, link.Url)
	require.NoError(t, err)
	assert.Equal(t, link, found)
}

func TestPostgresRepository_GetByUrl_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	repo := NewPostgresRepository(db)

	ctx := context.Background()
	_, err := repo.GetByUrl(ctx, "https://notexists.com")
	assert.ErrorIs(t, err, util.ErrLinkNotFound)
}

func TestPostgresRepository_CreateLink_ContextCanceled(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	repo := NewPostgresRepository(db)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	link := &models.Link{Id: "abc123", Url: "https://example.com"}
	_, err := repo.CreateLink(ctx, link)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestPostgresRepository_GetById_ContextCanceled(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	repo := NewPostgresRepository(db)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := repo.GetById(ctx, "any")
	assert.ErrorIs(t, err, context.Canceled)
}
