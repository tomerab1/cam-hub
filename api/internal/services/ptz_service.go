package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"tomerab.com/cam-hub/internal/api/v1/models"
	v1 "tomerab.com/cam-hub/internal/contracts/v1"
	"tomerab.com/cam-hub/internal/onvif"
	"tomerab.com/cam-hub/internal/onvif/ptz"
	"tomerab.com/cam-hub/internal/repos"
)

const (
	ptzCacheKeyPrefix = "ptz:token:"
	ptzCacheTTL       = time.Hour
)

type PtzService struct {
	CamRepo      repos.CameraRepoIface
	CamCredsRepo repos.CameraCredsRepoIface
	PtzTokenRepo repos.PtzTokenRepoIface
	Rdb          repos.RedisIface
	Logger       *slog.Logger
}

func (svc *PtzService) MoveCamera(ctx context.Context, uuid string, move v1.MoveCameraReq) error {
	token, err := svc.resolvePtzToken(ctx, uuid)
	if err != nil {
		return err
	}

	cam, err := svc.CamRepo.FindOne(ctx, uuid)
	if err != nil {
		return fmt.Errorf("camera not found: %w", err)
	}

	creds, err := svc.CamCredsRepo.FindOne(ctx, uuid)
	if err != nil {
		return fmt.Errorf("camera creds not found: %w", err)
	}

	client, err := onvif.NewOnvifClient(onvif.OnvifClientParams{
		Xaddr:    cam.Addr,
		Username: creds.Username,
		Password: creds.Password,
		Logger:   svc.Logger,
	})
	if err != nil {
		return fmt.Errorf("onvif client: %w", err)
	}

	if err := client.MoveCamera(ptz.MoveCameraDto{
		Token:       token,
		Translation: move.Translation,
	}); err != nil {
		if isInvalidToken(err) {
			newTok, rerr := client.GetPtzProfile()
			if rerr != nil {
				return fmt.Errorf("refresh ptz token: %w", rerr)
			}
			if uerr := svc.upsertAndCache(ctx, uuid, newTok.Token); uerr != nil {
				svc.Logger.Warn("failed to upsert/cache refreshed PTZ token", "err", uerr)
			}
			return client.MoveCamera(ptz.MoveCameraDto{
				Token:       token,
				Translation: move.Translation,
			})
		}
		return err
	}
	return nil
}

func (svc *PtzService) resolvePtzToken(ctx context.Context, uuid string) (string, error) {
	if svc.Rdb != nil {
		if tok, err := svc.Rdb.Get(ctx, ptzCacheKeyPrefix+uuid); err == nil && tok != "" {
			return tok, nil
		}
	}

	if rec, err := svc.PtzTokenRepo.FindOne(ctx, uuid); err == nil && rec != nil && rec.Token != "" {
		_ = svc.cacheToken(ctx, uuid, rec.Token)
		return rec.Token, nil
	}

	cam, err := svc.CamRepo.FindOne(ctx, uuid)
	if err != nil {
		return "", fmt.Errorf("camera not found: %w", err)
	}

	creds, err := svc.CamCredsRepo.FindOne(ctx, uuid)
	if err != nil {
		return "", fmt.Errorf("camera creds not found: %w", err)
	}

	client, err := onvif.NewOnvifClient(onvif.OnvifClientParams{
		Xaddr:    cam.Addr,
		Username: creds.Username,
		Password: creds.Password,
		Logger:   svc.Logger,
	})
	if err != nil {
		return "", fmt.Errorf("onvif client: %w", err)
	}

	ptzTok, err := client.GetPtzProfile()
	if err != nil {
		return "", err
	}

	if err := svc.upsertAndCache(ctx, uuid, ptzTok.Token); err != nil {
		svc.Logger.Warn("failed to upsert/cache PTZ token", "err", err)
	}

	return ptzTok.Token, nil
}

func (svc *PtzService) upsertAndCache(ctx context.Context, uuid, token string) error {
	if token == "" {
		return errors.New("empty ptz token")
	}

	if err := svc.PtzTokenRepo.UpsertToken(ctx, &models.PtzToken{UUID: uuid, Token: token}); err != nil {
		return err
	}

	return svc.cacheToken(ctx, uuid, token)
}

func (svc *PtzService) cacheToken(ctx context.Context, uuid, token string) error {
	if svc.Rdb == nil {
		return nil
	}

	return svc.Rdb.Set(ctx, ptzCacheKeyPrefix+uuid, token, ptzCacheTTL)
}

func isInvalidToken(err error) bool {
	if err == nil {
		return false
	}

	s := err.Error()
	return strings.Contains(s, "Invalid") && strings.Contains(s, "Token")
}
