package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"tomerab.com/cam-hub/internal/api/v1/models"
	v1 "tomerab.com/cam-hub/internal/contracts/v1"
	"tomerab.com/cam-hub/internal/repos"
)

type RecordingState = string

const (
	StatePromoted  RecordingState = "promoted"
	StateDiscarded RecordingState = "discarded"
)

var (
	UpsertFailed = errors.New("recordings: failed to upsert recording")
)

type RecordingsService struct {
	recordingsRepo repos.RecordingsRepoIface
	camerasRepo    repos.CameraRepoIface
	logger         *slog.Logger
}

func NewRecordingsService(
	logger *slog.Logger,
	recordingsRepo repos.RecordingsRepoIface,
	camsRepo repos.CameraRepoIface) *RecordingsService {
	return &RecordingsService{
		logger:         logger,
		recordingsRepo: recordingsRepo,
		camerasRepo:    camsRepo,
	}
}

func (svc *RecordingsService) Upsert(
	ctx context.Context,
	camUUID string,
	state RecordingState,
	req v1.AddRecordingReq) (*models.Recordings, error) {
	if _, err := svc.camerasRepo.FindOne(ctx, camUUID); err != nil {
		return nil, fmt.Errorf("camera (%s) does not exist", camUUID)
	}

	needsPublish := false
	if state == StatePromoted {
		needsPublish = true
	}

	rec, err := svc.recordingsRepo.Upsert(ctx, &models.Recordings{
		CamUUID:            camUUID,
		BucketName:         req.BucketName,
		VidBucketKey:       req.VidBucketKey,
		BestFrameBucketKey: req.BestFrameBucketKey,
		Evidence: models.Evidence{
			Conf: req.Evidence.Conf,
			Xmin: req.Evidence.Xmin,
			Xmax: req.Evidence.Xmax,
			Ymin: req.Evidence.Ymin,
			Ymax: req.Evidence.Ymax,
		},
		Score:         req.Score,
		State:         state,
		NeedsPublish:  needsPublish,
		RetentionDays: req.RetentionDays,
		StartTs:       req.StartTs,
		EndTs:         req.EndTs,
	})
	if err != nil {
		return nil, UpsertFailed
	}

	return rec, nil
}
