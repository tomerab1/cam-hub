package repos

import (
	"context"

	"github.com/georgysavva/scany/v2/pgxscan"
	"tomerab.com/cam-hub/internal/api/v1/models"
)

type RecordingsRepoIface interface {
	Upsert(ctx context.Context, rec *models.Recordings) (*models.Recordings, error)
}

type PgxRecordingsRepo struct {
	DB DBPoolIface
}

func NewPgxRecordingsRepo(db DBPoolIface) *PgxRecordingsRepo {
	return &PgxRecordingsRepo{
		DB: db,
	}
}

func (repo *PgxRecordingsRepo) Upsert(ctx context.Context, rec *models.Recordings) (*models.Recordings, error) {
	var out models.Recordings

	q := `
		INSERT INTO recordings (
			cam_id, bucket_name, vid_key, best_frame_key,
			evidence, score, state, needs_publish,
			retention_days, start_ts, end_ts)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
			ON CONFLICT (vid_key) DO UPDATE SET
				cam_id         = EXCLUDED.cam_id,
				bucket_name    = EXCLUDED.bucket_name,
				best_frame_key = EXCLUDED.best_frame_key,
				evidence       = EXCLUDED.evidence,
				score          = EXCLUDED.score,
				state          = EXCLUDED.state,
				needs_publish  = EXCLUDED.needs_publish,
				retention_days = EXCLUDED.retention_days,
				start_ts       = EXCLUDED.start_ts,
				end_ts         = EXCLUDED.end_ts,
				promoted_at    = timezone('UTC', now())::timestamp
			RETURNING *;
			`
	if err := pgxscan.Get(ctx, repo.DB, &out, q,
		rec.CamUUID,
		rec.BucketName,
		rec.VidBucketKey,
		rec.BestFrameBucketKey,
		rec.Evidence,
		rec.Score,
		rec.State,
		rec.NeedsPublish,
		rec.RetentionDays,
		rec.StartTs,
		rec.EndTs,
	); err != nil {
		return nil, err
	}

	return &out, nil
}
