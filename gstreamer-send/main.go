package main

import (
	"flag"
	"fmt"
	"math/rand"
	"strconv"

	gst "github.com/pion/example-webrtc-applications/internal/gstreamer-src"
	"github.com/pion/example-webrtc-applications/internal/signal"
	"github.com/pion/sdp/v2"
	"github.com/pion/webrtc/v2"
)

func main() {
	audioSrc := flag.String("audio-src", "audiotestsrc", "GStreamer audio src")
	videoSrc := flag.String("video-src", "videotestsrc", "GStreamer video src")
	flag.Parse()

	// Everything below is the pion-WebRTC API! Thanks for using it ❤️.

	// Wait for the offer to be pasted
	offer := webrtc.SessionDescription{}
	signal.Decode(signal.MustReadStdin(), &offer)

	// We make our own mediaEngine and place the sender's codecs in it so that we use the
	// dynamic media type from the sender in our answer.
	mediaEngine := webrtc.MediaEngine{}
	err := mediaEngine.PopulateFromSDP(offer)
	if err != nil {
		panic(err)
	}

	api := webrtc.NewAPI(webrtc.WithMediaEngine(mediaEngine))

	// Prepare the configuration
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	// Create a new RTCPeerConnection
	peerConnection, err := api.NewPeerConnection(config)
	if err != nil {
		panic(err)
	}

	// Set the remote SessionDescription
	err = peerConnection.SetRemoteDescription(offer)
	if err != nil {
		panic(err)
	}

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("Connection State has changed %s \n", connectionState.String())
	})

	// Get the codecs we will be making tracks for, so we can use their dynamic payload types.
	// Currently doing this by parsing the SDP, but maybe we could get these from the MediaEngine at some point?
	opusCodec, err := firstCodecOfType(offer, webrtc.Opus)
	if err != nil {
		panic(err)
	}
	vp8Codec, err := firstCodecOfType(offer, webrtc.VP8)
	if err != nil {
		panic(err)
	}
	// Create a audio track
	audioTrack, err := peerConnection.NewTrack(opusCodec.PayloadType, rand.Uint32(), "audio", "pion1")
	if err != nil {
		panic(err)
	}
	_, err = peerConnection.AddTrack(audioTrack)
	if err != nil {
		panic(err)
	}

	// Create a video track
	firstVideoTrack, err := peerConnection.NewTrack(vp8Codec.PayloadType, rand.Uint32(), "video", "pion2")
	if err != nil {
		panic(err)
	}
	_, err = peerConnection.AddTrack(firstVideoTrack)
	if err != nil {
		panic(err)
	}

	// Create a second video track
	secondVideoTrack, err := peerConnection.NewTrack(vp8Codec.PayloadType, rand.Uint32(), "video", "pion3")
	if err != nil {
		panic(err)
	}
	_, err = peerConnection.AddTrack(secondVideoTrack)
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

	// Start pushing buffers on these tracks
	gst.CreatePipeline(webrtc.Opus, []*webrtc.Track{audioTrack}, *audioSrc).Start()
	gst.CreatePipeline(webrtc.VP8, []*webrtc.Track{firstVideoTrack, secondVideoTrack}, *videoSrc).Start()

	// Block forever
	select {}
}

// firstCodecOfType returns the first codec of a chosen type from a session description
func firstCodecOfType(sd webrtc.SessionDescription, codecName string) (*sdp.Codec, error) {
	sdpsd := sdp.SessionDescription{}
	err := sdpsd.Unmarshal([]byte(sd.SDP))
	if err != nil {
		return nil, err
	}
	for _, md := range sdpsd.MediaDescriptions {
		for _, format := range md.MediaName.Formats {
			pt, err := strconv.Atoi(format)
			if err != nil {
				return nil, fmt.Errorf("format parse error")
			}
			payloadType := uint8(pt)
			payloadCodec, err := sdpsd.GetCodecForPayloadType(payloadType)
			if err != nil {
				return nil, fmt.Errorf("could not find codec for payload type %d", payloadType)
			}
			if payloadCodec.Name == codecName {
				return &payloadCodec, nil
			}
		}
	}
	return nil, fmt.Errorf("no codec of type %s found in SDP", codecName)
}
