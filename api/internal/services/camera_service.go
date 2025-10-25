package services

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"tomerab.com/cam-hub/internal/api/v1/models"
	v1 "tomerab.com/cam-hub/internal/contracts/v1"
	dvripclient "tomerab.com/cam-hub/internal/dvrip"
	inmemory "tomerab.com/cam-hub/internal/events/in_memory"
	"tomerab.com/cam-hub/internal/mtxapi"
	"tomerab.com/cam-hub/internal/onvif"
	"tomerab.com/cam-hub/internal/onvif/device"
	"tomerab.com/cam-hub/internal/repos"
)

const (
	UserLvlAdmin = "Administrator"
)

type CameraService struct {
	CamRepo            repos.CameraRepoIface
	CamCredsRepo       repos.CameraCredsRepoIface
	Rdms               repos.RedisIface
	InMemCache         *inmemory.InMemoryPubSub
	MtxClient          *mtxapi.MtxClient
	CamsEventProxyChan chan v1.CameraProxyEvent
	Logger             *slog.Logger
}

func (svc *CameraService) UpairCamera(ctx context.Context, uuid string) error {
	creds, err := svc.CamCredsRepo.FindOne(ctx, uuid)
	if err != nil {
		return err
	}
	cam, err := svc.CamRepo.FindOne(ctx, uuid)
	if err != nil {
		return err
	}

	client, err := dvripclient.New(
		getAddrWithoutPort(cam.Addr),
		os.Getenv("CAMERA_GLOB_ADMIN_USERNAME"),
		os.Getenv("CAMERA_GLOB_ADMIN_PASS"))
	if err != nil {
		return err
	}

	return client.DelUser(creds.Username)
}

func getAddrWithoutPort(addr string) string {
	return strings.TrimSuffix(addr, ":")
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
	if err := svc.tryCreateRootUser(client); err != nil {
		svc.Logger.Debug("onvif err", "err", err)
	}

	if err := svc.tryCreateUser(client, req); err != nil {
		svc.Logger.Debug("onvif err", "err", err)
	}

	return &info, nil
}

func (svc *CameraService) tryCreateRootUser(client *onvif.OnvifClient) error {
	return client.CreateUser(device.CreateUserDto{
		Username:  os.Getenv("CAMERA_GLOB_ADMIN_USERNAME"),
		Password:  os.Getenv("CAMERA_GLOB_ADMIN_PASS"),
		UserLevel: UserLvlAdmin,
	})
}

func (svc *CameraService) tryCreateUser(client *onvif.OnvifClient, req v1.PairDeviceReq) error {
	return client.CreateUser(device.CreateUserDto{
		Username:  req.Username,
		Password:  req.Password,
		UserLevel: UserLvlAdmin,
	})
}

func (svc *CameraService) buildCameraModel(uuid string, req v1.PairDeviceReq, info *device.GetDeviceInfoDto) *models.Camera {
	return &models.Camera{
		UUID:            uuid,
		Addr:            req.Addr,
		CameraName:      req.CameraName,
		HardwareId:      info.HardwareId,
		Model:           info.Model,
		Manufacturer:    info.Manufacturer,
		FirmwareVersion: info.FirmwareVersion,
		SerialNumber:    info.SerialNumber,
	}
}

func (svc *CameraService) storeCameraAndCredentials(ctx context.Context, camera *models.Camera, uuid string, req v1.PairDeviceReq) error {
	tx, err := svc.CamRepo.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := svc.CamRepo.UpsertCameraTx(ctx, tx, camera); err != nil {
		return fmt.Errorf("failed to upsert camera: %w", err)
	}

	creds := &models.CameraCreds{
		UUID:     uuid,
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

func (svc *CameraService) Pair(ctx context.Context, uuid string, req v1.PairDeviceReq) (*models.Camera, error) {
	devInfo, err := svc.connectAndGetDeviceInfo(req)

	if err != nil {
		return nil, fmt.Errorf("failed to connect to device: %w", err)
	}

	camera := svc.buildCameraModel(uuid, req, devInfo)
	err = svc.storeCameraAndCredentials(ctx, camera, uuid, req)
	if err != nil {
		return nil, fmt.Errorf("failed to store camera data: %w", err)
	}

	if err := svc.connectCameraToWifi(
		req.Addr,
		req.WifiName,
		req.WifiPassword,
	); err != nil {
		return nil, err
	}

	streamUrl, err := svc.MtxClient.Publish(ctx, uuid)
	if err != nil {
		return nil, err
	}

	svc.CamsEventProxyChan <- v1.CameraProxyEvent{
		CameraPairedEvent: &v1.CameraPairedEvent{
			UUID:      uuid,
			StreamUrl: streamUrl,
			Revision:  camera.Version,
		},
	}

	svc.Logger.Info("Camera paired successfully", "uuid", uuid, "addr", req.Addr)
	return camera, nil
}

func (svc *CameraService) connectCameraToWifi(addr, ssid, psk string) error {
	parts := strings.Split(addr, ":")
	addrWithoutPort := parts[0]

	client, err := dvripclient.New(
		addrWithoutPort,
		os.Getenv("CAMERA_GLOB_ADMIN_USERNAME"),
		os.Getenv("CAMERA_GLOB_ADMIN_PASS"),
	)

	if err != nil {
		return err
	}
	defer client.Close()

	return client.PairWifi(ssid, psk)
}

func (svc *CameraService) Unpair(ctx context.Context, uuid string) error {
	cam, err := svc.CamRepo.FindOne(ctx, uuid)
	if err != nil {
		return err
	}

	camAddrWithoutPort := strings.Split(cam.Addr, ":")[0]
	dvripClient, err := dvripclient.New(
		camAddrWithoutPort,
		os.Getenv("CAMERA_GLOB_ADMIN_USERNAME"),
		os.Getenv("CAMERA_GLOB_ADMIN_PASS"))
	if err != nil {
		return err
	}
	defer dvripClient.Close()

	creds, err := svc.CamCredsRepo.FindOne(ctx, uuid)
	if err != nil {
		return err
	}

	if err := dvripClient.DelUser(creds.Username); err != nil {
		return err
	}

	if err := svc.MtxClient.Delete(ctx, uuid); err != nil {
		return err
	}

	if err := svc.purgeCameraFromDataSources(ctx, uuid); err != nil {
		return err
	}

	svc.CamsEventProxyChan <- v1.CameraProxyEvent{
		CameraUnpairedEvent: &v1.CameraUnpairedEvent{
			UUID: uuid,
		},
	}

	return dvripClient.Reboot()
}

func (svc *CameraService) purgeCameraFromDataSources(ctx context.Context, uuid string) error {
	svc.InMemCache.Purge(uuid)
	if err := svc.Rdms.Del(ctx, fmt.Sprintf("cam:%s", uuid)); err != nil {
		return fmt.Errorf("failed to delete from redis: %w", err)
	}

	return svc.CamRepo.Delete(ctx, uuid)
}

func (svc *CameraService) GetCameras(ctx context.Context, offset, limit int) ([]*models.Camera, error) {
	return svc.CamRepo.FindMany(ctx, offset, limit)
}

func (svc *CameraService) GetAllUUIDS(ctx context.Context) ([]string, error) {
	return svc.CamRepo.FindAllUUIDS(ctx)
}

func (svc *CameraService) CameraExists(ctx context.Context, uuid string) bool {
	_, err := svc.CamRepo.FindOne(ctx, uuid)
	return err == nil
}
