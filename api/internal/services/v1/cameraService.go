package v1

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"tomerab.com/cam-hub/internal/api/v1/models"
	v1 "tomerab.com/cam-hub/internal/contracts/v1"
	"tomerab.com/cam-hub/internal/onvif"
)

type CameraService struct {
	DB     *pgxpool.Pool
	Logger *slog.Logger
}

func (camService *CameraService) Pair(ctx context.Context, req v1.PairDeviceReq) (*models.Camera, error) {
	client, err := onvif.NewOnvifClient(onvif.OnvifClientParams{
		Xaddr:    req.Addr,
		Username: req.Username,
		Password: req.Password,
		Logger:   camService.Logger,
	})
	if err != nil {
		return nil, err
	}

	info, err := client.GetDeviceInfo()
	if err != nil {
		return nil, err
	}

	tx, err := camService.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO cameras
			(id, name, manufacturer, model, firmwareVersion, serialNumber, hardwareId, addr)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			manufacturer = EXCLUDED.manufacturer,
			model = EXCLUDED.model,
			firmwareVersion = EXCLUDED.firmwareVersion,
			serialNumber = EXCLUDED.serialNumber,
			hardwareId = EXCLUDED.hardwareId,
			addr = EXCLUDED.addr
	`,
		req.UUID,
		req.CameraName,
		info.Manufacturer,
		info.Model,
		info.FirmwareVersion,
		info.SerialNumber,
		info.HardwareId,
		req.Addr,
	)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO camera_creds (id, username, password)
		VALUES ($1,$2,$3)
		ON CONFLICT (id) DO UPDATE SET
			username = EXCLUDED.username,
			password = EXCLUDED.password
	`,
		req.UUID, req.Username, req.Password,
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &models.Camera{
		UUID:            req.UUID,
		CameraName:      req.CameraName,
		Addr:            req.Addr,
		Manufacturer:    info.Manufacturer,
		FirmwareVersion: info.FirmwareVersion,
		SerialNumber:    info.SerialNumber,
		Model:           info.Model,
		HardwareId:      info.HardwareId,
	}, nil
}
