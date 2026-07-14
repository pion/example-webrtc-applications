// SPDX-FileCopyrightText: 2026 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

//go:build !js

// gstreamer-send-offer is a simple application that shows how to send video using Pion WebRTC and GStreamer
package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/go-gst/go-gst/pkg/gst"
	"github.com/go-gst/go-gst/pkg/gstapp"
	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"
)

func main() {
	audioSrc := flag.String("audio-src", "audiotestsrc", "GStreamer audio src")
	videoSrc := flag.String("video-src", "videotestsrc", "GStreamer video src")
	port := flag.Int("port", 8080, "http server port")
	flag.Parse()

	sdpChan := httpSDPServer(*port)

	// Initialize GStreamer
	gst.Init()

	// Everything below is the Pion WebRTC API! Thanks for using it ❤️.

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

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("Connection State has changed %s \n", connectionState.String())
	})

	// Create a audio track
	opusTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: "audio/opus"}, "audio", "pion1")
	if err != nil {
		panic(err)
	} else if _, err = peerConnection.AddTrack(opusTrack); err != nil {
		panic(err)
	}

	// Create a video track
	vp8Track, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: "video/vp8"}, "video", "pion2")
	if err != nil {
		panic(err)
	} else if _, err = peerConnection.AddTrack(vp8Track); err != nil {
		panic(err)
	}

	// Create an offer to send to the browser
	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		panic(err)
	}

	// Create channel that is blocked until ICE Gathering is complete
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	// Sets the LocalDescription, and starts our UDP listeners
	if err = peerConnection.SetLocalDescription(offer); err != nil {
		panic(err)
	}

	// Block until ICE Gathering is complete, disabling trickle ICE
	// we do this because we only can exchange one signaling message
	// in a production application you should exchange ICE Candidates via OnICECandidate
	<-gatherComplete

	// Output the offer in base64 so we can paste it in browser
	fmt.Println(encode(peerConnection.LocalDescription()))

	// Wait for the answer to be submitted via HTTP
	answer := webrtc.SessionDescription{}
	decode(<-sdpChan, &answer)

	// Set the remote SessionDescription
	err = peerConnection.SetRemoteDescription(answer)
	if err != nil {
		panic(err)
	}

	// Start pushing buffers on these tracks
	pipelineForCodec("opus", []*webrtc.TrackLocalStaticSample{opusTrack}, *audioSrc)
	pipelineForCodec("vp8", []*webrtc.TrackLocalStaticSample{vp8Track}, *videoSrc)

	// Block forever
	select {}
}

// Create the appropriate GStreamer pipeline depending on what codec we are working with.
func pipelineForCodec(codecName string, tracks []*webrtc.TrackLocalStaticSample, pipelineSrc string) { // nolint
	pipelineStr := "appsink name=appsink"
	switch codecName {
	case "vp8":
		pipelineStr = pipelineSrc + " ! vp8enc error-resilient=partitions keyframe-max-dist=10 auto-alt-ref=true cpu-used=5 deadline=1 ! " + pipelineStr // nolint
	case "vp9":
		pipelineStr = pipelineSrc + " ! vp9enc ! " + pipelineStr
	case "h264":
		pipelineStr = pipelineSrc + " ! video/x-raw,format=I420 ! x264enc speed-preset=ultrafast tune=zerolatency key-int-max=20 ! video/x-h264,stream-format=byte-stream ! " + pipelineStr // nolint
	case "opus":
		pipelineStr = pipelineSrc + " ! opusenc ! " + pipelineStr
	case "pcmu":
		pipelineStr = pipelineSrc + " ! audio/x-raw, rate=8000 ! mulawenc ! " + pipelineStr
	case "pcma":
		pipelineStr = pipelineSrc + " ! audio/x-raw, rate=8000 ! alawenc ! " + pipelineStr
	default:
		panic("Unhandled codec " + codecName) //nolint
	}

	element, err := gst.ParseLaunch(pipelineStr)
	if err != nil {
		panic(err)
	}
	pipeline, ok := element.(gst.Pipeline)
	if !ok {
		panic("GStreamer launch description did not produce a pipeline")
	}

	appSink, ok := pipeline.GetByName("appsink").(gstapp.AppSink)
	if !ok {
		panic("failed to find GStreamer appsink")
	}

	appSink.SetEmitSignals(true)

	appSink.ConnectNewSample(func(sink gstapp.AppSink) gst.FlowReturn {
		sample := sink.PullSample()
		if sample == nil {
			return gst.FlowEOS
		}

		buffer := sample.GetBuffer()
		if buffer == nil {
			return gst.FlowError
		}

		mapped, ok := buffer.Map(gst.MapRead)
		if !ok {
			return gst.FlowError
		}
		defer mapped.Unmap()
		samples := mapped.Data()

		var duration time.Duration
		if d := buffer.Duration(); d <= math.MaxInt64 {
			duration = time.Duration(d)
		}

		for _, t := range tracks {
			if err := t.WriteSample(media.Sample{Data: samples, Duration: duration}); err != nil {
				panic(err) //nolint
			}
		}

		return gst.FlowOK
	})

	if pipeline.SetState(gst.StatePlaying) == gst.StateChangeFailure {
		panic("failed to start GStreamer pipeline")
	}
}

// JSON encode + base64 a SessionDescription.
func encode(obj *webrtc.SessionDescription) string {
	b, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}

	return base64.StdEncoding.EncodeToString(b)
}

// Decode a base64 and unmarshal JSON into a SessionDescription.
func decode(in string, obj *webrtc.SessionDescription) {
	b, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		panic(err)
	}

	if err = json.Unmarshal(b, obj); err != nil {
		panic(err)
	}
}

// httpSDPServer starts a HTTP Server that consumes SDPs.
func httpSDPServer(port int) chan string {
	sdpChan := make(chan string)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		fmt.Fprintf(w, "done") //nolint: errcheck
		sdpChan <- string(body)
	})

	go func() {
		// nolint: gosec
		panic(http.ListenAndServe(":"+strconv.Itoa(port), nil))
	}()

	return sdpChan
}
