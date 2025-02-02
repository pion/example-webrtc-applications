// SPDX-FileCopyrightText: 2026 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

//go:build !js

// text-to-speech demonstrates Text-to-Speech using eSpeak NG and Pion's pure Go Opus encoder.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"time"

	"github.com/pion/opus"
	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"
)

func doSignaling(res http.ResponseWriter, req *http.Request) { //nolint:cyclop
	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		panic(err)
	}

	audioTrack, err := webrtc.NewTrackLocalStaticSample(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus},
		"audio",
		"pion",
	)
	if err != nil {
		panic(err)
	}

	if _, err = peerConnection.AddTrack(audioTrack); err != nil {
		panic(err)
	}

	wavAudio := make(chan []byte, 16)
	speechContext := context.WithoutCancel(req.Context())

	onDataChannelHandler := func(dataChannel *webrtc.DataChannel) {
		dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
			commandContext, cancel := context.WithTimeout(speechContext, 30*time.Second)
			defer cancel()

			cmd := exec.CommandContext( //nolint:gosec // The text is an argument to eSpeak, not a shell command.
				commandContext,
				"espeak-ng",
				"--stdout",
				"-v", "en-us",
				string(msg.Data),
			)

			wav, commandErr := cmd.Output()
			if commandErr != nil {
				panic(commandErr)
			}

			wavAudio <- wav
		})
	}
	peerConnection.OnDataChannel(onDataChannelHandler)
	go writeAudio(audioTrack, wavAudio)

	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("ICE Connection State has changed: %s\n", connectionState.String())
	})

	var offer webrtc.SessionDescription
	if err = json.NewDecoder(req.Body).Decode(&offer); err != nil {
		panic(err)
	}

	if err = peerConnection.SetRemoteDescription(offer); err != nil {
		panic(err)
	}

	// Create channel that is blocked until ICE Gathering is complete
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		panic(err)
	}
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

	res.Header().Set("Content-Type", "application/json")
	if _, err = res.Write(response); err != nil {
		panic(err)
	}
}

func writeAudio(audioTrack *webrtc.TrackLocalStaticSample, wavAudio <-chan []byte) {
	ticker := time.NewTicker(time.Millisecond * 20)
	defer ticker.Stop()

	encoder, err := opus.NewEncoder()
	if err != nil {
		panic(err)
	}
	encodedAudio := make([]byte, 1275)
	frame := make([]byte, 960*2)
	var pendingAudio []byte

	for range ticker.C {
		clear(frame)
		if len(pendingAudio) == 0 {
			select {
			case wav := <-wavAudio:
				pendingAudio = wavToPCM48kMono(wav)
			default:
			}
		}

		copied := copy(frame, pendingAudio)
		pendingAudio = pendingAudio[copied:]

		encodedLen, encodeErr := encoder.Encode(frame, encodedAudio)
		if encodeErr != nil {
			panic(encodeErr)
		}
		if writeErr := audioTrack.WriteSample(media.Sample{
			Data:     encodedAudio[:encodedLen],
			Duration: 20 * time.Millisecond,
		}); writeErr != nil {
			panic(writeErr)
		}
	}
}

func wavToPCM48kMono(wav []byte) []byte {
	// eSpeak returns mono 16-bit PCM at 22.05 kHz after a 44-byte WAV header.
	const (
		wavHeaderSize  = 44
		bytesPerSample = 2
		wavSampleRate  = 22050
		opusSampleRate = 48000
	)

	pcm := wav[wavHeaderSize:]
	inputSamples := len(pcm) / bytesPerSample
	outputSamples := inputSamples * opusSampleRate / wavSampleRate
	output := make([]byte, outputSamples*bytesPerSample)

	for outputSample := range outputSamples {
		inputSample := outputSample * wavSampleRate / opusSampleRate
		inputOffset := inputSample * bytesPerSample
		outputOffset := outputSample * bytesPerSample
		copy(output[outputOffset:outputOffset+bytesPerSample], pcm[inputOffset:inputOffset+bytesPerSample])
	}

	return output
}

func main() {
	http.Handle("/", http.FileServer(http.Dir(".")))
	http.HandleFunc("/doSignaling", doSignaling)

	fmt.Println("Open http://localhost:8080 to access this demo")
	// nolint: gosec
	panic(http.ListenAndServe(":8080", nil))
}
