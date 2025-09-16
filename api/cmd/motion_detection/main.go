package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"gocv.io/x/gocv"
	"tomerab.com/cam-hub/internal/motion"
)

var (
	lastMotionEvent time.Time
	motionCoolDown  = 10 * time.Second
)

type MotionVideoEvent struct {
	score int
	files []string
}

func main() {
	addr := flag.String("addr", "", "rtsp url")
	flag.Parse()
	if *addr == "" {
		panic("missing -addr")
	}

	parts := strings.Split(*addr, "/")
	cameraUUID := parts[len(parts)-1]

	cap, err := gocv.VideoCaptureFile(*addr)
	if err != nil {
		panic(err)
	}
	defer cap.Close()

	cap.Set(gocv.VideoCaptureBufferSize, 1)

	frame := gocv.NewMat()
	defer frame.Close()

	det := motion.NewMotionDetector(50, 15000, 150)
	defer det.Close()

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	motionChan := make(chan MotionVideoEvent, 8)
	recordingActive := false

	for {
		select {
		case <-ticker.C:
			if ok := cap.Read(&frame); !ok || frame.Empty() {
				continue
			}
			isMotion, score := det.Detect(&frame)
			if isMotion && time.Since(lastMotionEvent) > motionCoolDown && !recordingActive {
				lastMotionEvent = time.Now()
				recordingActive = true
				go func(motionTime time.Time, sc int) {
					onMotion(motionChan, cameraUUID, sc, motionTime)
				}(lastMotionEvent, score)
			}

		case ev := <-motionChan:
			fmt.Println(ev)
			recordingActive = false
		}
	}
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

	if found {
		out := make([]string, 0, 3)
		if idx-1 >= 0 {
			out = append(out, filepath.Join(dir, names[idx-1]))
		}
		out = append(out, filepath.Join(dir, names[idx]))
		if idx+1 < len(names) {
			out = append(out, filepath.Join(dir, names[idx+1]))
		}
		return out, nil
	}

	var out []string
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

func onMotion(motionChan chan<- MotionVideoEvent, uuid string, score int, motionTime time.Time) {
	for {
		fileList, err := getFileList(uuid, motionTime)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		if len(fileList) > 2 {
			if err := waitForFiles(fileList[len(fileList)-1], 2*time.Second); err != nil {
				fmt.Println(err.Error())
				return
			}
			motionChan <- MotionVideoEvent{
				score: score,
				files: fileList,
			}
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
}
