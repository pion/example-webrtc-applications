// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

//go:build !js
// +build !js

// unreal-pixel-streaming demonstrates how to connect to a Unreal Pixel Streaming instance and accept the inbound audio/video
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"time"

	"github.com/pion/webrtc/v3"
	"golang.org/x/net/websocket"
)

type websocketMessage struct {
	Type                  string                  `json:"type"`
	PeerConnectionOptions webrtc.Configuration    `json:"peerConnectionOptions"`
	Count                 int                     `json:"count"`
	SDP                   string                  `json:"sdp"`
	Candidate             webrtc.ICECandidateInit `json:"candidate"`
	IDs                   []string                `json:"ids"`
	StreamerID            string                  `json:"streamerId"`
}

func main() {
	url := flag.String("url", "ws://localhost/", "URL to UE5 Pixel Streaming WebSocket endpoint")
	origin := flag.String("origin", "http://localhost", "Origin that is passed in HTTP header")
	flag.Parse()

	conn, err := websocket.Dial(*url, "", *origin)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err = conn.Close(); err != nil {
			panic(err)
		}
	}()

	if err = websocket.JSON.Send(conn, websocketMessage{Type: "listStreamers"}); err != nil {
		panic(err)
	}

	peerConnection := &webrtc.PeerConnection{}
	peerConnectionConfig := webrtc.Configuration{}
	data := []byte{}
	jsonMessage := websocketMessage{}

	for {
		if err = websocket.Message.Receive(conn, &data); err != nil {
			panic(err)
		} else if err = json.Unmarshal(data, &jsonMessage); err != nil {
			panic(err)
		}

		switch jsonMessage.Type {
		case "config":
			peerConnectionConfig = jsonMessage.PeerConnectionOptions

		case "offer":
			peerConnection = createPeerConnection(conn, peerConnectionConfig)

			if err = peerConnection.SetRemoteDescription(webrtc.SessionDescription{Type: webrtc.SDPTypeOffer, SDP: jsonMessage.SDP}); err != nil {
				panic(err)
			}

			answer, answerErr := peerConnection.CreateAnswer(nil)
			if answerErr != nil {
				panic(answerErr)
			}

			if err = peerConnection.SetLocalDescription(answer); err != nil {
				panic(err)
			}

			if err = websocket.JSON.Send(conn, answer); err != nil {
				panic(err)
			}
		case "iceCandidate":
			if err = peerConnection.AddICECandidate(jsonMessage.Candidate); err != nil {
				panic(err)
			}
		case "playerCount":
			fmt.Println("Player Count", jsonMessage.Count)
		case "streamerList":
			if len(jsonMessage.IDs) >= 1 {
				if err = websocket.JSON.Send(conn, websocketMessage{Type: "subscribe", StreamerID: jsonMessage.IDs[0]}); err != nil {
					panic(err)
				}
			}
		default:
			fmt.Println("Unhandled type", jsonMessage.Type)
		}
	}
}

// Given a Configuration create a PeerConnection and set the appropriate handlers.
// nolint
func createPeerConnection(conn *websocket.Conn, configuration webrtc.Configuration) *webrtc.PeerConnection {
	m := &webrtc.MediaEngine{}

	if err := m.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264, ClockRate: 90000, Channels: 0, SDPFmtpLine: "level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42e034", RTCPFeedback: nil},
		PayloadType:        96,
	}, webrtc.RTPCodecTypeVideo); err != nil {
		panic(err)
	} else if err := m.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus, ClockRate: 48000, Channels: 2, SDPFmtpLine: "minptime=10;useinbandfec=1", RTCPFeedback: nil},
		PayloadType:        111,
	}, webrtc.RTPCodecTypeAudio); err != nil {
		panic(err)
	}

	peerConnection, err := webrtc.NewAPI(webrtc.WithMediaEngine(m)).NewPeerConnection(configuration)
	if err != nil {
		panic(err)
	}

	if _, err := peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio); err != nil {
		panic(err)
	} else if _, err := peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo); err != nil {
		panic(err)
	} else if _, err = peerConnection.CreateDataChannel("cirrus", nil); err != nil {
		panic(err)
	}

	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("Connection State has changed %s \n", connectionState.String())
	})

	peerConnection.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}
		if err = websocket.JSON.Send(conn, &websocketMessage{Type: "iceCandidate", Candidate: c.ToJSON()}); err != nil {
			panic(err)
		}
	})

	peerConnection.OnTrack(func(t *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		fmt.Printf("Track has started, of type %d: %s \n", t.PayloadType(), t.Codec().RTPCodecCapability.MimeType)
	})

	go func() {
		for range time.NewTicker(20 * time.Second).C {
			if err = websocket.JSON.Send(conn, &websocketMessage{Type: "keepalive"}); err != nil {
				panic(err)
			}
		}
	}()

	return peerConnection
}
