package main

import (
	"flag"
	"fmt"

	"github.com/pion/example-webrtc-applications/v3/sip-over-websocket-to-webrtc/softphone"
	"github.com/pion/sdp/v2"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media/oggwriter"
)

var (
	username  = flag.String("username", "1000", "Extension you wish to register as")
	password  = flag.String("password", "", "Password for the extension you wish to register as")
	extension = flag.String("extension", "9198", "Extension you wish to call")
	host      = flag.String("host", "", "Host that websocket is available on")
	port      = flag.String("port", "5066", "Port that websocket is available on")
)

func main() {
	flag.Parse()

	if *host == "" || *port == "" || *password == "" {
		panic("-host -port and -password are required")
	}

	conn := softphone.NewSoftPhone(softphone.SIPInfoResponse{
		Username:        *username,
		AuthorizationID: *username,
		Password:        *password,
		Domain:          *host,
		Transport:       "ws",
		OutboundProxy:   *host + ":" + *port,
	})

	pc, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		panic(err)
	}

	pc.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("Connection State has changed %s \n", connectionState.String())
	})

	oggFile, err := oggwriter.New("output.ogg", 48000, 2)
	if err != nil {
		panic(err)
	}

	pc.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		fmt.Println("Got Opus track, saving to disk as output.ogg")

		for {
			rtpPacket, _, readErr := track.ReadRTP()
			if readErr != nil {
				panic(readErr)
			}
			if readErr := oggFile.WriteRTP(rtpPacket); readErr != nil {
				panic(readErr)
			}
		}
	})

	if _, err = pc.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio); err != nil {
		panic(err)
	}

	offer, err := pc.CreateOffer(nil)
	if err != nil {
		panic(err)
	}

	if err := pc.SetLocalDescription(offer); err != nil {
		panic(err)
	}

	gotAnswer := false

	conn.OnOK(func(okBody string) {
		if gotAnswer {
			return
		}
		gotAnswer = true

		okBody += "a=mid:0\r\n"
		if err := pc.SetRemoteDescription(webrtc.SessionDescription{Type: webrtc.SDPTypeAnswer, SDP: okBody}); err != nil {
			panic(err)
		}
	})
	conn.Invite(*extension, rewriteSDP(offer.SDP))

	select {}
}

// Apply the following transformations for FreeSWITCH
// * Add fake srflx candidate to each media section
// * Add msid to each media section
// * Make bundle first attribute at session level.
func rewriteSDP(in string) string {
	parsed := &sdp.SessionDescription{}
	if err := parsed.Unmarshal([]byte(in)); err != nil {
		panic(err)
	}

	// Reverse global attributes
	for i, j := 0, len(parsed.Attributes)-1; i < j; i, j = i+1, j-1 {
		parsed.Attributes[i], parsed.Attributes[j] = parsed.Attributes[j], parsed.Attributes[i]
	}

	parsed.MediaDescriptions[0].Attributes = append(parsed.MediaDescriptions[0].Attributes, sdp.Attribute{
		Key:   "candidate",
		Value: "79019993 1 udp 1686052607 1.1.1.1 9 typ srflx",
	})

	out, err := parsed.Marshal()
	if err != nil {
		panic(err)
	}

	return string(out)
}
