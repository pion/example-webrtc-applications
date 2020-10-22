package main

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/signal"
	"time"

	"github.com/at-wat/ebml-go/webm"
	"github.com/pion/example-webrtc-applications/utils"

	webrtcsignal "github.com/pion/example-webrtc-applications/internal/signal"
	"github.com/pion/rtcp"
	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v2"
	"github.com/pion/webrtc/v2/pkg/media/samplebuilder"
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
	audioBuilder, videoBuilder     *samplebuilder.SampleBuilder
	audioTimestamp, videoTimestamp uint32
}

func newWebmSaver() *webmSaver {
	return &webmSaver{
		audioBuilder: samplebuilder.New(10, &codecs.OpusPacket{}),
		videoBuilder: samplebuilder.New(10, &codecs.VP8Packet{}),
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
			s.audioTimestamp += sample.Samples
			t := s.audioTimestamp / 48
			if _, err := s.audioWriter.Write(true, int64(t), sample.Data); err != nil {
				panic(err)
			}
		}
	}
}
func (s *webmSaver) PushVP8(rtpPacket *rtp.Packet) {
	s.videoBuilder.Push(rtpPacket)

	for {
		sample := s.videoBuilder.Pop()
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
				// Initialize WebM saver using received frame size.
				s.InitWriter(width, height)
			}
		}
		if s.videoWriter != nil {
			s.videoTimestamp += sample.Samples
			t := s.videoTimestamp / 90
			if _, err := s.videoWriter.Write(videoKeyframe, int64(t), sample.Data); err != nil {
				panic(err)
			}
		}
	}
}
func (s *webmSaver) InitWriter(width, height int) {
	w, err := os.OpenFile("test.webm", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		panic(err)
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
				CodecID:         "V_VP8",
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
	// Everything below is the pion-WebRTC API! Thanks for using it ❤️.

	// Prepare the configuration
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	// Create a MediaEngine object to configure the supported codec
	m := webrtc.MediaEngine{}

	// Setup the codecs you want to use.
	// Only support VP8 and OPUS, this makes our WebM muxer code simpler
	m.RegisterCodec(webrtc.NewRTPVP8Codec(webrtc.DefaultPayloadTypeVP8, 90000))
	m.RegisterCodec(webrtc.NewRTPOpusCodec(webrtc.DefaultPayloadTypeOpus, 48000))

	se := &webrtc.SettingEngine{}
	se.SetSRTPReplayProtectionWindow(1024)

	// Create the API object with the MediaEngine
	api := webrtc.NewAPI(webrtc.WithMediaEngine(m), webrtc.WithSettingEngine(*se))

	// Create a new RTCPeerConnection
	peerConnection, err := api.NewPeerConnection(config)
	if err != nil {
		panic(err)
	}

	if _, err = peerConnection.AddTransceiver(webrtc.RTPCodecTypeAudio); err != nil {
		panic(err)
	} else if _, err = peerConnection.AddTransceiver(webrtc.RTPCodecTypeVideo); err != nil {
		panic(err)
	}

	// Set a handler for when a new remote track starts, this handler copies inbound RTP packets,
	// replaces the SSRC and sends them back
	peerConnection.OnTrack(func(track *webrtc.Track, receiver *webrtc.RTPReceiver) {
		// Send a PLI on an interval so that the publisher is pushing a keyframe every rtcpPLIInterval
		// This is a temporary fix until we implement incoming RTCP events, then we would push a PLI only when a viewer requests it
		go func() {
			ticker := time.NewTicker(time.Second * 3)
			for range ticker.C {
				errSend := peerConnection.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: track.SSRC()}})
				if errSend != nil {
					fmt.Println(errSend)
				}
			}
		}()

		receiveLog, err := utils.NewReceiveLog(1024)
		if err != nil {
			panic(err)
		}
		var senderSSRC = rand.Uint32()
		go func() {
			ticker := time.NewTicker(time.Millisecond * 50)
			for range ticker.C {
				missing := receiveLog.MissingSeqNumbers(10)
				nack := &rtcp.TransportLayerNack{
					SenderSSRC: senderSSRC,
					MediaSSRC:  track.SSRC(),
					Nacks:      utils.NackPairs(missing),
				}
				errSend := peerConnection.WriteRTCP([]rtcp.Packet{nack})
				if errSend != nil {
					fmt.Println(errSend)
				}
			}
		}()

		sequenceUnwrapper := utils.NewSequenceUnwrapper(16)
		jitterBuffer := utils.NewJitterBuffer(512)

		fmt.Printf("Track has started, of type %d: %s \n", track.PayloadType(), track.Codec().Name)
		for {
			// Read RTP packets being sent to Pion
			rtp, readErr := track.ReadRTP()
			if readErr != nil {
				if readErr == io.EOF {
					return
				}
				panic(readErr)
			}

			receiveLog.Add(rtp.SequenceNumber)

			seq := sequenceUnwrapper.Unwrap(uint64(rtp.SequenceNumber))
			if !jitterBuffer.Add(seq, rtp) {
				continue
			}

			// jitterBuffer.SetNextPacketsStart() can be called in case of a keyframe, so the buffer content is dropped up until the keyframe

			for _, rtp := range jitterBuffer.NextPackets() {
				switch track.Kind() {
				case webrtc.RTPCodecTypeAudio:
					saver.PushOpus(rtp)
				case webrtc.RTPCodecTypeVideo:
					saver.PushVP8(rtp)
				}
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
	webrtcsignal.Decode(webrtcsignal.MustReadStdin(), &offer)

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
	fmt.Println(webrtcsignal.Encode(answer))

	peerConnection.GetReceivers()

	return peerConnection
}

func SendBufferExample(track *webrtc.Track, receiver *webrtc.RTPReceiver) {
	sendBuffer, err := utils.NewSendBuffer(1024)
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			time.Sleep(100 / time.Millisecond)
			// ... do stuff

			someRTP := &rtp.Packet{}
			sendBuffer.Add(someRTP)
			track.WriteRTP(someRTP) // write some packet
		}
	}()

	for {
		packets, err := receiver.ReadRTCP()
		if err != nil {
			panic(err)
		}

		for _, p := range packets {
			switch packet := p.(type) {
			case *rtcp.TransportLayerNack:
				missing := utils.NackParsToSequenceNumbers(packet.Nacks)
				for _, m := range missing {
					rtp := sendBuffer.Get(m)
					if rtp != nil {
						track.WriteRTP(rtp)
					}
				}
			}
		}
	}
}
