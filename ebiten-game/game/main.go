// SPDX-FileCopyrightText: 2026 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package main

//nolint:gci
import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "image/jpeg"
	_ "image/png"

	"github.com/ebitengine/debugui"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/kelindar/binary"
	"github.com/pion/webrtc/v4"
)

func (g *game) getSignalingURL() string {
	return "http://" + g.signalingIP + ":" + strconv.Itoa(g.port)
}

// implements ebiten.game interface.
type game struct {
	debugUI             debugui.DebugUI
	inputCapturingState debugui.InputCapturingState

	logBuf       string
	logSubmitBuf string
	logUpdated   bool

	lobbyID           string
	isHost            bool
	registeredPlayers map[int]struct{}
	httpClient        *http.Client
	peerConnection    *webrtc.PeerConnection

	localDebugInformation  string
	remoteDebugInformation string

	img *ebiten.Image

	localPlayerID int
	posX          float64
	posY          float64
	remotePosX    float64
	remotePosY    float64

	signalingIP string
	port        int
}

func NewGame() (*game, error) {
	img, _, err := ebitenutil.NewImageFromFile("gopher.png")
	if err != nil {
		log.Fatal(err)
	}
	game := &game{
		posX:              40,
		posY:              40,
		remotePosX:        40,
		remotePosY:        40,
		img:               img,
		signalingIP:       "127.0.0.1",
		port:              3000,
		registeredPlayers: make(map[int]struct{}),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	return game, err
}

// Layout implements Game.
func (g *game) Layout(outsideWidth int, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

// called every tick (default 60 times a second)
// updates game logical state.
func (g *game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		g.posY--
	}

	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		g.posY++
	}

	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.posX--
	}

	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.posX++
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
	op.GeoM.Translate(g.posX, g.posY)
	screen.DrawImage(g.img, op)

	// draw remote
	op2 := &ebiten.DrawImageOptions{}
	op2.GeoM.Translate(g.remotePosX, g.remotePosY)
	screen.DrawImage(g.img, op2)

	g.debugUI.Draw(screen)
}

const messageSize = 32

type playerData struct {
	ID int
}

// nolint:cyclop
// poll for offer from signaling server for player.
func (g *game) pollForPlayerOffer(playerID int, ticker *time.Ticker) {
	for range ticker.C {
		g.writeLog(fmt.Sprintf("Polling for offer for %d\n", playerID))
		// hardcode that there is only one other player and they have player_id 1.
		getURL := g.getSignalingURL() + "/offer/get?lobby_id=" + g.lobbyID + "&player_id=" + strconv.Itoa(playerID)
		g.writeLog(fmt.Sprintln(getURL))
		req, err := http.NewRequestWithContext(context.Background(), "GET", getURL, nil)
		if err != nil {
			panic(err)
		}
		offerResp, err := g.httpClient.Do(req)
		if err != nil {
			panic(err)
		}
		// if we don't have an offer yet, continue polling.
		// nolint:nestif
		if offerResp.StatusCode != http.StatusOK {
			err = offerResp.Body.Close() // close body to avoid resource leak.
			if err != nil {
				panic(err)
			}
		} else {
			// we have now received an offer, we can now give an answer back
			body := new(bytes.Buffer)
			_, err = body.ReadFrom(offerResp.Body)
			if err != nil {
				panic(err)
			}
			err = offerResp.Body.Close() // close body to avoid resource leak.
			if err != nil {
				panic(err)
			}
			// we have an offer.
			g.writeLog(fmt.Sprintf("Got offer %v\n", body.String()))
			offer := webrtc.SessionDescription{}
			err = json.NewDecoder(body).Decode(&offer)
			if err != nil {
				panic(err)
			}
			// Set the remote SessionDescription.
			err = g.peerConnection.SetRemoteDescription(offer)
			if err != nil {
				panic(err)
			}
			// Create answer.
			answer, err := g.peerConnection.CreateAnswer(nil)
			if err != nil {
				panic(err)
			}

			// Create channel that is blocked until ICE Gathering is complete.
			gatherComplete := webrtc.GatheringCompletePromise(g.peerConnection)

			// Sets the LocalDescription, and starts our UDP listeners.
			err = g.peerConnection.SetLocalDescription(answer)
			if err != nil {
				panic(err)
			}

			// Block until ICE Gathering is complete, disabling trickle ICE.
			// we do this because we only can exchange one signaling message
			// in a production application you should exchange ICE Candidates via OnICECandidate.
			<-gatherComplete
			// send answer we generated to the signaling server.
			answerJSON, err := json.Marshal(g.peerConnection.LocalDescription())
			if err != nil {
				panic(err)
			}
			postURL := g.getSignalingURL() + "/answer/post?lobby_id=" + g.lobbyID + "&player_id=" + strconv.Itoa(playerID)
			g.writeLog(fmt.Sprintln(postURL))
			postReq, err := http.NewRequestWithContext(context.Background(), "POST", postURL, bytes.NewBuffer(answerJSON))
			if err != nil {
				panic(err)
			}
			postReq.Header.Set("Content-Type", "application/json")
			postResponse, err := g.httpClient.Do(postReq)
			if err != nil {
				panic(err)
			}

			err = postResponse.Body.Close() // close body to avoid resource leak.
			if err != nil {
				panic(err)
			}

			return
		}
	}
}

