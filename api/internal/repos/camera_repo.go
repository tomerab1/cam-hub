package repos

import (
	"context"
	"fmt"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"tomerab.com/cam-hub/internal/api/v1/models"
)

type CameraRepoIface interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	UpsertCameraTx(ctx context.Context, tx pgx.Tx, cam *models.Camera) error
	UpsertCamera(ctx context.Context, cam *models.Camera) error
	FindExistingPaired(ctx context.Context, uuids []string) ([]bool, error)
	FindOne(ctx context.Context, uuid string) (*models.Camera, error)
	FindMany(ctx context.Context, offset, limit int) ([]*models.Camera, error)
	Save(ctx context.Context, cam *models.Camera) error
	Delete(ctx context.Context, uuid string) error
}

type PgxCameraRepo struct {
	DB DBPoolIface
}

func NewPgxCameraRepo(db DBPoolIface) *PgxCameraRepo {
	return &PgxCameraRepo{
		DB: db,
	}
}

func (repo *PgxCameraRepo) Begin(ctx context.Context) (pgx.Tx, error) {
	return repo.DB.Begin(ctx)
}

func (repo *PgxCameraRepo) UpsertCameraTx(ctx context.Context, tx pgx.Tx, cam *models.Camera) error {
	// TODO(tomer): Remove this, better to fail fast
	// if cam == nil {
	// 	return fmt.Errorf("invalid argument: camera is null")
	// }

	_, err := tx.Exec(ctx, `
		INSERT INTO cameras (
			id, name, manufacturer, model, firmwareVersion, serialNumber, hardwareId, addr, isPaired
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			manufacturer = EXCLUDED.manufacturer,
			model = EXCLUDED.model,
			firmwareVersion = EXCLUDED.firmwareVersion,
			serialNumber = EXCLUDED.serialNumber,
			hardwareId = EXCLUDED.hardwareId,
			addr = EXCLUDED.addr,
			isPaired = EXCLUDED.isPaired
	`,
		cam.UUID,
		cam.CameraName,
		cam.Manufacturer,
		cam.Model,
		cam.FirmwareVersion,
		cam.SerialNumber,
		cam.HardwareId,
		cam.Addr,
		cam.IsPaired,
	)

	return err
}

func (repo *PgxCameraRepo) UpsertCamera(ctx context.Context, cam *models.Camera) error {
	_, err := repo.DB.Exec(ctx, `
		INSERT INTO cameras (
			id, name, manufacturer, model, firmwareVersion, serialNumber, hardwareId, addr, isPaired
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			manufacturer = EXCLUDED.manufacturer,
			model = EXCLUDED.model,
			firmwareVersion = EXCLUDED.firmwareVersion,
			serialNumber = EXCLUDED.serialNumber,
			hardwareId = EXCLUDED.hardwareId,
			addr = EXCLUDED.addr,
			isPaired = EXCLUDED.isPaired
	`,
		cam.UUID,
		cam.CameraName,
		cam.Manufacturer,
		cam.Model,
		cam.FirmwareVersion,
		cam.SerialNumber,
		cam.HardwareId,
		cam.Addr,
		cam.IsPaired,
	)

	return err
}

func (repo *PgxCameraRepo) FindExistingPaired(ctx context.Context, uuids []string) ([]bool, error) {
	batch := &pgx.Batch{}
	for _, uuid := range uuids {
		// Get only rows that exists in the db and are paired
		batch.Queue(`SELECT id
					 FROM cameras
					 WHERE id = $1 and ispaired = true
		`, uuid)
	}

	batchResults := repo.DB.SendBatch(ctx, batch)
	defer batchResults.Close()

	results := make([]bool, len(uuids))
	for i := range uuids {
		var tmp string
		err := batchResults.QueryRow().Scan(&tmp)
		if err == pgx.ErrNoRows {
			results[i] = false
			continue
		}
		if err != nil {
			return nil, err
		}

		results[i] = true
	}

	return results, nil
}

func (repo *PgxCameraRepo) FindOne(ctx context.Context, uuid string) (*models.Camera, error) {
	var cam models.Camera
	err := pgxscan.Get(ctx, repo.DB, &cam, `SELECT
													id,
													name,
													manufacturer,
													model,
													firmwareversion,
													serialnumber,
													hardwareid,
													addr,
													ispaired
												FROM cameras
												WHERE id = $1`, uuid)

	return &cam, err
}

func (repo *PgxCameraRepo) Save(ctx context.Context, cam *models.Camera) error {
	tag, err := repo.DB.Exec(ctx, `UPDATE cameras
		SET id = $1,
			name = $2,
			manufacturer = $3,
			model = $4,
			firmwareversion = $5,
			serialnumber = $6,
			hardwareid = $7,
			addr = $8,
			ispaired = $9
		WHERE id = $1
	`, cam.UUID, cam.CameraName, cam.Manufacturer, cam.Model, cam.FirmwareVersion, cam.SerialNumber,
		cam.HardwareId, cam.Addr, cam.IsPaired)

	if tag.RowsAffected() != 1 {
		return fmt.Errorf("save failed: no rows were affected (id=%s)", cam.UUID)
	}

	return err
}

func (repo *PgxCameraRepo) Delete(ctx context.Context, uuid string) error {
	tag, err := repo.DB.Exec(ctx, `DELETE FROM cameras WHERE id = $1`, uuid)

	if !tag.Delete() {
		return fmt.Errorf("failed to delete camera with uuid=%s", uuid)
	}

	return err
}

func (repo *PgxCameraRepo) FindMany(ctx context.Context, offset, limit int) ([]*models.Camera, error) {
	var cams []*models.Camera
	err := pgxscan.Select(ctx, repo.DB, &cams, `SELECT
													id,
													name,
													manufacturer,
													model,
													firmwareversion,
													serialnumber,
													hardwareid,
													addr,
													ispaired
												FROM cameras
												LIMIT $1 OFFSET $2
												`, limit, offset)
	return cams, err
}
