// SPDX-FileCopyrightText: 2025 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"os"
	"time"

	"github.com/ebitengine/debugui"
	"github.com/pion/webrtc/v4"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/kelindar/binary"
)

var img *ebiten.Image

var (
	posX       = 40.0
	posY       = 40.0
	remotePosX = 40.0
	remotePosY = 40.0
)

var lobbyID string

var signalingIP = "127.0.0.1"
var port = 3000

func getSignalingURL() string {
	return "http://" + signalingIP + ":" + strconv.Itoa(port)
}

// players registered by host.
var registeredPlayers = make(map[int]struct{})

// client to the HTTP signaling server.
var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

func init() {
	var err error
	img, _, err = ebitenutil.NewImageFromFile("gopher.png")
	if err != nil {
		log.Fatal(err)
	}
}

// implements ebiten.game interface.
type game struct {
	debugUI             debugui.DebugUI
	inputCapturingState debugui.InputCapturingState

	logBuf       string
	logSubmitBuf string
	logUpdated   bool

	lobbyID string
	isHost  bool

	localDebugInformation  string
	remoteDebugInformation string
}

func NewGame() (*game, error) {
	g := &game{}

	return g, nil
}

// Layout implements Game.
func (g *game) Layout(outsideWidth int, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

// called every tick (default 60 times a second)
// updates game logical state.
func (g *game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		posY--
	}

	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		posY++
	}

	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		posX--
	}

	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		posX++
	}

	inputCaptured, err := g.debugUI.Update(func(ctx *debugui.Context) error {
		g.logWindow(ctx)

		return nil
	})
	if err != nil {
		return err
	}
	g.inputCapturingState = inputCaptured
	return nil
}

// called every frame, depends on the monitor refresh rate
// which will probably be at least 60 times per second.
func (g *game) Draw(screen *ebiten.Image) {
	// prints something on the screen
	debugString := fmt.Sprintf("FPS: %f", ebiten.ActualFPS())
	debugString += "\n" + g.localDebugInformation + "\n" + g.remoteDebugInformation
	ebitenutil.DebugPrint(screen, debugString)

	// draw image
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(posX, posY)
	screen.DrawImage(img, op)

	// draw remote
	op2 := &ebiten.DrawImageOptions{}
	op2.GeoM.Translate(remotePosX, remotePosY)
	screen.DrawImage(img, op2)

	g.debugUI.Draw(screen)
}

var (
	// probably move all webrtc networking stuff to a struct i can manage.
	peerConnection *webrtc.PeerConnection
)

const messageSize = 32

type playerData struct {
	Id int
}

