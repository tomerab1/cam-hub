package services

import (
	"context"
	"fmt"
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
	CamRepo      repos.CameraRepoIface
	CamCredsRepo repos.CameraCredsRepoIface
	Logger       *slog.Logger
}

func (svc *CameraService) connectAndGetDeviceInfo(req v1.PairDeviceReq) (*device.GetDeviceInfoDto, error) {
	client, err := onvif.NewOnvifClient(onvif.OnvifClientParams{
		Xaddr:    req.Addr,
		Username: req.Username,
		Password: req.Password,
		Logger:   svc.Logger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create ONVIF client: %w", err)
	}

	info, err := client.GetDeviceInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get device info: %w", err)
	}

	// TODO(tomer): Check failure reason, if critical return err to user.
	if err := svc.tryCreateUser(client, req); err != nil {
		svc.Logger.Warn("Failed to create user on device", "error", err, "uuid", req.UUID)
	}

	return &info, nil
}

func (svc *CameraService) tryCreateUser(client *onvif.OnvifClient, req v1.PairDeviceReq) error {
	return client.CreateUser(device.CreateUserDto{
		Username:  req.Username,
		Password:  req.Password,
		UserLevel: UserLvlAdmin,
	})
}

func (svc *CameraService) buildCameraModel(req v1.PairDeviceReq, info *device.GetDeviceInfoDto) *models.Camera {
	return &models.Camera{
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
}

func (svc *CameraService) storeCameraAndCredentials(ctx context.Context, camera *models.Camera, req v1.PairDeviceReq) error {
	tx, err := svc.CamRepo.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := svc.CamRepo.UpsertCamera(ctx, tx, camera); err != nil {
		return fmt.Errorf("failed to upsert camera: %w", err)
	}

	creds := &models.CameraCreds{
		UUID:     req.UUID,
		Username: req.Username,
		Password: req.Password,
	}
	if err := svc.CamCredsRepo.InsertCreds(ctx, tx, creds); err != nil {
		return fmt.Errorf("failed to insert credentials: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (svc *CameraService) Pair(ctx context.Context, req v1.PairDeviceReq) (*models.Camera, error) {
	devInfo, err := svc.connectAndGetDeviceInfo(req)

	if err != nil {
		return nil, fmt.Errorf("failed to connect to device: %w", err)
	}

	camera := svc.buildCameraModel(req, devInfo)

	err = svc.storeCameraAndCredentials(ctx, camera, req)
	if err != nil {
		return nil, fmt.Errorf("failed to store camera data: %w", err)
	}

	svc.Logger.Info("Camera paired successfully", "uuid", req.UUID, "addr", req.Addr)
	return camera, nil
}

func (svc *CameraService) Unpair(ctx context.Context, req v1.UnpairDeviceReq) error {
	cam, err := svc.CamRepo.FindOne(ctx, req.UUID)
	if err != nil {
		return err
	}

	cam.IsPaired = false
	return svc.CamRepo.Save(ctx, cam)
}
