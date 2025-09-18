package main

import (
	"flag"
	"strings"
	"time"

	"gocv.io/x/gocv"
	"tomerab.com/cam-hub/internal/motion"
)

var (
	lastMotionEvent time.Time
	motionCoolDown  = 10 * time.Second
)

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

	runner := motion.NewRunner(8)

	for range ticker.C {
		if ok := cap.Read(&frame); !ok || frame.Empty() {
			continue
		}

		isMotion, score := det.Detect(&frame)
		if isMotion && time.Since(lastMotionEvent) > motionCoolDown {
			lastMotionEvent = time.Now()
			go runner.PostJob(motion.MotionCtx{
				UUID:      cameraUUID,
				Score:     score,
				TimePoint: lastMotionEvent,
			})
		}
	}
}
