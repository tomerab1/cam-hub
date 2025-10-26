package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/go-co-op/gocron/v2"
	"github.com/redis/go-redis/v9"
	v1 "tomerab.com/cam-hub/internal/contracts/v1"
	"tomerab.com/cam-hub/internal/mtxapi"
	"tomerab.com/cam-hub/internal/onvif"
	"tomerab.com/cam-hub/internal/repos"
)

const (
	hydrateUpdatedAddress uint8 = iota
	hydrateNewDevice
	hydrateErr
	hydrateNone
)

type DiscoveryService struct {
	Rdb              repos.RedisIface
	CamerasRepo      repos.CameraRepoIface
	MtxClient        *mtxapi.MtxClient
	Sched            gocron.Scheduler
	Logger           *slog.Logger
	SseChan          chan v1.DiscoveryEvent
	CamsProxyEventCh chan v1.CameraProxyEvent
}

func (svc *DiscoveryService) InitJobs(ctx context.Context) error {
	job, err := svc.Sched.NewJob(
		gocron.DurationJob(time.Minute),
		gocron.NewTask(func() {
			runCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			svc.Logger.Info("Running discovery tick")
			svc.Discover(runCtx)
		}),
		gocron.WithSingletonMode(gocron.LimitModeWait),
		gocron.JobOption(gocron.WithStartImmediately()),
	)

	if err != nil {
		return err
	}

	svc.Logger.Info("Scheduled discovery", "jobid", job.ID())
	return nil
}

func (svc *DiscoveryService) Discover(ctx context.Context) {
	matches := onvif.DiscoverNewCameras(ctx, svc.Logger)
	svc.Logger.Info("found matches", "matches", matches)

	for _, match := range matches.Matches {
		status, version := svc.hydrateDb(ctx, match.UUID, match.Xaddr)
		if err := svc.updateCache(ctx, match.UUID, match.Xaddr); err != nil {
			svc.Logger.Error(err.Error())
		}

		switch status {
		case hydrateUpdatedAddress:
			svc.SseChan <- v1.DiscoveryEvent{
				Type: "device_ip_changed",
				UUID: match.UUID,
				Addr: match.Xaddr,
				At:   time.Now(),
			}

			_, err := svc.MtxClient.Publish(ctx, match.UUID)
			if err != nil {
				svc.Logger.Warn("failed to publish stream to mediamtx", "newDevice", true, "err", err)
				continue
			}

			svc.CamsProxyEventCh <- v1.CameraProxyEvent{
				CameraPairedEvent: &v1.CameraPairedEvent{
					UUID:      match.UUID,
					StreamUrl: fmt.Sprintf("rtsp://localhost:8554/%s", match.UUID),
					Revision:  version,
				},
			}
		case hydrateNewDevice:
			svc.SseChan <- v1.DiscoveryEvent{
				Type: "device_new",
				UUID: match.UUID,
				Addr: match.Xaddr,
				At:   time.Now(),
			}
		default:
			_, err := svc.MtxClient.Publish(ctx, match.UUID)
			if err != nil {
				svc.Logger.Warn("failed to publish stream to mediamtx", "newDevice", true, "err", err)
				continue
			}

			svc.CamsProxyEventCh <- v1.CameraProxyEvent{
				CameraPairedEvent: &v1.CameraPairedEvent{
					UUID:      match.UUID,
					StreamUrl: fmt.Sprintf("rtsp://localhost:8554/%s", match.UUID),
					Revision:  version,
				},
			}
			if os.Getenv("ENV_TYPE") == "dev" {
				// For testing the ui.
				svc.SseChan <- v1.DiscoveryEvent{
					Type: "device_new",
					UUID: match.UUID,
					Addr: match.Xaddr,
					At:   time.Now(),
				}
			}
		}
	}
}

func (svc *DiscoveryService) hydrateDb(ctx context.Context, uuid string, addr string) (uint8, int) {
	cam, err := svc.CamerasRepo.FindOne(ctx, uuid)
	if pgxscan.NotFound(err) {
		return hydrateNewDevice, cam.Version
	}
	if err != nil {
		svc.Logger.Error(err.Error())
		return hydrateErr, -1
	}

	// if camera is already paired and the address has changed.
	if cam.Addr != addr {
		cam.Addr = addr
		if err := svc.CamerasRepo.Save(ctx, cam); err != nil {
			return hydrateErr, -1
		}
		return hydrateUpdatedAddress, cam.Version
	}

	return hydrateNone, cam.Version
}

func (svc *DiscoveryService) updateCache(ctx context.Context, uuid string, addr string) error {
	storedAddr, err := svc.Rdb.Get(ctx, "cam:"+uuid)
	if err != nil && !errors.Is(err, redis.Nil) {
		return err
	}
	if errors.Is(err, redis.Nil) || storedAddr != addr {
		return svc.Rdb.Set(ctx, "cam:"+uuid, addr, 0)
	}

	return nil
}
