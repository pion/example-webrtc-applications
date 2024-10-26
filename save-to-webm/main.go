// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

//go:build !js
// +build !js

// save-to-webm is a simple application that shows how to receive audio and video using Pion and then save to WebM container.
package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/at-wat/ebml-go/webm"
	"github.com/pion/interceptor/pkg/jitterbuffer"
	"github.com/pion/rtcp"
	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media/samplebuilder"
)

const (
	naluTypeBitmask = 0b11111
	naluTypeSPS     = 7
)

func main() {
	saver := newWebmSaver()
	peerConnection := createWebRTCConn(saver)

	closed := make(chan os.Signal, 1)
	signal.Notify(closed, os.Interrupt)
	<-closed

	if err := peerConnection.Close(); err != nil {
		panic(err)
	}
	saver.Close()
}

type webmSaver struct {
	audioWriter, videoWriter       webm.BlockWriteCloser
	audioBuilder, vp8Builder       *samplebuilder.SampleBuilder
	audioTimestamp, videoTimestamp time.Duration

	h264JitterBuffer   *jitterbuffer.JitterBuffer
	lastVideoTimestamp uint32
}

func newWebmSaver() *webmSaver {
	return &webmSaver{
		audioBuilder:     samplebuilder.New(10, &codecs.OpusPacket{}, 48000),
		vp8Builder:       samplebuilder.New(10, &codecs.VP8Packet{}, 90000),
		h264JitterBuffer: jitterbuffer.New(),
	}
}

func (s *webmSaver) Close() {
	fmt.Printf("Finalizing webm...\n")
	if s.audioWriter != nil {
		if err := s.audioWriter.Close(); err != nil {
			panic(err)
		}
	}
	if s.videoWriter != nil {
		if err := s.videoWriter.Close(); err != nil {
			panic(err)
		}
	}
}

func (s *webmSaver) PushOpus(rtpPacket *rtp.Packet) {
	s.audioBuilder.Push(rtpPacket)

	for {
		sample := s.audioBuilder.Pop()
		if sample == nil {
			return
		}
		if s.audioWriter != nil {
			s.audioTimestamp += sample.Duration
			if _, err := s.audioWriter.Write(true, int64(s.audioTimestamp/time.Millisecond), sample.Data); err != nil {
				panic(err)
			}
		}
	}
}

func (s *webmSaver) PushH264(rtpPacket *rtp.Packet) {
	s.h264JitterBuffer.Push(rtpPacket)

	pkt, err := s.h264JitterBuffer.Peek(true)
	if err != nil {
		return
	}

	pkts := []*rtp.Packet{pkt}
	for {
		pkt, err = s.h264JitterBuffer.PeekAtSequence(pkts[len(pkts)-1].SequenceNumber + 1)
		if err != nil {
			return
		}

		// We have popped a whole frame, lets write it
		if pkts[0].Timestamp != pkt.Timestamp {
			break
		}

		pkts = append(pkts, pkt)
	}

	h264Packet := &codecs.H264Packet{}
	data := []byte{}
	for i := range pkts {
		if _, err = s.h264JitterBuffer.PopAtSequence(pkts[i].SequenceNumber); err != nil {
			panic(err)
		}

		out, err := h264Packet.Unmarshal(pkts[i].Payload)
		if err != nil {
			panic(err)
		}
		data = append(data, out...)
	}

	videoKeyframe := (data[4] & naluTypeBitmask) == naluTypeSPS
	if s.videoWriter == nil && videoKeyframe {
		if s.videoWriter == nil || s.audioWriter == nil {
			s.InitWriter(true, 1280, 720)
		}
	}

	samples := uint32(0)
	if s.lastVideoTimestamp != 0 {
		samples = pkts[0].Timestamp - s.lastVideoTimestamp
	}
	s.lastVideoTimestamp = pkts[0].Timestamp

	if s.videoWriter != nil {
		s.videoTimestamp += time.Duration(float64(samples) / float64(90000) * float64(time.Second))
		if _, err := s.videoWriter.Write(videoKeyframe, int64(s.videoTimestamp/time.Millisecond), data); err != nil {
			panic(err)
		}
	}
}

func (s *webmSaver) PushVP8(rtpPacket *rtp.Packet) {
	s.vp8Builder.Push(rtpPacket)

	for {
		sample := s.vp8Builder.Pop()
		if sample == nil {
			return
		}
		// Read VP8 header.
		videoKeyframe := (sample.Data[0]&0x1 == 0)
		if videoKeyframe {
			// Keyframe has frame information.
			raw := uint(sample.Data[6]) | uint(sample.Data[7])<<8 | uint(sample.Data[8])<<16 | uint(sample.Data[9])<<24
			width := int(raw & 0x3FFF)
			height := int((raw >> 16) & 0x3FFF)

			if s.videoWriter == nil || s.audioWriter == nil {
				s.InitWriter(false, width, height)
			}
		}
		if s.videoWriter != nil {
			s.videoTimestamp += sample.Duration
			if _, err := s.videoWriter.Write(videoKeyframe, int64(s.videoTimestamp/time.Millisecond), sample.Data); err != nil {
				panic(err)
			}
		}
	}
}