func (g *game) pollLobbyAsHost() {
	ticker := time.NewTicker(1 * time.Second)
	for t := range ticker.C {
		g.writeLog(fmt.Sprintln("Polling for lobby ID {", g.lobbyID, "} at", t))
		idURL := g.getSignalingURL() + "/lobby/unregisteredPlayers?id=" + g.lobbyID
		g.writeLog(fmt.Sprintln(idURL))
		idReq, err := http.NewRequestWithContext(context.Background(), "GET", idURL, nil)
		if err != nil {
			panic(err)
		}
		idResp, err := g.httpClient.Do(idReq)
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
		err = idResp.Body.Close() // close body to avoid resource leak.
		if err != nil {
			panic(err)
		}
		g.writeLog(fmt.Sprintf("Player IDs: %v\n", playerIds))
		// poll for all of the unregistered players.
		for _, playerID := range playerIds {
			// only start goroutine if playerID hasn't been registered yet.
			if _, ok := g.registeredPlayers[playerID]; !ok {
				g.registeredPlayers[playerID] = struct{}{}
				ticker := time.NewTicker(1 * time.Second)
				go g.pollForPlayerOffer(playerID, ticker)
			}
		}
	}
}

func (g *game) onHostReceivedDataChannel(d *webrtc.DataChannel) {
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
}

func (g *game) startHost() {
	g.writeLog("Hosting a lobby")
	// Host creates lobby.
	req, err := http.NewRequestWithContext(context.Background(), "GET", g.getSignalingURL()+"/lobby/host", nil)
	if err != nil {
		panic(err)
	}
	lobbyResp, err := g.httpClient.Do(req)
	if err != nil {
		panic(err)
	}
	bodyBytes, err := io.ReadAll(lobbyResp.Body)
	if err != nil {
		panic(err)
	}
	g.lobbyID = string(bodyBytes)
	lobbyIDStr := fmt.Sprintf("Lobby ID: %s\n", g.lobbyID)
	g.writeLog(lobbyIDStr)

	err = lobbyResp.Body.Close()
	if err != nil {
		panic(err)
	}

	// Register data channel creation handling.
	g.peerConnection.OnDataChannel(g.onHostReceivedDataChannel)

	go g.pollLobbyAsHost()
}

func (g *game) pollLobbyAsClient(ticker *time.Ticker, pData playerData) {
	answer := webrtc.SessionDescription{}
	for range ticker.C {
		g.writeLog(fmt.Sprintln("Polling for answer"))
		URL := g.getSignalingURL() + "/answer/get?lobby_id=" + g.lobbyID + "&player_id=" + strconv.Itoa(pData.ID)
		fmt.Println(URL)
		answerReq, err := http.NewRequestWithContext(context.Background(), "GET", URL, nil)
		if err != nil {
			panic(err)
		}
		answerResp, err := g.httpClient.Do(answerReq)
		if err != nil {
			panic(err)
		}
		if answerResp.StatusCode != http.StatusOK {
			continue
		}
		body := new(bytes.Buffer)
		_, err = body.ReadFrom(answerResp.Body)
		if err != nil {
			panic(err)
		}
		g.writeLog(fmt.Sprintf("Got answer %v\n", body.String()))
		err = json.NewDecoder(body).Decode(&answer)
		if err != nil {
			panic(err)
		}

		err = answerResp.Body.Close() // close body to avoid resource leak.
		if err != nil {
			panic(err)
		}

		if err := g.peerConnection.SetRemoteDescription(answer); err != nil {
			panic(err)
		}

		// if we have successfully set the remote description, we can break out of the loop.
		ticker.Stop()

		return
	}
}

