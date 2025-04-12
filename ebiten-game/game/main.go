// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	//"runtime"

	//"github.com/pion/randutil"

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
	//"github.com/hajimehoshi/ebiten/v2/inpututil"
)

var img *ebiten.Image

var (
	pos_x        = 40.0
	pos_y        = 40.0
	remote_pos_x = 40.0
	remote_pos_y = 40.0
)

var lobby_id string

var signalingIP = "127.0.0.1"
var port = 3000

func getSignalingURL() string {
	return "http://" + signalingIP + ":" + strconv.Itoa(port)
}

// players registered by host
var registered_players = make(map[int]struct{})

// client to the HTTP signaling server
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

// implements ebiten.Game interface
type Game struct {
	debugUI             debugui.DebugUI
	inputCapturingState debugui.InputCapturingState

	logBuf       string
	logSubmitBuf string
	logUpdated   bool

	lobby_id string
	isHost   bool

	localDebugInformation  string
	remoteDebugInformation string
}

func NewGame() (*Game, error) {
	g := &Game{}

	return g, nil
}

// Layout implements Game.
func (g *Game) Layout(outsideWidth int, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

// called every tick (default 60 times a second)
// updates game logical state
func (g *Game) Update() error {

	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		pos_y -= 1
	}

	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		pos_y += 1
	}

	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		pos_x -= 1
	}

	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		pos_x += 1
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
// which will probably be at least 60 times per second
func (g *Game) Draw(screen *ebiten.Image) {
	// prints something on the screen
	debugString := fmt.Sprintf("FPS: %f", ebiten.ActualFPS())
	debugString += "\n" + g.localDebugInformation + "\n" + g.remoteDebugInformation
	ebitenutil.DebugPrint(screen, debugString)

	// draw image
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(pos_x, pos_y)
	screen.DrawImage(img, op)

	// draw remote
	op2 := &ebiten.DrawImageOptions{}
	op2.GeoM.Translate(remote_pos_x, remote_pos_y)
	screen.DrawImage(img, op2)

	g.debugUI.Draw(screen)
}

var (
	// probably move all webrtc networking stuff to a struct i can manage
	peerConnection *webrtc.PeerConnection
)

const messageSize = 32

type PlayerData struct {
	Id int
}

