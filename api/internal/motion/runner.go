package motion

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"time"
)

type MotionCtx struct {
	UUID      string
	Score     int
	TimePoint time.Time
}

type Runner struct {
	logger *slog.Logger
	sem    chan struct{}
}

func NewRunner(maxJobs int) *Runner {
	return &Runner{
		logger: slog.New(slog.NewTextHandler(os.Stdout, nil)),
		sem:    make(chan struct{}, maxJobs),
	}
}

func (runner *Runner) PostJob(ctx MotionCtx) error {
	runner.sem <- struct{}{}
	defer func() { <-runner.sem }()
	return process(runner.logger, ctx)
}

func process(logger *slog.Logger, ctx MotionCtx) error {
	parts, err := onMotion(ctx.UUID, ctx.Score, ctx.TimePoint)
	if err != nil {
		return err
	}

	outFileName := fmt.Sprintf("motion_%s_%s.mp4", ctx.UUID, ctx.TimePoint.Format("2006-01-02_15-04-05"))
	logger.Info("File part", "parts", parts.files)
	if err := concatVideoFiles(outFileName, parts.files); err != nil {
		return err
	}
	logger.Info("Created new file", "path", outFileName)

	// TODO(tomer): Write to minio

	// TODO(tomer): Write to rabbitmq

	return nil
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
				if err := waitForFiles(fileList[len(fileList)-1], 2*time.Second); err != nil {
					return nil, err
				}
				return &VideoArtifactParts{score: score, files: fileList}, nil
			}
		}
	}
}

func concatVideoFiles(outputFileName string, fileList []string) error {
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

	cmd := exec.Command(cmdName, cmdArgs...)
	//cmd.Stdout = os.Stdout
	//cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return err
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
