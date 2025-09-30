package repos

import (
	"context"

	"tomerab.com/cam-hub/internal/api/v1/models"
)

type RecordingsRepoIface interface {
	Upsert(ctx context.Context, recording *models.Recordings) error
	FindOne(ctx context.Context, uuid string) (*models.Recordings, error)
}

type PgxRecordingsRepo struct {
	DB DBPoolIface
}

func NewPgxRecordingsRepo(db DBPoolIface) *PgxRecordingsRepo {
	return &PgxRecordingsRepo{
		DB: db,
	}
}

func (repo *PgxRecordingsRepo) Upsert(ctx context.Context, recording *models.Recordings) error {
	repo.DB.Exec(ctx, `
		INSERT INTO recordings 
		(id, cam_id, bucket_name, vid_key, best_frame_key, evidence, store, state, 
		 needs_publish, retention_days, start_ts, end_ts)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		ON CONFLICT (id) DO UPDATE SET
		cam_id = EXCLUDED.cam_id,
		bucket_name = EXCLUDED.bucketname,
		vid_key = EXCLUDED.vid_key,
		best_frame_key = EXCLUDED.best_frame_key,
		evidence = EXCLUDED.evidence,
		store = EXCLUDED.store,
		state = EXCLUDED.state,
		needs_publish = EXCLUDED.needs_publish
		retention_days = EXCLUDED.retention_days,
		start_ts = EXCLUDED.start_ts,
		end_ts = EXCLUDED.end_ts
	`)

	return nil
}
