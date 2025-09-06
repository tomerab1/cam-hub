package services

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/redis/go-redis/v9"
	v1 "tomerab.com/cam-hub/internal/contracts/v1"
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
	Rdb         repos.RedisIface
	CamerasRepo repos.CameraRepoIface
	Sched       gocron.Scheduler
	Logger      *slog.Logger
	SseChan     chan v1.DiscoveryEvent
}

func (svc *DiscoveryService) InitJobs(ctx context.Context) error {
	job, err := svc.Sched.NewJob(
		gocron.DurationJob(1*time.Minute),
		gocron.NewTask(func() {
			runCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()
			svc.Logger.Info("Running discovery tick")
			svc.Discover(runCtx)
		}),
		gocron.WithSingletonMode(gocron.LimitModeWait),
	)

	if err != nil {
		return err
	}

	svc.Logger.Info("Scheduled discovery", "jobid", job.ID())
	return nil
}

func (svc *DiscoveryService) Discover(ctx context.Context) {
	matches := onvif.DiscoverNewCameras(svc.Logger)

	for _, match := range matches.Matches {
		status := svc.hydrateDb(ctx, match.UUID, match.Xaddr)
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
		case hydrateNewDevice:
			svc.SseChan <- v1.DiscoveryEvent{
				Type: "device_new",
				UUID: match.UUID,
				Addr: match.Xaddr,
				At:   time.Now(),
			}
		default:
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

func (svc *DiscoveryService) hydrateDb(ctx context.Context, uuid string, addr string) uint8 {
	cam, err := svc.CamerasRepo.FindOne(ctx, uuid)
	if err != nil {
		svc.Logger.Error(err.Error())
		return hydrateErr
	}

	if cam == nil {
		return hydrateNewDevice
	}

	// if camera is already paired and the address has changed.
	if cam.Addr != addr {
		cam.Addr = addr
		svc.CamerasRepo.Save(ctx, cam)
		return hydrateUpdatedAddress
	}

	return hydrateNone
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
