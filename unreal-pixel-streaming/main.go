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
}

func main() {
	url := flag.String("url", "", "URL to UE4 Pixel Streaming WebSocket endpoint")
	origin := flag.String("origin", "", "Origin that is passed in HTTP header")
	flag.Parse()

	if *url == "" || *origin == "" {
		panic("both url and origin are required arguments")
	}

	conn, err := websocket.Dial(*url, "", *origin)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err = conn.Close(); err != nil {
			panic(err)
		}
	}()

	peerConnection := &webrtc.PeerConnection{}
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
			peerConnection = createPeerConnection(conn, jsonMessage.PeerConnectionOptions)

			offer, offerErr := peerConnection.CreateOffer(nil)
			if offerErr != nil {
				panic(offerErr)
			}

			if err = peerConnection.SetLocalDescription(offer); err != nil {
				panic(err)
			}

			if err = websocket.JSON.Send(conn, offer); err != nil {
				panic(err)
			}
		case "answer":
			if err = peerConnection.SetRemoteDescription(webrtc.SessionDescription{Type: webrtc.SDPTypeAnswer, SDP: jsonMessage.SDP}); err != nil {
				panic(err)
			}
		case "iceCandidate":
			if err = peerConnection.AddICECandidate(jsonMessage.Candidate); err != nil {
				panic(err)
			}
		case "playerCount":
			fmt.Println("Player Count", jsonMessage.Count)
		default:
			fmt.Println("Unhandled type", jsonMessage.Type)
		}
	}
}

// Given a Configuration create a PeerConnection and set the appropriate handlers.
//nolint
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
