package frameanalyzer

import (
	"context"
	"os"
	"syscall"

	"github.com/joho/godotenv"
	"gopkg.in/lumberjack.v3"
	"tomerab.com/cam-hub/internal/events/rabbitmq"
	frameanalyzer "tomerab.com/cam-hub/internal/frame_analyzer"
	"tomerab.com/cam-hub/internal/utils"
)

func main() {
	if err := godotenv.Load(); err != nil {
		panic(err.Error())
	}

	fileHandler, err := lumberjack.New(
		lumberjack.WithFileName(os.Getenv("LOGGER_PATH")+"/frame_analyzer.log"),
		lumberjack.WithMaxBytes(25*lumberjack.MB),
		lumberjack.WithMaxDays(14),
		lumberjack.WithCompress(),
	)
	if err != nil {
		panic(err.Error())
	}

	bus, err := rabbitmq.NewBus(os.Getenv("RABBITMQ_ADDR"))
	if err != nil {
		panic(err.Error())
	}

	ctx, cancel := utils.GracefullShutdown(context.Background(), func() {}, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	analyzer := frameanalyzer.New(fileHandler, bus)
	analyzer.Run(ctx)
}
