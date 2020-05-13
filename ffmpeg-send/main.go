package main

import (
	"fmt"
	"math/rand"
	"net"

	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v2"

	"github.com/pion/example-webrtc-applications/internal/ffmpeg"
	"github.com/pion/example-webrtc-applications/internal/signal"
)

func main() {
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

	s, err := net.ResolveUDPAddr("udp4", ":0")
	if err != nil {
		fmt.Println(err)
		return
	}

	connection, err := net.ListenUDP("udp4", s)
	if err != nil {
		fmt.Println(err)
		return
	}

	stdOut := ffmpeg.CreateH264Pipe("x11grab")
	buf := make([]byte, ffmpeg.FrameSize)
	packetizer := rtp.NewPacketizer(1400, 96, 5000, &codecs.H264Payloader{}, rtp.NewFixedSequencer(0), 90000)

	for {
		n, err := stdOut.Read(buf)
		if err != nil {
			panic(err)
		}

		for _, f := range packetizer.Packetize(buf[:n], 90000) {
			raw, _ := f.Marshal()

			connection.WriteTo(raw, &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 50001})
		}

		// panic("fug")
		// err = track.WriteSample(media.Sample{
		// 	Data:    buf[:n],
		// 	Samples: 90000, // TODO: correctly determine samples
		// })
		// if err != nil {
		// 	panic(err)
		// }
	}
}
