package repos

import (
	"context"
	"fmt"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"tomerab.com/cam-hub/internal/api/v1/models"
)

type CameraCredsRepoIface interface {
	InsertCreds(ctx context.Context, tx pgx.Tx, creds *models.CameraCreds) error
	FindOne(ctx context.Context, uuid string) (*models.CameraCreds, error)
}

type PgxCameraCredsRepo struct {
	DB DBPoolIface
}

func NewPgxCameraCredsRepo(db DBPoolIface) *PgxCameraCredsRepo {
	return &PgxCameraCredsRepo{
		DB: db,
	}
}

func (repo *PgxCameraCredsRepo) InsertCreds(ctx context.Context, tx pgx.Tx, creds *models.CameraCreds) error {
	tag, err := tx.Exec(ctx, `
		INSERT INTO camera_creds (id, username, password)
		VALUES ($1,$2,$3)
		ON CONFLICT (id) DO UPDATE SET
			username = EXCLUDED.username,
			password = EXCLUDED.password
	`,
		creds.UUID, creds.Username, creds.Password,
	)

	if !(tag.Insert() || tag.Update()) {
		return fmt.Errorf("failed to insert/update creds: uuid=%s, err=%v", creds.UUID, err)
	}

	return err
}

func (repo *PgxCameraCredsRepo) FindOne(ctx context.Context, uuid string) (*models.CameraCreds, error) {
	var creds models.CameraCreds
	err := pgxscan.Get(ctx, repo.DB, &creds, `SELECT id, username, password
												FROM camera_creds
												WHERE id = $1
	`, uuid)

	return &creds, err
}
