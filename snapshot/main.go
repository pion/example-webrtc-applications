package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/jpeg"
	"net/http"
	"strconv"
	"time"

	"github.com/pion/rtcp"
	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media/samplebuilder"
	"golang.org/x/image/vp8"
)

// Channel for PeerConnection to push RTP Packets
// This is the read from HTTP Handler for generating jpeg
var rtpChan chan *rtp.Packet // nolint:gochecknoglobals

func signaling(w http.ResponseWriter, r *http.Request) {
	// Create a new PeerConnection
	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	})
	if err != nil {
		panic(err)
	}

	// Set a handler for when a new remote track starts, this handler saves buffers to SampleBuilder
	// so we can generate a snapshot
	peerConnection.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		// Send a PLI on an interval so that the publisher is pushing a keyframe every rtcpPLIInterval
		go func() {
			ticker := time.NewTicker(time.Second * 3)
			for range ticker.C {
				errSend := peerConnection.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: uint32(track.SSRC())}})
				if errSend != nil {
					fmt.Println(errSend)
				}
			}
		}()

		for {
			// Read RTP Packets in a loop
			rtpPacket, _, readErr := track.ReadRTP()
			if readErr != nil {
				panic(readErr)
			}

			// Use a lossy channel to send packets to snapshot handler
			// We don't want to block and queue up old data
			select {
			case rtpChan <- rtpPacket:
			default:
			}
		}
	})

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("ICE Connection State has changed: %s\n", connectionState.String())
	})

	var offer webrtc.SessionDescription
	if err = json.NewDecoder(r.Body).Decode(&offer); err != nil {
		panic(err)
	}

	if err = peerConnection.SetRemoteDescription(offer); err != nil {
		panic(err)
	}

	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		panic(err)
	}

	// Create channel that is blocked until ICE Gathering is complete
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	if err = peerConnection.SetLocalDescription(answer); err != nil {
		panic(err)
	}

	// Block until ICE Gathering is complete, disabling trickle ICE
	// we do this because we only can exchange one signaling message
	// in a production application you should exchange ICE Candidates via OnICECandidate
	<-gatherComplete

	response, err := json.Marshal(*peerConnection.LocalDescription())
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(response); err != nil {
		panic(err)
	}
}

func snapshot(w http.ResponseWriter, r *http.Request) {
	// Initialized with 20 maxLate, my samples sometimes 10-15 packets
	sampleBuilder := samplebuilder.New(20, &codecs.VP8Packet{}, 90000)
	decoder := vp8.NewDecoder()

	for {
		// Pull RTP Packet from rtpChan
		sampleBuilder.Push(<-rtpChan)

		// Use SampleBuilder to generate full picture from many RTP Packets
		sample := sampleBuilder.Pop()
		if sample == nil {
			continue
		}

		// Read VP8 header.
		videoKeyframe := (sample.Data[0]&0x1 == 0)
		if !videoKeyframe {
			continue
		}

		// Begin VP8-to-image decode: Init->DecodeFrameHeader->DecodeFrame
		decoder.Init(bytes.NewReader(sample.Data), len(sample.Data))

		// Decode header
		if _, err := decoder.DecodeFrameHeader(); err != nil {
			panic(err)
		}

		// Decode Frame
		img, err := decoder.DecodeFrame()
		if err != nil {
			panic(err)
		}

		// Encode to (RGB) jpeg
		buffer := new(bytes.Buffer)
		if err = jpeg.Encode(buffer, img, nil); err != nil {
			panic(err)
		}

		// Serve image
		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))

		// Write jpeg as HTTP Response
		if _, err = w.Write(buffer.Bytes()); err != nil {
			panic(err)
		}
		return
	}
}

func main() {
	rtpChan = make(chan *rtp.Packet)

	http.Handle("/", http.FileServer(http.Dir(".")))
	http.HandleFunc("/signal", signaling)
	http.HandleFunc("/snapshot", snapshot)

	fmt.Println("Open http://localhost:8080 to access this demo")
	panic(http.ListenAndServe(":8080", nil))
}
