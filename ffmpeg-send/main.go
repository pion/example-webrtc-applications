package main

import (
	"fmt"
	"math/rand"

	"github.com/pion/webrtc/v2"
	"github.com/pion/webrtc/v2/pkg/media"

	"github.com/pion/example-webrtc-applications/internal/ffmpeg"
	"github.com/pion/example-webrtc-applications/internal/signal"
)

// gstreamerReceiveMain is launched in a goroutine because the main thread is needed
// for Glib's main loop (Gstreamer uses Glib)
func ffmpegSendMain() {
	// Everything below is the pion-WebRTC API! Thanks for using it ❤️.

	// Prepare the configuration
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	// Create a new RTCPeerConnection
	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		panic(err)
	}

	track, err := peerConnection.NewTrack(webrtc.DefaultPayloadTypeH264, rand.Uint32(), "video", "stream")
	if err != nil {
		panic(err)
	}

	_, err = peerConnection.AddTrack(track)
	if err != nil {
		panic(err)
	}

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("Connection State has changed %s \n", connectionState.String())
	})

	// Wait for the offer to be pasted
	offer := webrtc.SessionDescription{}
	signal.Decode(signal.MustReadStdin(), &offer)

	// Check that h264 codec is supported by offer sdp
	mediaEngine := webrtc.MediaEngine{}
	if err := mediaEngine.PopulateFromSDP(offer); err != nil {
		fmt.Println("webrtc could not create media engine.", err)
		panic(err)
	}

	var h264PayloadType uint8
	for _, videoCodec := range mediaEngine.GetCodecsByKind(webrtc.RTPCodecTypeVideo) {
		if videoCodec.Name == "H264" {
			h264PayloadType = videoCodec.PayloadType
			break
		}
	}

	if h264PayloadType == 0 {
		fmt.Println("Remote peer does not support H264")
		panic(err)
	}

	// Set the remote SessionDescription
	err = peerConnection.SetRemoteDescription(offer)
	if err != nil {
		panic(err)
	}

	// Create an answer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		panic(err)
	}

	// Sets the LocalDescription, and starts our UDP listeners
	err = peerConnection.SetLocalDescription(answer)
	if err != nil {
		panic(err)
	}

	// Output the answer in base64 so we can paste it in browser
	fmt.Println(signal.Encode(answer))

	go func() {
		stdOut := ffmpeg.CreateH264Pipe()
		buf := make([]byte, ffmpeg.FrameSize)
		for {
			n, err := stdOut.Read(buf)
			if err != nil {
				panic(err)
			}
			err = track.WriteSample(media.Sample{
				Data:    buf[:n],
				Samples: 90000, // TODO: correctly determine samples
			})
			if err != nil {
				panic(err)
			}
		}
	}()

	// Block forever
	select {}
}

func main() {
	// Start a new thread to do the actual work for this application
	ffmpegSendMain()
}