func (s *webmSaver) InitWriter(isH264 bool, width, height int) {
	w, err := os.OpenFile("test.webm", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		panic(err)
	}

	videoMimeType := "V_VP8"
	if isH264 {
		videoMimeType = "V_MPEG4/ISO/AVC"
	}

	ws, err := webm.NewSimpleBlockWriter(w,
		[]webm.TrackEntry{
			{
				Name:            "Audio",
				TrackNumber:     1,
				TrackUID:        12345,
				CodecID:         "A_OPUS",
				TrackType:       2,
				DefaultDuration: 20000000,
				Audio: &webm.Audio{
					SamplingFrequency: 48000.0,
					Channels:          2,
				},
			}, {
				Name:            "Video",
				TrackNumber:     2,
				TrackUID:        67890,
				CodecID:         videoMimeType,
				TrackType:       1,
				DefaultDuration: 33333333,
				Video: &webm.Video{
					PixelWidth:  uint64(width),
					PixelHeight: uint64(height),
				},
			},
		})
	if err != nil {
		panic(err)
	}
	fmt.Printf("WebM saver has started with video width=%d, height=%d\n", width, height)
	s.audioWriter = ws[0]
	s.videoWriter = ws[1]
}

func createWebRTCConn(saver *webmSaver) *webrtc.PeerConnection {
	// Everything below is the Pion WebRTC API! Thanks for using it ❤️.

	// Prepare the configuration
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	// Create a MediaEngine object to configure the supported codec
	m := &webrtc.MediaEngine{}

	// Setup the codecs you want to use.
	// This example supports VP8 or H264. Some browsers may only support one (or the other)
	if err := m.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8, ClockRate: 90000, Channels: 0, SDPFmtpLine: "", RTCPFeedback: nil},
		PayloadType:        96,
	}, webrtc.RTPCodecTypeVideo); err != nil {
		panic(err)
	}
	if err := m.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264, ClockRate: 90000, Channels: 0, SDPFmtpLine: "", RTCPFeedback: nil},
		PayloadType:        98,
	}, webrtc.RTPCodecTypeVideo); err != nil {
		panic(err)
	}
	if err := m.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus, ClockRate: 48000, Channels: 0, SDPFmtpLine: "", RTCPFeedback: nil},
		PayloadType:        111,
	}, webrtc.RTPCodecTypeAudio); err != nil {
		panic(err)
	}

	// Create the API object with the MediaEngine
	api := webrtc.NewAPI(webrtc.WithMediaEngine(m))

	// Create a new RTCPeerConnection
	peerConnection, err := api.NewPeerConnection(config)
	if err != nil {
		panic(err)
	}

	// Set a handler for when a new remote track starts, this handler copies inbound RTP packets,
	// replaces the SSRC and sends them back
	peerConnection.OnTrack(func(track *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		if track.Kind() == webrtc.RTPCodecTypeVideo {
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
		}

		fmt.Printf("Track has started, of type %d: %s \n", track.PayloadType(), track.Codec().RTPCodecCapability.MimeType)
		for {
			// Read RTP packets being sent to Pion
			rtp, _, readErr := track.ReadRTP()
			if readErr != nil {
				if errors.Is(readErr, io.EOF) {
					return
				}
				panic(readErr)
			}

			switch track.Codec().MimeType {
			case webrtc.MimeTypeOpus:
				saver.PushOpus(rtp)
			case webrtc.MimeTypeVP8:
				saver.PushVP8(rtp)
			case webrtc.MimeTypeH264:
				saver.PushH264(rtp)
			}
		}
	})
	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("Connection State has changed %s \n", connectionState.String())
	})

	// Wait for the offer to be pasted
	offer := webrtc.SessionDescription{}
	decode(readUntilNewline(), &offer)

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

	// Create channel that is blocked until ICE Gathering is complete
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	// Sets the LocalDescription, and starts our UDP listeners
	err = peerConnection.SetLocalDescription(answer)
	if err != nil {
		panic(err)
	}

	// Block until ICE Gathering is complete, disabling trickle ICE
	// we do this because we only can exchange one signaling message
	// in a production application you should exchange ICE Candidates via OnICECandidate
	<-gatherComplete

	// Output the answer in base64 so we can paste it in browser
	fmt.Println(encode(peerConnection.LocalDescription()))

	return peerConnection
}

// Read from stdin until we get a newline
func readUntilNewline() (in string) {
	var err error

	r := bufio.NewReader(os.Stdin)
	for {
		in, err = r.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			panic(err)
		}

		if in = strings.TrimSpace(in); len(in) > 0 {
			break
		}
	}

	fmt.Println("")
	return
}

// JSON encode + base64 a SessionDescription
func encode(obj *webrtc.SessionDescription) string {
	b, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}

	return base64.StdEncoding.EncodeToString(b)
}

// Decode a base64 and unmarshal JSON into a SessionDescription
func decode(in string, obj *webrtc.SessionDescription) {
	b, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		panic(err)
	}

	if err = json.Unmarshal(b, obj); err != nil {
		panic(err)
	}
}
