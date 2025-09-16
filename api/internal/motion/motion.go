// Implementation adapted from GoCV motion-detect example and other public resources.
// Wrapped into a reusable Go component for integration in a camera hub project.
// Original algorithm: OpenCV BackgroundSubtractorMOG2 + threshold + morphology.

package motion

import (
	"image"

	"gocv.io/x/gocv"
)

type MotionDetector struct {
	Threshold     float32 // binary threshold after MOG2 (e.g., 50..80)
	MinAreaPixels int     // minimum foreground pixels to call "motion"
	WarmupFrames  int     // frames to ignore while background stabilizes

	// Internals (reused buffers)
	mog2   gocv.BackgroundSubtractorMOG2
	gray   gocv.Mat
	mask   gocv.Mat
	kErode gocv.Mat
	kDil   gocv.Mat
	kClose gocv.Mat

	frameCount int
}

func NewMotionDetector(threshold float32, minAreaPixels, warmupFrames int) *MotionDetector {
	return &MotionDetector{
		Threshold:     threshold,
		MinAreaPixels: minAreaPixels,
		WarmupFrames:  warmupFrames,

		mog2:   gocv.NewBackgroundSubtractorMOG2WithParams(1024, 16, false),
		gray:   gocv.NewMat(),
		mask:   gocv.NewMat(),
		kErode: gocv.GetStructuringElement(gocv.MorphEllipse, image.Pt(2, 2)),
		kDil:   gocv.GetStructuringElement(gocv.MorphEllipse, image.Pt(5, 5)),
		kClose: gocv.GetStructuringElement(gocv.MorphEllipse, image.Pt(3, 3)),
	}
}

func (d *MotionDetector) Close() {
	d.mog2.Close()
	d.gray.Close()
	d.mask.Close()
	d.kErode.Close()
	d.kDil.Close()
	d.kClose.Close()
}

// Detect takes a BGR frame (Mat) and returns:
//
//	isMotion: true if motion pixels >= MinAreaPixels
//	score:    number of motion pixels in the cleaned mask
func (d *MotionDetector) Detect(frame *gocv.Mat) (isMotion bool, score int) {
	if frame == nil || frame.Empty() {
		return false, 0
	}

	d.frameCount++

	// 1) grayscale
	gocv.CvtColor(*frame, &d.gray, gocv.ColorBGRToGray)

	// 2) background subtraction
	d.mog2.Apply(d.gray, &d.mask)

	// 3) binarize
	gocv.Threshold(d.mask, &d.mask, d.Threshold, 255, gocv.ThresholdBinary)

	// 4) light cleanup
	gocv.Erode(d.mask, &d.mask, d.kErode)
	gocv.Dilate(d.mask, &d.mask, d.kDil)
	gocv.MorphologyEx(d.mask, &d.mask, gocv.MorphClose, d.kClose)

	// 5) score
	score = gocv.CountNonZero(d.mask)

	// Warmup: let background settle
	if d.frameCount <= d.WarmupFrames {
		return false, 0
	}

	return score >= d.MinAreaPixels, score
}

// Simple scoring function that checks how many pixels were masked to be non-zero
// out of the total pixels.
func (d *MotionDetector) ScoreRatio(frameWidth, frameHeight int, score int) float64 {
	total := frameWidth * frameHeight
	if total <= 0 {
		return 0
	}
	return float64(score) / float64(total)
}
