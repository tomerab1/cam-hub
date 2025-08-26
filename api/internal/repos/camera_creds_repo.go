package repos

import (
	"context"
	"fmt"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"tomerab.com/cam-hub/internal/api/v1/models"
)

type PgxCameraCredsRepo struct {
	DB DBPoolIface
}

func NewPgxCameraCredsRepo(db DBPoolIface) *PgxCameraCredsRepo {
	return &PgxCameraCredsRepo{
		DB: db,
	}
}

func (repo *PgxCameraCredsRepo) InsertCreds(ctx context.Context, tx pgx.Tx, creds *models.CameraCreds) error {
	if creds == nil {
		return fmt.Errorf("invalid argument: creds are null")
	}

	_, err := tx.Exec(ctx, `
		INSERT INTO camera_creds (id, username, password)
		VALUES ($1,$2,$3)
		ON CONFLICT (id) DO UPDATE SET
			username = EXCLUDED.username,
			password = EXCLUDED.password
	`,
		creds.UUID, creds.Username, creds.Password,
	)

	return err
}

func (repo *PgxCameraCredsRepo) FindOne(ctx context.Context, uuid string) (*models.CameraCreds, error) {
	var creds models.CameraCreds
	err := pgxscan.Get(ctx, repo.DB, &creds, `SELECT username, password
												FROM camera_creds
												WHERE id = $1
	`, uuid)

	return &creds, err
}
