// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

// data-channels-detach is an example that shows how you can detach a data channel.
// This allows direct access the underlying [pion/datachannel](https://github.com/pion/datachannel).
// This allows you to interact with the data channel using a more idiomatic API based on
// the `io.ReadWriteCloser` interface.
package main

import (
	"bufio"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/ebitengine/debugui"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/pion/webrtc/v4"
)

type NetworkMessage struct {
	X float64
	Y float64
}

var img *ebiten.Image
var player_x float64 = 0
var player_y float64 = 0

var remote_player_x float64 = 0
var remote_player_y float64 = 0

const PlayerSpeed = 2

var isClient bool = false
var isHost bool = false

func init() {
	var err error
	img, _, err = ebitenutil.NewImageFromFile("gopher.png")
	if err != nil {
		log.Fatal(err)
	}
}

type Game struct {
	debugui        debugui.DebugUI
	peerConnection *webrtc.PeerConnection
}

func (g *Game) Update() error {
	// Update player position based on input
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		player_y -= PlayerSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		player_y += PlayerSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		player_x -= PlayerSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		player_x += PlayerSpeed
	}

	// ui stuff
	if _, err := g.debugui.Update(func(ctx *debugui.Context) error {
		ctx.Window("Test", image.Rect(60, 60, 160, 180), func(layout debugui.ContainerLayout) {
			ctx.Button("Host Button").On(func() {
				if !isHost {
					g.runHost()
					isHost = true
				}
			})
			ctx.Button("Client Button").On(func() {
				if !isClient {
					g.runClient()
					isClient = true
				}
			})
		})
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// local player
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(player_x, player_y)
	screen.DrawImage(img, op)
	// remote player
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(remote_player_x, remote_player_y)
	screen.DrawImage(img, op)

	// render debug UI
	g.debugui.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth, outsideHeight
}

func (g *Game) networkSetup() {
	// Since this behavior diverges from the WebRTC API it has to be
	// enabled using a settings engine. Mixing both detached and the
	// OnMessage DataChannel API is not supported.

	// Create a SettingEngine and enable Detach
	s := webrtc.SettingEngine{}
	s.DetachDataChannels()

	// Create an API object with the engine
	api := webrtc.NewAPI(webrtc.WithSettingEngine(s))

	// Everything below is the Pion WebRTC API! Thanks for using it ❤️.

	// Prepare the configuration
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	// Create a new RTCPeerConnection using the API object
	peerConnection, err := api.NewPeerConnection(config)
	if err != nil {
		panic(err)
	}
	/*
		defer func() {
			if cErr := peerConnection.Close(); cErr != nil {
				fmt.Printf("cannot close peerConnection: %v\n", cErr)
			}
		}()
	*/

	g.peerConnection = peerConnection

	// Set the handler for Peer connection state
	// This will notify you when the peer has connected/disconnected
	g.peerConnection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		fmt.Printf("Peer Connection State has changed: %s\n", state.String())

		if state == webrtc.PeerConnectionStateFailed {
			// Wait until PeerConnection has had no network activity for 30 seconds or another failure.
			// It may be reconnected using an ICE Restart.
			// Use webrtc.PeerConnectionStateDisconnected if you are interested in detecting faster timeout.
			// Note that the PeerConnection may come back from PeerConnectionStateDisconnected.
			fmt.Println("Peer Connection has gone to failed exiting")
			os.Exit(0)
		}

		if state == webrtc.PeerConnectionStateClosed {
			// PeerConnection was explicitly closed. This usually happens from a DTLS CloseNotify
			fmt.Println("Peer Connection has gone to closed exiting")
			os.Exit(0)
		}
	})
}

