package repos

import (
	"context"
	"testing"

	"github.com/pashagolub/pgxmock/v4"
	"tomerab.com/cam-hub/internal/api/v1/models"
)

func makeRecording(id string, cameraUUID string) *models.Recordings {
	return &models.Recordings{
		Id:            id,
		CamUUID:       cameraUUID,
		BucketName:    "media",
		VidBucketKey:  "cam/" + cameraUUID + "/2025-09-30T12-00-00.mp4",
		Evidence:      models.Evidence{Conf: 0.9, Xmin: 0, Ymin: 0, Xmax: 1, Ymax: 1},
		Score:         0.9,
		State:         "promoted",
		NeedsPublish:  true,
		RetentionDays: 14,
	}
}
func setupRecordingsRepoTest(t *testing.T) (*PgxRecordingsRepo, pgxmock.PgxPoolIface, context.Context) {
	t.Helper()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock pool: %v", err)
	}

	t.Cleanup(func() {
		mock.Close()
	})

	repo := NewPgxRecordingsRepo(mock)
	ctx := context.Background()

	return repo, mock, ctx
}

func TestUpsert(t *testing.T) {
	repo, mock, ctx := setupRecordingsRepoTest(t)

	t.Run("insert new recording - should succeed", func(t *testing.T) {
		recording := makeRecording("1", "1")
		mock.ExpectQuery(`INSERT INTO recordings(?s).*`).
			WithArgs(
				recording.CamUUID,
				recording.BucketName,
				recording.VidBucketKey,
				recording.BestFrameBucketKey,
				recording.Evidence,
				recording.Score,
				recording.State,
				recording.NeedsPublish,
				recording.RetentionDays,
				recording.StartTs,
				recording.EndTs,
			).
			WillReturnRows(
				pgxmock.NewRows([]string{
					"id",
					"cam_id",
					"bucket_name",
					"vid_key",
					"best_frame_key",
					"evidence",
					"score",
					"state",
					"needs_publish",
					"promoted_at",
					"retention_days",
					"start_ts",
					"end_ts",
				}).AddRow(
					"generated-uuid",
					recording.CamUUID,
					recording.BucketName,
					recording.VidBucketKey,
					recording.BestFrameBucketKey,
					[]byte(`{"conf":0.9,"x_min":0,"y_min":0,"x_max":1,"y_max":1}`),
					recording.Score,
					recording.State,
					recording.NeedsPublish,
					recording.PromotedAt,
					recording.RetentionDays,
					recording.StartTs,
					recording.EndTs,
				),
			)

		if _, err := repo.Upsert(ctx, recording); err != nil {
			t.Errorf("upsert failed: %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unfulfilled expectations: %v", err)
		}
	})

	t.Run("update new recording - should succeed", func(t *testing.T) {
		recording := makeRecording("1", "1")
		mock.ExpectQuery(`INSERT INTO recordings(?s).*`).
			WithArgs(
				recording.CamUUID,
				recording.BucketName,
				recording.VidBucketKey,
				recording.BestFrameBucketKey,
				recording.Evidence,
				recording.Score,
				recording.State,
				recording.NeedsPublish,
				recording.RetentionDays,
				recording.StartTs,
				recording.EndTs,
			).
			WillReturnRows(
				pgxmock.NewRows([]string{
					"id",
					"cam_id",
					"bucket_name",
					"vid_key",
					"best_frame_key",
					"evidence",
					"score",
					"state",
					"needs_publish",
					"promoted_at",
					"retention_days",
					"start_ts",
					"end_ts",
				}).AddRow(
					"generated-uuid",
					recording.CamUUID,
					recording.BucketName,
					recording.VidBucketKey,
					recording.BestFrameBucketKey,
					[]byte(`{"conf":0.9,"x_min":0,"y_min":0,"x_max":1,"y_max":1}`),
					recording.Score,
					recording.State,
					recording.NeedsPublish,
					recording.PromotedAt,
					recording.RetentionDays,
					recording.StartTs,
					recording.EndTs,
				),
			)

		if _, err := repo.Upsert(ctx, recording); err != nil {
			t.Errorf("upsert failed: %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unfulfilled expectations: %v", err)
		}
	})
}
