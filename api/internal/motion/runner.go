package motion

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"time"

	"golang.org/x/sync/errgroup"
	v1 "tomerab.com/cam-hub/internal/contracts/v1"
	"tomerab.com/cam-hub/internal/events"
	objectstorage "tomerab.com/cam-hub/internal/object_storage"
)

var (
	bucketName = os.Getenv("MINIO_BUCKET_NAME")
)

type MotionCtx struct {
	UUID      string
	Score     int
	TimePoint time.Time
}

type Runner struct {
	logger      *slog.Logger
	ctx         context.Context
	bus         events.BusIface
	minioClient *objectstorage.MinIOStore
	sem         chan struct{}
}

func NewRunner(ctx context.Context, minioClient *objectstorage.MinIOStore, bus events.BusIface, logger *slog.Logger, maxJobs int) *Runner {
	logger.Info("new runner created")
	return &Runner{
		logger:      logger,
		ctx:         ctx,
		bus:         bus,
		minioClient: minioClient,
		sem:         make(chan struct{}, maxJobs),
	}
}

func (runner *Runner) PostJob(ctx MotionCtx) error {
	runner.sem <- struct{}{}
	defer func() { <-runner.sem }()
	return runner.process(ctx)
}

func (runner *Runner) process(ctx MotionCtx) error {
	parts, err := onMotion(ctx.UUID, ctx.Score, ctx.TimePoint)
	if err != nil {
		return err
	}

	outFileName := fmt.Sprintf("motion_%s_%s.mp4", ctx.UUID, ctx.TimePoint.Format("2006-01-02_15-04-05"))
	runner.logger.Info("File part", "parts", parts.files)
	if err := concatVideoFiles(runner.ctx, runner.logger, outFileName, parts.files); err != nil {
		return err
	}
	runner.logger.Info("Created new file", "path", outFileName)

	framePaths, err := extractFrames(runner.ctx, runner.logger, ctx, parts.files[1])
	if err != nil {
		return err
	}

	objPath := ctx.UUID + "/" + ctx.TimePoint.UTC().Format("2006-01-02_15-04-05")
	if err := runner.uploadVideoToStore(bucketName, objPath, outFileName); err != nil {
		runner.logger.Error("failed to upload video to store", "err", err.Error())
		return err
	}
	if err := runner.uploadFramesToStore(bucketName, objPath, framePaths); err != nil {
		runner.logger.Error("failed to upload frames to store", "err", err.Error())
		return err
	}

	toRm := []string{}
	toRm = append(toRm, framePaths...)
	if err := runner.deletePaths(append(toRm, outFileName)); err != nil {
		return err
	}

	bytes, err := json.Marshal(v1.AnalyzeImgsEvent{
		UUID:  ctx.UUID,
		Paths: framePaths, // TODO(tomer): Change it to be the ffmpeg extracted frames
	})
	if err != nil {
		return err
	}

	return runner.bus.Publish(runner.ctx, "", "motion.analyze", bytes, nil)
}

func (runner *Runner) deletePaths(paths []string) error {
	var eg errgroup.Group

	for _, path := range paths {
		p := path // capture for lambda

		eg.Go(func() error {
			return os.Remove(p)
		})
	}

	return eg.Wait()
}

func (runner *Runner) uploadVideoToStore(bucketName, objPath, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	fileStat, err := file.Stat()
	if err != nil {
		return err
	}

	_, err = runner.minioClient.PutObject(bucketName, objPath+"/"+file.Name(), file, fileStat.Size(), time.Now().Add(time.Minute))
	if err != nil {
		return err
	}

	return nil
}

func (runner *Runner) uploadFramesToStore(bucketName, objPath string, paths []string) error {
	var eg errgroup.Group

	for _, path := range paths {
		p := path // capture for lambda

		eg.Go(func() error {
			file, err := os.Open(p)
			if err != nil {
				return err
			}
			defer file.Close()

			fileStat, err := file.Stat()
			if err != nil {
				return err
			}

			_, err = runner.minioClient.PutObject(bucketName, objPath+"/"+file.Name(), file, fileStat.Size(), time.Now().Add(time.Minute))
			if err != nil {
				return err
			}

			return nil
		})
	}

	return eg.Wait()
}

