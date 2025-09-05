// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

//go:build gocv
// +build gocv

package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
	"log"
	"net/http"
	"os/exec"
	"time"

	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"
	"github.com/pion/webrtc/v4/pkg/media/ivfreader"
	gocv "gocv.io/x/gocv"
)

const (
	videoWidth  = 640
	videoHeight = 480
	MinimumArea = 3000
)

var (
	status              = "Ready"
	imgDelta, imgThresh gocv.Mat
	mog2                gocv.BackgroundSubtractorMOG2

	//go:embed index.html
	indexHTML string
)

func main() {
	imgDelta = gocv.NewMat()
	imgThresh = gocv.NewMat()
	mog2 = gocv.NewBackgroundSubtractorMOG2()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, indexHTML)
	})

	// POST /offer will handle the browser's WebRTC offer
	http.HandleFunc("/offer", handleOffer)

	fmt.Println("Listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleOffer(w http.ResponseWriter, r *http.Request) {
	// Read the Offer from the browser
	var offer webrtc.SessionDescription
	if err := json.NewDecoder(r.Body).Decode(&offer); err != nil {
		http.Error(w, "invalid offer", http.StatusBadRequest)
		return
	}

	// Create a new PeerConnection
	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		http.Error(w, "failed to create PeerConnection", http.StatusInternalServerError)
		return
	}

	// Create a video track for VP8
	videoTrack, err := webrtc.NewTrackLocalStaticSample(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8},
		"video",
		"gocv",
	)
	if err != nil {
		http.Error(w, "failed to create video track", http.StatusInternalServerError)
		return
	}

	// Add the track to the PeerConnection
	rtpSender, err := peerConnection.AddTrack(videoTrack)
	if err != nil {
		http.Error(w, "failed to add track", http.StatusInternalServerError)
		return
	}

	// Read RTCP (for NACK, etc.) in a separate goroutine
	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
				return
			}
		}
	}()

	// Watch for ICE connection state
	iceConnectedCtx, iceConnectedCancel := context.WithCancel(context.Background())
	peerConnection.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		log.Printf("ICE state: %s\n", state)
		if state == webrtc.ICEConnectionStateConnected {
			iceConnectedCancel()
		}
	})

	// Set the remote description (the browser's Offer)
	if err := peerConnection.SetRemoteDescription(offer); err != nil {
		http.Error(w, "failed to set remote desc", http.StatusInternalServerError)
		return
	}

	// Create an Answer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		http.Error(w, "failed to create answer", http.StatusInternalServerError)
		return
	}

	// Gather ICE candidates
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
	if err := peerConnection.SetLocalDescription(answer); err != nil {
		http.Error(w, "failed to set local desc", http.StatusInternalServerError)
		return
	}
	<-gatherComplete

	// Write the Answer back to the browser
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(peerConnection.LocalDescription())

	// Once ICE is connected, start reading frames from the camera via GoCV,
	// pipe them into FFmpeg for VP8 encoding, and push the IVF frames into the track.
	go func() {
		<-iceConnectedCtx.Done()

		if err := startCameraAndStream(videoTrack); err != nil {
			log.Printf("camera streaming error: %v\n", err)
		}
	}()
}

// This was taken from the GoCV examples
// https://github.com/hybridgroup/gocv/blob/master/cmd/motion-detect/main.go
func detectMotion(img *gocv.Mat) {
	status = "Ready"
	statusColor := color.RGBA{0, 255, 0, 0}

	// first phase of cleaning up image, obtain foreground only
	mog2.Apply(*img, &imgDelta)

	// remaining cleanup of the image to use for finding contours.
	// first use threshold
	gocv.Threshold(imgDelta, &imgThresh, 25, 255, gocv.ThresholdBinary)

	// then dilate
	kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(3, 3))
	gocv.Dilate(imgThresh, &imgThresh, kernel)
	kernel.Close()

	// now find contours
	contours := gocv.FindContours(imgThresh, gocv.RetrievalExternal, gocv.ChainApproxSimple)

	for i := 0; i < contours.Size(); i++ {
		area := gocv.ContourArea(contours.At(i))
		if area < MinimumArea {
			continue
		}

		status = "Motion detected"
		statusColor = color.RGBA{255, 0, 0, 0}
		gocv.DrawContours(img, contours, i, statusColor, 2)

		rect := gocv.BoundingRect(contours.At(i))
		gocv.Rectangle(img, rect, color.RGBA{0, 0, 255, 0}, 2)
	}

	contours.Close()

	gocv.PutText(img, status, image.Pt(10, 20), gocv.FontHersheyPlain, 1.2, statusColor, 2)
}

// startCameraAndStream opens the webcam with GoCV, sends raw frames to FFmpeg (via stdin),
// reads IVF from FFmpeg (via stdout), and writes them into the WebRTC video track.
func startCameraAndStream(videoTrack *webrtc.TrackLocalStaticSample) error {
	webcam, err := gocv.OpenVideoCapture(0)
	if err != nil {
		return fmt.Errorf("cannot open camera: %w", err)
	}
	defer webcam.Close()

	webcam.Set(gocv.VideoCaptureFrameWidth, videoWidth)
	webcam.Set(gocv.VideoCaptureFrameHeight, videoHeight)

	ffmpeg := exec.Command(
		"ffmpeg",
		"-y",
		"-f", "rawvideo",
		"-pixel_format", "bgr24",
		"-video_size", fmt.Sprintf("%dx%d", videoWidth, videoHeight),
		"-framerate", "30",
		"-i", "pipe:0",
		"-c:v", "libvpx",
		"-b:v", "1M",
		"-f", "ivf",
		"pipe:1",
	)

	stdin, err := ffmpeg.StdinPipe()
	if err != nil {
		return fmt.Errorf("ffmpeg stdin error: %w", err)
	}
	stdout, err := ffmpeg.StdoutPipe()
	if err != nil {
		return fmt.Errorf("ffmpeg stdout error: %w", err)
	}

	// Start FFmpeg
	if err := ffmpeg.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	// Goroutine to write raw frames to FFmpeg stdin
	go func() {
		defer stdin.Close()

		frame := gocv.NewMat()
		defer frame.Close()

		ticker := time.NewTicker(time.Millisecond * 33) // ~30fps
		defer ticker.Stop()

		for range ticker.C {
			if ok := webcam.Read(&frame); !ok {
				log.Println("cannot read frame from camera")
				continue
			}
			if frame.Empty() {
				continue
			}

			detectMotion(&frame)

			if _, err = stdin.Write(frame.ToBytes()); err != nil {
				log.Println("Failed to send frame to ffmpeg")
			}
		}
	}()

	// Read IVF from FFmpeg stdout; parse frames with ivfreader
	ivf, _, err := ivfreader.NewWith(stdout)
	if err != nil {
		return fmt.Errorf("ivfreader init error: %w", err)
	}

	// Loop reading IVF frames; push them to the video track
	for {
		frame, _, err := ivf.ParseNextFrame()
		if errors.Is(err, io.EOF) {
			log.Println("ffmpeg ended (EOF)")
			break
		} else if err != nil {
			return fmt.Errorf("ivf parse error: %w", err)
		}

		if err := videoTrack.WriteSample(media.Sample{
			Data:     frame,
			Duration: time.Second / 30,
		}); err != nil {
			return fmt.Errorf("write sample error: %w", err)
		}
	}

	if err := ffmpeg.Wait(); err != nil {
		return fmt.Errorf("ffmpeg wait error: %w", err)
	}

	return nil
}
