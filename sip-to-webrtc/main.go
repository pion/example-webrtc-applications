// SPDX-FileCopyrightText: 2026 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

//go:build !js
// +build !js

// sip demonstrates how to bridge SIP traffic and WebRTC
package main

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/emiago/sipgo"
	"github.com/emiago/sipgo/sip"
	"github.com/pion/sdp/v3"
	"github.com/pion/webrtc/v4"
)

// nolint
var (
	audioTrack     *webrtc.TrackLocalStaticRTP
	unicastAddress = flag.String("unicast-address", "", "IP of SIP Server (your public IP)")
	sipPort        = flag.Int("sip-port", 5060, "Port to listen for SIP Traffic")

	contentTypeHeaderSDP = sip.ContentTypeHeader("application/sdp")
)

func main() { // nolint
	// Parse the flags passed to program
	flag.Parse()

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		panic(err)
	}
	for _, address := range addrs {
		if *unicastAddress != "" {
			break
		}

		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				*unicastAddress = ipnet.IP.String()
			}
		}
	}

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
	audioTrack, err = webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypePCMU}, "audio", "pion") // nolint
	if err != nil {
		panic(err)
	}
	if _, err = peerConnection.AddTrack(audioTrack); err != nil {
		panic(err)
	}

	// Wait for the offer to be pasted
	offer := webrtc.SessionDescription{}
	decode(readUntilNewline(), &offer)

	// Set the remote SessionDescription
	if err = peerConnection.SetRemoteDescription(offer); err != nil {
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
	if err = peerConnection.SetLocalDescription(answer); err != nil {
		panic(err)
	}

	<-gatherComplete

	// Output the answer in base64 so we can paste it in browser
	fmt.Println(encode(peerConnection.LocalDescription()))

	// Create the SIP UA
	sipUserAgent, err := sipgo.NewUA()
	if err != nil {
		panic(err)
	}

	// Create the SIP Server
	sipServer, err := sipgo.NewServer(sipUserAgent)
	if err != nil {
		panic(err)
	}

	sipServer.OnInvite(func(req *sip.Request, tx sip.ServerTransaction) {
		rtpListenerPort := startRTPListener()

		res := sip.NewResponseFromRequest(req, 200, "OK", generateAnswer(req.Body(), *unicastAddress, rtpListenerPort))
		res.AppendHeader(&sip.ContactHeader{Address: sip.Uri{Host: *unicastAddress, Port: *sipPort}})
		res.AppendHeader(&contentTypeHeaderSDP)
		if err = tx.Respond(res); err != nil {
			panic(err)
		}

		fmt.Printf("Accepting SIP Invite: %s\n", req.From())
	})

	sipServer.OnBye(func(req *sip.Request, tx sip.ServerTransaction) {
		if err = tx.Respond(sip.NewResponseFromRequest(req, 200, "OK", nil)); err != nil {
			panic(err)
		}
	})

	sipServer.OnAck(func(req *sip.Request, tx sip.ServerTransaction) {
		if err = tx.Respond(sip.NewResponseFromRequest(req, 200, "OK", nil)); err != nil {
			panic(err)
		}
	})

	fmt.Println("Starting SIP Listener")

	// Start Listening for SIP Traffic
	panic(sipServer.ListenAndServe(context.TODO(), "udp", fmt.Sprintf("0.0.0.0:%d", *sipPort)))
}

func startRTPListener() int {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		Port: 0,
		IP:   net.ParseIP("0.0.0.0"),
	})
	if err != nil {
		panic(err)
	}

	go func() {
		buff := make([]byte, 1500)

		for {
			n, _, err := conn.ReadFromUDP(buff)
			if err != nil {
				panic(err)
			}

			if _, err := audioTrack.Write(buff[:n]); err != nil {
				panic(err)
			}
		}
	}()

	udpAddr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok {
		panic("Failed to cast *net.UDPAddr")
	}

	return udpAddr.Port
}

func generateAnswer(offer []byte, unicastAddress string, rtpListenerPort int) []byte {
	offerParsed := sdp.SessionDescription{}
	if err := offerParsed.Unmarshal(offer); err != nil {
		panic(err)
	}

	answer := sdp.SessionDescription{
		Version: 0,
		Origin: sdp.Origin{
			Username:       "-",
			SessionID:      offerParsed.Origin.SessionID,
			SessionVersion: offerParsed.Origin.SessionID + 2,
			NetworkType:    "IN",
			AddressType:    "IP4",
			UnicastAddress: unicastAddress,
		},
		SessionName: "Pion",
		ConnectionInformation: &sdp.ConnectionInformation{
			NetworkType: "IN",
			AddressType: "IP4",
			Address:     &sdp.Address{Address: unicastAddress},
		},
		TimeDescriptions: []sdp.TimeDescription{
			{
				Timing: sdp.Timing{
					StartTime: 0,
					StopTime:  0,
				},
			},
		},
		MediaDescriptions: []*sdp.MediaDescription{
			{
				MediaName: sdp.MediaName{
					Media:   "audio",
					Port:    sdp.RangedPort{Value: rtpListenerPort},
					Protos:  []string{"RTP", "AVP"},
					Formats: []string{"0"},
				},
				Attributes: []sdp.Attribute{
					{Key: "rtpmap", Value: "0 PCMU/8000"},
					{Key: "ptime", Value: "20"},
					{Key: "maxptime", Value: "150"},
					{Key: "recvonly"},
				},
			},
		},
	}

	answerByte, err := answer.Marshal()
	if err != nil {
		panic(err)
	}

	return answerByte
}

// Read from stdin until we get a newline.
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