func (g *game) onClientReceivedICECandidate(candidate *webrtc.ICECandidate) {
	if candidate != nil {
		offerJSON, err := json.Marshal(g.peerConnection.LocalDescription())
		if err != nil {
			panic(err)
		}
		postURL := g.getSignalingURL() + "/offer/post?lobby_id=" + g.lobbyID + "&player_id=" + strconv.Itoa(g.localPlayerID)
		g.writeLog(fmt.Sprintln(postURL))
		postReq, err := http.NewRequestWithContext(context.Background(), "POST", postURL, bytes.NewBuffer(offerJSON))
		if err != nil {
			panic(err)
		}
		postReq.Header.Set("Content-Type", "application/json")
		postResponse, err := g.httpClient.Do(postReq)
		if err != nil {
			panic(err)
		}
		err = postResponse.Body.Close() // close body to avoid resource leak.
		if err != nil {
			panic(err)
		}
	}
}

func (g *game) startClient() {
	g.writeLog("Joining lobby: " + g.lobbyID)
	// the following is for the client joining the lobby.
	// get lobby id from text input.
	URL := g.getSignalingURL() + "/lobby/join?id=" + g.lobbyID
	joinReq, err := http.NewRequestWithContext(context.Background(), "GET", URL, nil)
	if err != nil {
		panic(err)
	}
	response, err := g.httpClient.Do(joinReq)
	if err != nil {
		panic(err)
	}
	if response.StatusCode != http.StatusOK {
		g.writeLog("Failed to join lobby, probably doesn't exist.\n")

		return
	}
	var pData playerData
	err = json.NewDecoder(response.Body).Decode(&pData)
	if err != nil {
		panic(err)
	}
	err = response.Body.Close() // close body to avoid resource leak.
	if err != nil {
		panic(err)
	}
	g.localPlayerID = pData.ID
	g.writeLog(fmt.Sprintf("Player ID: %v\n", g.localPlayerID))
	// Create a datachannel with label 'data'.
	dataChannel, err := g.peerConnection.CreateDataChannel("data", nil)
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
	offer, err := g.peerConnection.CreateOffer(nil)
	if err != nil {
		panic(err)
	}

	// Sets the LocalDescription, and starts our UDP listeners.
	err = g.peerConnection.SetLocalDescription(offer)
	if err != nil {
		panic(err)
	}

	// print out possible offers from different ICE Candidates.
	g.peerConnection.OnICECandidate(g.onClientReceivedICECandidate)

	// read answer from other peer (wait till we actually get something).
	ticker := time.NewTicker(1 * time.Second)
	go g.pollLobbyAsClient(ticker, pData)
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
	g.peerConnection = pc

	// Set the handler for Peer connection state.
	// This will notify you when the peer has connected/disconnected.
	g.peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
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
	g.peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		g.writeLog(fmt.Sprintf("ICE Connection State has changed: %s\n", connectionState.String()))
	})

	// the one that gives the answer is the host.
	if g.isHost { //nolint:nestif
		g.startHost()
	} else {
		g.startClient()
	}
}

func (g *game) closeConnection() {
	if cErr := g.peerConnection.Close(); cErr != nil {
		fmt.Printf("cannot close peerConnection: %v\n", cErr)
	}
	// this doesn't work, fix this.
	if g.isHost {
		// delete lobby if host.
		URL := g.getSignalingURL() + "/lobby/delete"
		fmt.Println(URL)
		hostReq, err := http.NewRequestWithContext(context.Background(), "GET", URL, nil)
		if err != nil {
			panic(err)
		}
		hostResponse, err := g.httpClient.Do(hostReq)
		if err != nil {
			panic(err)
		}
		err = hostResponse.Body.Close()
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

	game, err := NewGame()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := ebiten.RunGame(game); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	// close the connection when the game ends.
	game.closeConnection()
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

		game.remotePosX = packet.PosX
		game.remotePosY = packet.PosY

		game.remoteDebugInformation = fmt.Sprintf("Message from DataChannel: %f %f", packet.PosX, packet.PosY)
	}
}

// WriteLoop shows how to write to the datachannel directly.
func WriteLoop(g *game, d io.Writer) {
	ticker := time.NewTicker(time.Millisecond * 20)
	defer ticker.Stop()
	for range ticker.C {
		packet := &Packet{g.posX, g.posY}
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