// extractFrames runs ffmpeg to extract up to 4 PNG frames at 10fps from
// the given motionVideoPath. The output files are named with the motion
// event UUID and timestamp, e.g. "motion_frame_<uuid>_<time>_0001.png".
//
// It returns the full list of expected output paths. If ffmpeg fails, an
// error is returned and details are logged with the provided logger.
//
// ctx is used to cancel the ffmpeg process early.
func extractFrames(ctx context.Context, logger *slog.Logger, motionCtx MotionCtx, motionVideoPath string) ([]string, error) {
	outFileName := fmt.Sprintf("motion_frame_%s_%s_%%04d.png", motionCtx.UUID, motionCtx.TimePoint.Format("2006-01-02_15-04-05"))
	args := []string{
		"-i", motionVideoPath,
		"-r", "1",
		"-vframes", "4",
		outFileName,
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		logger.Error(
			"ffmpeg failed",
			"motion_uuid", motionCtx.UUID,
			"timepoint", motionCtx.TimePoint,
			"stderr", stderr.String(),
			"error", err.Error(),
		)
		return nil, fmt.Errorf("ffmpeg failed: %w", err)
	}

	paths := make([]string, 0, 4)
	for i := 1; i <= cap(paths); i++ {
		paths = append(paths, fmt.Sprintf("motion_frame_%s_%s_%04d.png", motionCtx.UUID, motionCtx.TimePoint.Format("2006-01-02_15-04-05"), i))
	}

	return paths, nil
}

func getFileNames(entires []os.DirEntry) []string {
	lst := make([]string, 0)
	for _, entry := range entires {
		if !entry.IsDir() {
			lst = append(lst, entry.Name())
		}
	}

	return lst
}

func recordingDir(camID string) string {
	return filepath.Join("../recordings", camID)
}

func fileKeyFor(t time.Time) string {
	base := t.UTC().Format("2006-01-02_15-04-05")
	usec := t.Nanosecond() / 1_000
	return fmt.Sprintf("%s-%06d", base, usec)
}

func getFileList(camID string, motionTime time.Time) ([]string, error) {
	dir := recordingDir(camID)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	names := getFileNames(entries)
	key := fileKeyFor(motionTime)
	idx, found := slices.BinarySearch(names, key)

	out := make([]string, 0, 3)
	if found {
		if idx-1 >= 0 {
			out = append(out, filepath.Join(dir, names[idx-1]))
		}
		out = append(out, filepath.Join(dir, names[idx]))
		if idx+1 < len(names) {
			out = append(out, filepath.Join(dir, names[idx+1]))
		}
		return out, nil
	}

	if idx-1 >= 0 {
		out = append(out, filepath.Join(dir, names[idx-1]))
	}
	if idx < len(names) {
		out = append(out, filepath.Join(dir, names[idx]))
	}
	if idx+1 < len(names) {
		out = append(out, filepath.Join(dir, names[idx+1]))
	}

	return out, nil
}

func waitForFiles(path string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		file, err := os.Stat(path)
		if err == nil {
			sz := file.Size()
			if sz > 0 {
				time.Sleep(200 * time.Millisecond)
				file, err := os.Stat(path)
				if err == nil && file.Size() == sz {
					return nil
				}
			}
		}

		time.Sleep(200 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for %s", path)
}

func onMotion(uuid string, score int, motionTime time.Time) (*VideoArtifactParts, error) {
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return nil, fmt.Errorf("timeout waiting for video files")
		case <-ticker.C:
			fileList, err := getFileList(uuid, motionTime)
			if err != nil {
				return nil, err
			}
			if len(fileList) > 2 {
				if err := waitForFiles(fileList[len(fileList)-1], 5*time.Second); err != nil {
					return nil, err
				}
				return &VideoArtifactParts{score: score, files: fileList}, nil
			}
		}
	}
}

func concatVideoFiles(ctx context.Context, logger *slog.Logger, outputFileName string, fileList []string) error {
	listFileName := fmt.Sprintf("tmp-%s", outputFileName)
	if err := writeFileList(listFileName, reformatFileList(fileList)); err != nil {
		return err
	}
	defer os.Remove(listFileName)

	cmdName := "ffmpeg"
	cmdArgs := []string{
		"-f", "concat",
		"-safe", "0",
		"-i", listFileName,
		"-c", "copy", outputFileName,
	}

	cmd := exec.CommandContext(ctx, cmdName, cmdArgs...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		logger.Error(
			"ffmpeg failed",
			"fileList", fileList,
			"stderr", stderr.String(),
			"error", err.Error(),
		)
		return fmt.Errorf("ffmpeg failed: %w", err)
	}

	return nil
}

func reformatFileList(fileList []string) string {
	listContent := ""
	for _, fileName := range fileList {
		listContent += fmt.Sprintf("file '%s'\n", fileName)
	}

	return listContent
}

func writeFileList(dstName string, content string) error {
	return os.WriteFile(dstName, []byte(content), 0644)
}
