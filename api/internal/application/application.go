package application

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"gopkg.in/lumberjack.v3"
	v1 "tomerab.com/cam-hub/internal/contracts/v1"
	"tomerab.com/cam-hub/internal/events"
	inmemory "tomerab.com/cam-hub/internal/events/in_memory"
	"tomerab.com/cam-hub/internal/mtxapi"
	"tomerab.com/cam-hub/internal/services"
)

type Application struct {
	Logger             *slog.Logger
	DB                 *pgxpool.Pool
	HttpClient         *http.Client
	CameraService      *services.CameraService
	DiscoveryService   *services.DiscoveryService
	PtzService         *services.PtzService
	MtxClient          *mtxapi.MtxClient
	Bus                events.BusIface
	PubSub             *inmemory.InMemoryPubSub
	SseChan            chan v1.DiscoveryEvent
	CamsEventProxyChan chan v1.CameraProxyEvent
	LogSink            lumberjack.Writer
}

func (app *Application) OnStartup(ctx context.Context) {
	go app.publishPairedCamsJob(ctx)
}

func (app *Application) WriteJSON(w http.ResponseWriter, r *http.Request, data any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		app.Logger.Warn("Error writing json response", "err", err)
	}
}

func (app *Application) publishPairedCamsJob(ctx context.Context) error {
	var (
		keyPair   = os.Getenv("RABBITMQ_PAIR_KEY")
		keyUnpair = os.Getenv("RABBITMQ_UNPAIR_KEY")
	)

	for {
		select {
		case <-ctx.Done():
			return nil
		case ev := <-app.CamsEventProxyChan:
			if ev.CameraPairedEvent != nil {
				if err := app.publishEvent(ctx, keyPair, &ev); err != nil {
					return err
				}
			}
			if ev.CameraUnpairedEvent != nil {
				if err := app.publishEvent(ctx, keyUnpair, &ev); err != nil {
					return err
				}
			}
		}
	}
}

func (app *Application) publishEvent(ctx context.Context, key string, ev *v1.CameraProxyEvent) error {
	var (
		keyPair   = os.Getenv("RABBITMQ_PAIR_KEY")
		keyUnpair = os.Getenv("RABBITMQ_UNPAIR_KEY")
	)
	var bytes []byte
	var err error

	switch key {
	case keyPair:
		bytes, err = json.Marshal(ev.CameraPairedEvent)
	case keyUnpair:
		bytes, err = json.Marshal(ev.CameraUnpairedEvent)
	}

	if err != nil {
		return err
	}

	return app.Bus.Publish(ctx, "", key, bytes, nil)
}