func (g *Game) runClient() {
	g.networkSetup()
	// Create a data channel with the default label and options
	dataChannel, err := g.peerConnection.CreateDataChannel("data", nil)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created DataChannel %s %d\n", dataChannel.Label(), dataChannel.ID())

	// Register channel opening handling
	dataChannel.OnOpen(func() {
		fmt.Printf("Data channel '%s'-'%d' open.\n", dataChannel.Label(), dataChannel.ID())

		// Detach the data channel
		raw, dErr := dataChannel.Detach()
		if dErr != nil {
			panic(dErr)
		}

		// Handle reading from the data channel
		go ReadLoop(raw)

		// Handle writing to the data channel
		go WriteLoop(raw)
	})

	offer, err := g.peerConnection.CreateOffer(nil)
	if err != nil {
		panic(err)
	}

	err = g.peerConnection.SetLocalDescription(offer)
	if err != nil {
		panic(err)
	}

	// Output the answer in base64 so we can paste it in browser
	fmt.Println("Printing SDP Offer, give this to the client:")
	fmt.Println(encode(&offer))

	// Wait for the answer to be pasted
	fmt.Println("Waiting for answer from client:")
	answer := webrtc.SessionDescription{}
	decode(readUntilNewline(), &answer)

	// Set the remote SessionDescription
	err = g.peerConnection.SetRemoteDescription(answer)
	if err != nil {
		panic(err)
	}

	fmt.Println("Remote description set, client should now be able to connect")
}

func (g *Game) runHost() {
	g.networkSetup()
	// callback for when we receive a new data channel
	// Register data channel creation handling
	g.peerConnection.OnDataChannel(func(dataChannel *webrtc.DataChannel) {
		fmt.Printf("New DataChannel %s %d\n", dataChannel.Label(), dataChannel.ID())

		// Register channel opening handling
		dataChannel.OnOpen(func() {
			fmt.Printf("Data channel '%s'-'%d' open.\n", dataChannel.Label(), dataChannel.ID())

			// Detach the data channel
			raw, dErr := dataChannel.Detach()
			if dErr != nil {
				panic(dErr)
			}

			// Handle reading from the data channel
			go ReadLoop(raw)

			// Handle writing to the data channel
			go WriteLoop(raw)
		})
	})
	fmt.Println("Waiting for SDP Offer from host:")
	// Wait for the offer to be pasted
	offer := webrtc.SessionDescription{}
	decode(readUntilNewline(), &offer)

	// Set the remote SessionDescription
	err := g.peerConnection.SetRemoteDescription(offer)
	if err != nil {
		panic(err)
	}

	// Create answer
	answer, err := g.peerConnection.CreateAnswer(nil)
	if err != nil {
		panic(err)
	}

	// Create channel that is blocked until ICE Gathering is complete
	gatherComplete := webrtc.GatheringCompletePromise(g.peerConnection)

	// Sets the LocalDescription, and starts our UDP listeners
	err = g.peerConnection.SetLocalDescription(answer)
	if err != nil {
		panic(err)
	}

	// Block until ICE Gathering is complete, disabling trickle ICE
	// we do this because we only can exchange one signaling message
	// in a production application you should exchange ICE Candidates via OnICECandidate
	<-gatherComplete

	// Output the answer in base64 so we can paste it in browser
	fmt.Println("Printing SDP Answer, give this to the host:")
	fmt.Println(encode(g.peerConnection.CurrentLocalDescription()))
}

func main() {
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Render an image")
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}

// ReadLoop shows how to read from the datachannel directly.
func ReadLoop(d io.Reader) {
	dec := gob.NewDecoder(d)
	for {
		var message NetworkMessage
		if err := dec.Decode(&message); err != nil {
			fmt.Println("Datachannel closed; Exit the readloop:", err)

			return
		}

		//fmt.Printf("Message from DataChannel: %#v\n", message)
		remote_player_x = message.X
		remote_player_y = message.Y
	}
}

// WriteLoop shows how to write to the datachannel directly.
func WriteLoop(d io.Writer) {
	enc := gob.NewEncoder(d)
	ticker := time.NewTicker(16 * time.Millisecond) // roughly 60 FPS
	defer ticker.Stop()
	for range ticker.C {
		message := NetworkMessage{
			X: player_x,
			Y: player_y,
		}
		//fmt.Printf("Sending %#v \n", message)
		if err := enc.Encode(message); err != nil {
			fmt.Println("Datachannel closed; Exit the writeloop:", err)
			panic(err)
		}
	}
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