func (game *Game) startConnection() {
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
	pc, err := api.NewPeerConnection(config)
	if err != nil {
		panic(err)
	}

	// Set the global variable to the newly created RTCPeerConnection
	peerConnection = pc

	// Set the handler for Peer connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		game.writeLog(fmt.Sprintf("Peer Connection State has changed: %s\n", s.String()))

		if s == webrtc.PeerConnectionStateFailed {
			// Wait until PeerConnection has had no network activity for 30 seconds or another failure. It may be reconnected using an ICE Restart.
			// Use webrtc.PeerConnectionStateDisconnected if you are interested in detecting faster timeout.
			// Note that the PeerConnection may come back from PeerConnectionStateDisconnected.
			game.writeLog(fmt.Sprintln("Peer Connection has gone to failed exiting"))
			os.Exit(0)
		}

		if s == webrtc.PeerConnectionStateClosed {
			// PeerConnection was explicitly closed. This usually happens from a DTLS CloseNotify
			game.writeLog(fmt.Sprintln("Peer Connection has gone to closed exiting"))
			os.Exit(0)
		}
	})

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		game.writeLog(fmt.Sprintf("ICE Connection State has changed: %s\n", connectionState.String()))
	})

	// the one that gives the answer is the host
	if game.isHost {
		game.writeLog("Hosting a lobby")
		// Host creates lobby
		lobby_resp, err := httpClient.Get(getSignalingURL() + "/lobby/host")
		if err != nil {
			panic(err)
		}
		bodyBytes, err := io.ReadAll(lobby_resp.Body)
		if err != nil {
			panic(err)
		}
		lobby_id = string(bodyBytes)
		lobby_id_str := fmt.Sprintf("Lobby ID: %s\n", lobby_id)
		game.writeLog(lobby_id_str)

		// Register data channel creation handling
		peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {
			game.writeLog(fmt.Sprintf("New DataChannel %s %d\n", d.Label(), d.ID()))

			// Register channel opening handling
			d.OnOpen(func() {

				s := fmt.Sprintf("Data channel '%s'-'%d' open on host side!", d.Label(), d.ID())
				game.writeLog(s)

				// Detach the data channel
				raw, dErr := d.Detach()
				if dErr != nil {
					panic(dErr)
				}

				// Handle reading from the data channel
				go ReadLoop(game, raw)

				// Handle writing to the data channel
				go WriteLoop(game, raw)
			})
		})

		// poll for offer from signaling server for player
		pollForPlayerOffer := func(player_id int) {
			ticker := time.NewTicker(1 * time.Second)
			for {
				select {
				case t := <-ticker.C:
					game.writeLog(fmt.Sprintln("Tick at", t))
					game.writeLog(fmt.Sprintf("Polling for offer for %d\n", player_id))
					// hardcode that there is only one other player and they have player_id 1
					getUrl := getSignalingURL() + "/offer/get?lobby_id=" + lobby_id + "&player_id=" + strconv.Itoa(player_id)
					game.writeLog(fmt.Sprintln(getUrl))
					offer_resp, err := httpClient.Get(getUrl)
					if err != nil {
						panic(err)
					}
					if offer_resp.StatusCode != http.StatusOK {
						continue
					}
					body := new(bytes.Buffer)
					body.ReadFrom(offer_resp.Body)
					game.writeLog(fmt.Sprintf("Got offer %v\n", body.String()))
					offer := webrtc.SessionDescription{}
					err = json.NewDecoder(body).Decode(&offer)
					if err != nil {
						panic(err)
					}
					// Set the remote SessionDescription
					err = peerConnection.SetRemoteDescription(offer)
					if err != nil {
						panic(err)
					}
					// Create answer
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
					// send answer we generated to the signaling server
					answerJson, err := json.Marshal(peerConnection.LocalDescription())
					if err != nil {
						panic(err)
					}
					postUrl := getSignalingURL() + "/answer/post?lobby_id=" + lobby_id + "&player_id=" + strconv.Itoa(player_id)
					game.writeLog(fmt.Sprintln(postUrl))
					httpClient.Post(postUrl, "application/json", bytes.NewBuffer(answerJson))
					// if we have successfully set the remote description, we can break out of the loop
					ticker.Stop()
					return
				}
			}
		}

		go func() {
			ticker := time.NewTicker(1 * time.Second)
			for {
				select {
				case t := <-ticker.C:
					game.writeLog(fmt.Sprintln("Polling for lobby ID {", lobby_id, "} at", t))
					idUrl := getSignalingURL() + "/lobby/unregisteredPlayers?id=" + lobby_id
					game.writeLog(fmt.Sprintln(idUrl))
					id_resp, err := httpClient.Get(idUrl)
					if err != nil {
						panic(err)
					}
					if id_resp.StatusCode != http.StatusOK {
						continue
					}
					var player_ids []int
					err = json.NewDecoder(id_resp.Body).Decode(&player_ids)
					if err != nil {
						panic(err)
					}
					game.writeLog(fmt.Sprintf("Player IDs: %v\n", player_ids))
					// poll for all of the unregistered players
					for _, player_id := range player_ids {
						// only start goroutine if player_id hasn't been registered yet
						if _, ok := registered_players[player_id]; !ok {
							registered_players[player_id] = struct{}{}
							go pollForPlayerOffer(player_id)
						}
					}
				}
			}
		}()
	} else {
		game.writeLog("Joining lobby: " + lobby_id)
		// the following is for the client joining the lobby
		// get lobby id from text input
		lobby_id = game.lobby_id
		response, err := httpClient.Get(getSignalingURL() + "/lobby/join?id=" + lobby_id)
		if err != nil {
			panic(err)
		}
		var player_data PlayerData
		err = json.NewDecoder(response.Body).Decode(&player_data)
		if err != nil {
			panic(err)
		}
		game.writeLog(fmt.Sprintf("Player ID: %v\n", player_data))
		// Create a datachannel with label 'data'
		dataChannel, err := peerConnection.CreateDataChannel("data", nil)
		if err != nil {
			panic(err)
		}

		// Register channel opening handling
		dataChannel.OnOpen(func() {
			s := fmt.Sprintf("Data channel '%s'-'%d' open on client side!", dataChannel.Label(), dataChannel.ID())
			game.writeLog(s)

			// Detach the data channel
			raw, dErr := dataChannel.Detach()
			if dErr != nil {
				panic(dErr)
			}

			// Handle reading from the data channel
			go ReadLoop(game, raw)

			// Handle writing to the data channel
			go WriteLoop(game, raw)
		})

		// Create an offer to send to the browser
		offer, err := peerConnection.CreateOffer(nil)
		if err != nil {
			panic(err)
		}

		// Sets the LocalDescription, and starts our UDP listeners
		err = peerConnection.SetLocalDescription(offer)
		if err != nil {
			panic(err)
		}

		// print out possible offers from different ICE Candidates
		peerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
			if candidate != nil {
				offerJson, err := json.Marshal(peerConnection.LocalDescription())
				if err != nil {
					panic(err)
				}
				postUrl := getSignalingURL() + "/offer/post?lobby_id=" + lobby_id + "&player_id=" + strconv.Itoa(player_data.Id)
				game.writeLog(fmt.Sprintln(postUrl))
				httpClient.Post(postUrl, "application/json", bytes.NewBuffer(offerJson))
			}
		})

		answer := webrtc.SessionDescription{}
		// read answer from other peer (wait till we actually get something)
		ticker := time.NewTicker(1 * time.Second)
		go func() {
			for {
				select {
				case t := <-ticker.C:
					game.writeLog(fmt.Sprintln("Tick at", t))
					game.writeLog(fmt.Sprintln("Polling for answer"))
					url := getSignalingURL() + "/answer/get?lobby_id=" + lobby_id + "&player_id=" + strconv.Itoa(player_data.Id)
					fmt.Println(url)
					answer_resp, err := httpClient.Get(url)
					if err != nil {
						panic(err)
					}
					if answer_resp.StatusCode != http.StatusOK {
						continue
					}
					body := new(bytes.Buffer)
					body.ReadFrom(answer_resp.Body)
					game.writeLog(fmt.Sprintf("Got answer %v\n", body.String()))
					err = json.NewDecoder(body).Decode(&answer)
					if err != nil {
						panic(err)
					}

					if err := peerConnection.SetRemoteDescription(answer); err != nil {
						panic(err)
					}

					// if we have successfully set the remote description, we can break out of the loop
					ticker.Stop()
					return
				}
			}
		}()
	}
}