func (g *game) startConnection() {
	// Since this behavior diverges from the WebRTC API it has to be
	// enabled using a settings engine. Mixing both detached and the
	// OnMessage DataChannel API is not supported.

	// Create a SettingEngine and enable Detach.
	s := webrtc.SettingEngine{}
	s.DetachDataChannels()

	// Create an API object with the engine.
	api := webrtc.NewAPI(webrtc.WithSettingEngine(s))

	// Everything below is the Pion WebRTC API! Thanks for using it ❤️.

	// Prepare the configuration.
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	// Create a new RTCPeerConnection using the API object.
	pc, err := api.NewPeerConnection(config)
	if err != nil {
		panic(err)
	}

	// Set the global variable to the newly created RTCPeerConnection.
	peerConnection = pc

	// Set the handler for Peer connection state.
	// This will notify you when the peer has connected/disconnected.
	peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		g.writeLog(fmt.Sprintf("Peer Connection State has changed: %s\n", s.String()))

		if s == webrtc.PeerConnectionStateFailed {
			// Wait until PeerConnection has had no network activity for 30 seconds or another failure.
			// It may be reconnected using an ICE Restart.
			// Use webrtc.PeerConnectionStateDisconnected if you are interested in detecting faster timeout.
			// Note that the PeerConnection may come back from PeerConnectionStateDisconnected.
			g.writeLog(fmt.Sprintln("Peer Connection has gone to failed exiting"))
			os.Exit(0)
		}

		if s == webrtc.PeerConnectionStateClosed {
			// PeerConnection was explicitly closed. This usually happens from a DTLS CloseNotify
			g.writeLog(fmt.Sprintln("Peer Connection has gone to closed exiting"))
			os.Exit(0)
		}
	})

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected.
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		g.writeLog(fmt.Sprintf("ICE Connection State has changed: %s\n", connectionState.String()))
	})

	// the one that gives the answer is the host.
	if g.isHost { //nolint:nestif
		g.writeLog("Hosting a lobby")
		// Host creates lobby.
		lobbyResp, err := httpClient.Get(getSignalingURL() + "/lobby/host")
		if err != nil {
			panic(err)
		}
		bodyBytes, err := io.ReadAll(lobbyResp.Body)
		if err != nil {
			panic(err)
		}
		lobbyID = string(bodyBytes)
		lobbyIDStr := fmt.Sprintf("Lobby ID: %s\n", lobbyID)
		g.writeLog(lobbyIDStr)

		// Register data channel creation handling.
		peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {
			g.writeLog(fmt.Sprintf("New DataChannel %s %d\n", d.Label(), d.ID()))

			// Register channel opening handling.
			d.OnOpen(func() {
				s := fmt.Sprintf("Data channel '%s'-'%d' open on host side!", d.Label(), d.ID())
				g.writeLog(s)

				// Detach the data channel.
				raw, dErr := d.Detach()
				if dErr != nil {
					panic(dErr)
				}

				// Handle reading from the data channel.
				go ReadLoop(g, raw)

				// Handle writing to the data channel.
				go WriteLoop(g, raw)
			})
		})

		// poll for offer from signaling server for player.
		pollForPlayerOffer := func(playerID int) {
			ticker := time.NewTicker(1 * time.Second)
			for range ticker.C {
				g.writeLog(fmt.Sprintf("Polling for offer for %d\n", playerID))
				// hardcode that there is only one other player and they have player_id 1.
				getUrl := getSignalingURL() + "/offer/get?lobby_id=" + lobbyID + "&player_id=" + strconv.Itoa(playerID)
				g.writeLog(fmt.Sprintln(getUrl))
				offerResp, err := httpClient.Get(getUrl)
				if err != nil {
					panic(err)
				}
				if offerResp.StatusCode != http.StatusOK {
					continue
				}
				body := new(bytes.Buffer)
				_, err = body.ReadFrom(offerResp.Body)
				if err != nil {
					panic(err)
				}

				g.writeLog(fmt.Sprintf("Got offer %v\n", body.String()))
				offer := webrtc.SessionDescription{}
				err = json.NewDecoder(body).Decode(&offer)
				if err != nil {
					panic(err)
				}
				// Set the remote SessionDescription.
				err = peerConnection.SetRemoteDescription(offer)
				if err != nil {
					panic(err)
				}
				// Create answer.
				answer, err := peerConnection.CreateAnswer(nil)
				if err != nil {
					panic(err)
				}

				// Create channel that is blocked until ICE Gathering is complete.
				gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

				// Sets the LocalDescription, and starts our UDP listeners.
				err = peerConnection.SetLocalDescription(answer)
				if err != nil {
					panic(err)
				}

				// Block until ICE Gathering is complete, disabling trickle ICE.
				// we do this because we only can exchange one signaling message
				// in a production application you should exchange ICE Candidates via OnICECandidate.
				<-gatherComplete
				// send answer we generated to the signaling server.
				answerJson, err := json.Marshal(peerConnection.LocalDescription())
				if err != nil {
					panic(err)
				}
				postUrl := getSignalingURL() + "/answer/post?lobby_id=" + lobbyID + "&player_id=" + strconv.Itoa(playerID)
				g.writeLog(fmt.Sprintln(postUrl))
				_, err = httpClient.Post(postUrl, "application/json", bytes.NewBuffer(answerJson))
				if err != nil {
					panic(err)
				}

				// if we have successfully set the remote description, we can break out of the loop.
				ticker.Stop()

				return
			}
		}

		go func() {
			ticker := time.NewTicker(1 * time.Second)
			for t := range ticker.C {
				g.writeLog(fmt.Sprintln("Polling for lobby ID {", lobbyID, "} at", t))
				idUrl := getSignalingURL() + "/lobby/unregisteredPlayers?id=" + lobbyID
				g.writeLog(fmt.Sprintln(idUrl))
				idResp, err := httpClient.Get(idUrl)
				if err != nil {
					panic(err)
				}
				if idResp.StatusCode != http.StatusOK {
					continue
				}
				var playerIds []int
				err = json.NewDecoder(idResp.Body).Decode(&playerIds)
				if err != nil {
					panic(err)
				}
				g.writeLog(fmt.Sprintf("Player IDs: %v\n", playerIds))
				// poll for all of the unregistered players.
				for _, playerID := range playerIds {
					// only start goroutine if playerID hasn't been registered yet.
					if _, ok := registeredPlayers[playerID]; !ok {
						registeredPlayers[playerID] = struct{}{}
						go pollForPlayerOffer(playerID)
					}
				}
			}
		}()
	} else {
		g.writeLog("Joining lobby: " + lobbyID)
		// the following is for the client joining the lobby.
		// get lobby id from text input.
		lobbyID = g.lobbyID
		response, err := httpClient.Get(getSignalingURL() + "/lobby/join?id=" + lobbyID)
		if err != nil {
			panic(err)
		}
		var playerData playerData
		err = json.NewDecoder(response.Body).Decode(&playerData)
		if err != nil {
			panic(err)
		}
		g.writeLog(fmt.Sprintf("Player ID: %v\n", playerData))
		// Create a datachannel with label 'data'.
		dataChannel, err := peerConnection.CreateDataChannel("data", nil)
		if err != nil {
			panic(err)
		}

		// Register channel opening handling.
		dataChannel.OnOpen(func() {
			s := fmt.Sprintf("Data channel '%s'-'%d' open on client side!", dataChannel.Label(), dataChannel.ID())
			g.writeLog(s)

			// Detach the data channel.
			raw, dErr := dataChannel.Detach()
			if dErr != nil {
				panic(dErr)
			}

			// Handle reading from the data channel.
			go ReadLoop(g, raw)

			// Handle writing to the data channel.
			go WriteLoop(g, raw)
		})

		// Create an offer to send to the browser.
		offer, err := peerConnection.CreateOffer(nil)
		if err != nil {
			panic(err)
		}

		// Sets the LocalDescription, and starts our UDP listeners.
		err = peerConnection.SetLocalDescription(offer)
		if err != nil {
			panic(err)
		}

		// print out possible offers from different ICE Candidates.
		peerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
			if candidate != nil {
				offerJson, err := json.Marshal(peerConnection.LocalDescription())
				if err != nil {
					panic(err)
				}
				postUrl := getSignalingURL() + "/offer/post?lobby_id=" + lobbyID + "&player_id=" + strconv.Itoa(playerData.Id)
				g.writeLog(fmt.Sprintln(postUrl))
				_, err = httpClient.Post(postUrl, "application/json", bytes.NewBuffer(offerJson))
				if err != nil {
					panic(err)
				}
			}
		})

		answer := webrtc.SessionDescription{}
		// read answer from other peer (wait till we actually get something).
		ticker := time.NewTicker(1 * time.Second)
		go func() {
			for range ticker.C {
				g.writeLog(fmt.Sprintln("Polling for answer"))
				url := getSignalingURL() + "/answer/get?lobby_id=" + lobbyID + "&player_id=" + strconv.Itoa(playerData.Id)
				fmt.Println(url)
				answerResp, err := httpClient.Get(url)
				if err != nil {
					panic(err)
				}
				if answerResp.StatusCode != http.StatusOK {
					continue
				}
				body := new(bytes.Buffer)
				body.ReadFrom(answerResp.Body)
				g.writeLog(fmt.Sprintf("Got answer %v\n", body.String()))
				err = json.NewDecoder(body).Decode(&answer)
				if err != nil {
					panic(err)
				}

				if err := peerConnection.SetRemoteDescription(answer); err != nil {
					panic(err)
				}

				// if we have successfully set the remote description, we can break out of the loop.
				ticker.Stop()

				return
			}
		}()
	}
}

