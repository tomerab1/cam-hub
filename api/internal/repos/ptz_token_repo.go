package repos

import (
	"context"
	"errors"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"tomerab.com/cam-hub/internal/api/v1/models"
)

var (
	ErrNoRowsAffected = errors.New("no rows were affected")
	ErrNotFound       = errors.New("ptz token not found")
)

type PtzTokenRepoIface interface {
	UpsertToken(ctx context.Context, token *models.PtzToken) error
	FindOne(ctx context.Context, uuid string) (*models.PtzToken, error)
}

type PgxPtzTokenRepo struct {
	DB DBPoolIface
}

func NewPgxPtzTokenRepo(db DBPoolIface) *PgxPtzTokenRepo {
	return &PgxPtzTokenRepo{
		DB: db,
	}
}

func (repo *PgxPtzTokenRepo) UpsertToken(ctx context.Context, token *models.PtzToken) error {
	tag, err := repo.DB.Exec(ctx,
		`INSERT INTO ptz_tokens (id, token)
			VALUES ($1, $2)
			ON CONFLICT (id) DO UPDATE SET token = EXCLUDED.token`,
		token.UUID,
		token.Token)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrNoRowsAffected
	}

	return nil
}

func (repo *PgxPtzTokenRepo) FindOne(ctx context.Context, uuid string) (*models.PtzToken, error) {
	var ptzToken models.PtzToken
	if err := pgxscan.Get(ctx,
		repo.DB,
		&ptzToken,
		`SELECT id, token FROM ptz_tokens WHERE id = $1`,
		uuid); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &ptzToken, nil
}