func (g *Game) closeConnection() {
	if cErr := peerConnection.Close(); cErr != nil {
		fmt.Printf("cannot close peerConnection: %v\n", cErr)
	}
	// TODO: this doesn't work, fix this
	if g.isHost {
		// delete lobby if host
		url := getSignalingURL() + "/lobby/delete"
		fmt.Println(url)
		httpClient.Get(url)
	}
}

// entry point of the program
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

	// close the connection when the game ends
	g.closeConnection()
}

type Packet struct {
	Pos_x float64
	Pos_y float64
}

// ReadLoop shows how to read from the datachannel directly
func ReadLoop(g *Game, d io.Reader) {
	for {
		buffer := make([]byte, messageSize)
		_, err := io.ReadFull(d, buffer)
		if err != nil {
			g.writeLog(fmt.Sprintln("Datachannel closed; Exit the readloop:", err))
			return
		}

		var packet Packet
		err = binary.Unmarshal(buffer, &packet)
		if err != nil {
			panic(err)
		}

		remote_pos_x = packet.Pos_x
		remote_pos_y = packet.Pos_y

		g.remoteDebugInformation = fmt.Sprintf("Message from DataChannel: %f %f", packet.Pos_x, packet.Pos_y)
	}
}

// WriteLoop shows how to write to the datachannel directly
func WriteLoop(g *Game, d io.Writer) {
	ticker := time.NewTicker(time.Millisecond * 20)
	defer ticker.Stop()
	for range ticker.C {
		packet := &Packet{pos_x, pos_y}
		g.localDebugInformation = fmt.Sprintf("Sending x:%f y:%f", packet.Pos_x, packet.Pos_y)
		encoded, err := binary.Marshal(packet)
		if err != nil {
			panic(err)
		}

		if _, err := d.Write(encoded); err != nil {
			panic(err)
		}
	}
}