func (g *game) closeConnection() {
	if cErr := peerConnection.Close(); cErr != nil {
		fmt.Printf("cannot close peerConnection: %v\n", cErr)
	}
	// this doesn't work, fix this.
	if g.isHost {
		// delete lobby if host.
		url := getSignalingURL() + "/lobby/delete"
		fmt.Println(url)
		_, err := httpClient.Get(url)
		if err != nil {
			panic(err)
		}
	}
}

// entry point of the program.
func main() {
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Hello, World!")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	g, err := NewGame()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := ebiten.RunGame(g); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// close the connection when the game ends.
	g.closeConnection()
}

type Packet struct {
	PosX float64
	PosY float64
}

// ReadLoop shows how to read from the datachannel directly.
func ReadLoop(game *game, d io.Reader) {
	for {
		buffer := make([]byte, messageSize)
		_, err := io.ReadFull(d, buffer)
		if err != nil {
			game.writeLog(fmt.Sprintln("Datachannel closed; Exit the readloop:", err))

			return
		}

		var packet Packet
		err = binary.Unmarshal(buffer, &packet)
		if err != nil {
			panic(err)
		}

		remotePosX = packet.PosX
		remotePosY = packet.PosY

		game.remoteDebugInformation = fmt.Sprintf("Message from DataChannel: %f %f", packet.PosX, packet.PosY)
	}
}

// WriteLoop shows how to write to the datachannel directly.
func WriteLoop(g *game, d io.Writer) {
	ticker := time.NewTicker(time.Millisecond * 20)
	defer ticker.Stop()
	for range ticker.C {
		packet := &Packet{posX, posY}
		g.localDebugInformation = fmt.Sprintf("Sending x:%f y:%f", packet.PosX, packet.PosY)
		encoded, err := binary.Marshal(packet)
		if err != nil {
			panic(err)
		}

		if _, err := d.Write(encoded); err != nil {
			panic(err)
		}
	}
}
