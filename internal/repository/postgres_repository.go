package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Sene4ka/url_shortener/internal/models"
	"github.com/Sene4ka/url_shortener/internal/util"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (p *PostgresRepository) CreateLink(ctx context.Context, link *models.Link) (*models.Link, error) {
	query := `
        INSERT INTO links (id, url)
        VALUES ($1, $2)
        ON CONFLICT (url) DO NOTHING
        RETURNING id, url
    `

	row := p.db.QueryRow(ctx, query, link.Id, link.Url)
	var created models.Link
	err := row.Scan(&created.Id, &created.Url)
	if err == nil {
		return &created, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		existing, err := p.GetByUrl(ctx, link.Url)
		if err != nil {
			return nil, err
		}
		return existing, nil
	}

	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" {
		return nil, util.ErrIDExists
	}

	return nil, fmt.Errorf("failed to create link: %w", err)
}

func (p *PostgresRepository) GetById(ctx context.Context, id string) (*models.Link, error) {
	query := `SELECT id, url FROM links WHERE id = $1`

	row := p.db.QueryRow(ctx, query, id)
	var link models.Link

	err := row.Scan(&link.Id, &link.Url)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, util.ErrLinkNotFound
		}
		return nil, fmt.Errorf("failed to get link by id: %w", err)
	}

	return &link, nil
}

func (p *PostgresRepository) GetByUrl(ctx context.Context, url string) (*models.Link, error) {
	query := `SELECT id, url FROM links WHERE url = $1`

	row := p.db.QueryRow(ctx, query, url)

	var existingLink models.Link
	err := row.Scan(&existingLink.Id, &existingLink.Url)
	if err != nil {
		return nil, util.ErrLinkNotFound
	}

	return &existingLink, nil
}
