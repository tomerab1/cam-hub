package services

import (
	"context"
	"log/slog"

	"tomerab.com/cam-hub/internal/api/v1/models"
	v1 "tomerab.com/cam-hub/internal/contracts/v1"
	"tomerab.com/cam-hub/internal/onvif"
	"tomerab.com/cam-hub/internal/onvif/device"
	"tomerab.com/cam-hub/internal/repos"
)

const (
	UserLvlAdmin = "Administrator"
)

type CameraService struct {
	CamRepo      *repos.PgxCameraRepo
	CamCredsRepo *repos.PgxCameraCredsRepo
	Logger       *slog.Logger
}

func (svc *CameraService) Pair(ctx context.Context, req v1.PairDeviceReq) (*models.Camera, error) {
	client, err := onvif.NewOnvifClient(onvif.OnvifClientParams{
		Xaddr:    req.Addr,
		Username: req.Username,
		Password: req.Password,
		Logger:   svc.Logger,
	})
	if err != nil {
		return nil, err
	}

	info, err := client.GetDeviceInfo()
	if err != nil {
		return nil, err
	}

	err = client.CreateUser(device.CreateUserDto{
		Username:  req.Username,
		Password:  req.Password,
		UserLevel: UserLvlAdmin,
	})

	if err != nil {
		svc.Logger.Debug(err.Error())
	}

	tx, err := svc.CamRepo.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	cam := &models.Camera{
		UUID:            req.UUID,
		Addr:            req.Addr,
		CameraName:      req.CameraName,
		HardwareId:      info.HardwareId,
		Model:           info.Model,
		Manufacturer:    info.Manufacturer,
		FirmwareVersion: info.FirmwareVersion,
		SerialNumber:    info.SerialNumber,
		IsPaired:        true,
	}
	err = svc.CamRepo.UpsertCamera(ctx, tx, cam)

	if err != nil {
		return nil, err
	}

	err = svc.CamCredsRepo.InsertCreds(ctx, tx, &models.CameraCreds{
		UUID:     req.UUID,
		Username: req.Username,
		Password: req.Password,
	})

	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return cam, nil
}

func (svc *CameraService) Unpair(ctx context.Context, req v1.UnpairDeviceReq) error {
	cam, err := svc.CamRepo.FindOne(ctx, req.UUID)
	if err != nil {
		return err
	}

	cam.IsPaired = false
	return svc.CamRepo.Save(ctx, cam)
}
