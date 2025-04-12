package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	//"runtime"

	//"github.com/pion/randutil"
	"image/color"
	_ "image/png"
	"io"
	"log"
	"os"
	"time"

	"github.com/pion/webrtc/v4"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/kelindar/binary"

	//"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/gofont/goregular"
)

var img *ebiten.Image

var (
	pos_x        = 40.0
	pos_y        = 40.0
	remote_pos_x = 40.0
	remote_pos_y = 40.0
)

var lobby_id string
var isHost = false

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
	ui         *ebitenui.UI
	hostButton *widget.Button
	joinButton *widget.Button
	//This parameter is so you can keep track of the textInput widget to update and retrieve
	//its values in other parts of your game
	standardTextInput *widget.TextInput
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

	// update the UI
	g.ui.Update()

	// if update returns non nil error, game suspends
	return nil
}

// called every frame, depends on the monitor refresh rate
// which will probably be at least 60 times per second
func (g *Game) Draw(screen *ebiten.Image) {
	// draw the UI onto the screen
	g.ui.Draw(screen)

	// prints something on the screen
	ebitenutil.DebugPrint(screen, fmt.Sprintf("FPS: %f", ebiten.ActualFPS()))

	// draw image
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(pos_x, pos_y)
	screen.DrawImage(img, op)

	// draw remote
	op2 := &ebiten.DrawImageOptions{}
	op2.GeoM.Translate(remote_pos_x, remote_pos_y)
	screen.DrawImage(img, op2)
}

var (
	// probably move all webrtc networking stuff to a struct i can manage
	peerConnection *webrtc.PeerConnection
)

const messageSize = 32

type PlayerData struct {
	Id int
}

func startConnection(game *Game) {
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
		fmt.Printf("Peer Connection State has changed: %s\n", s.String())

		if s == webrtc.PeerConnectionStateFailed {
			// Wait until PeerConnection has had no network activity for 30 seconds or another failure. It may be reconnected using an ICE Restart.
			// Use webrtc.PeerConnectionStateDisconnected if you are interested in detecting faster timeout.
			// Note that the PeerConnection may come back from PeerConnectionStateDisconnected.
			fmt.Println("Peer Connection has gone to failed exiting")
			os.Exit(0)
		}

		if s == webrtc.PeerConnectionStateClosed {
			// PeerConnection was explicitly closed. This usually happens from a DTLS CloseNotify
			fmt.Println("Peer Connection has gone to closed exiting")
			os.Exit(0)
		}
	})

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("ICE Connection State has changed: %s\n", connectionState.String())
	})

	// the one that gives the answer is the host
	if isHost {

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
		fmt.Printf("Lobby ID: %s\n", lobby_id)
		game.standardTextInput.SetText(lobby_id)

		// Register data channel creation handling
		peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {
			fmt.Printf("New DataChannel %s %d\n", d.Label(), d.ID())

			// Register channel opening handling
			d.OnOpen(func() {
				fmt.Printf("Data channel '%s'-'%d' open.\n", d.Label(), d.ID())

				// Detach the data channel
				raw, dErr := d.Detach()
				if dErr != nil {
					panic(dErr)
				}

				// Handle reading from the data channel
				go ReadLoop(raw)

				// Handle writing to the data channel
				go WriteLoop(raw)
			})
		})

		// poll for offer from signaling server for player
		pollForPlayerOffer := func(player_id int) {
			ticker := time.NewTicker(1 * time.Second)
			for {
				select {
				case t := <-ticker.C:
					fmt.Println("Tick at", t)
					fmt.Printf("Polling for offer for %d\n", player_id)
					// hardcode that there is only one other player and they have player_id 1
					getUrl := getSignalingURL() + "/offer/get?lobby_id=" + lobby_id + "&player_id=" + strconv.Itoa(player_id)
					fmt.Println(getUrl)
					offer_resp, err := httpClient.Get(getUrl)
					if err != nil {
						panic(err)
					}
					if offer_resp.StatusCode != http.StatusOK {
						continue
					}
					body := new(bytes.Buffer)
					body.ReadFrom(offer_resp.Body)
					fmt.Printf("Got offer %v\n", body.String())
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
					fmt.Println(postUrl)
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
					fmt.Println("Tick at", t)
					idUrl := getSignalingURL() + "/lobby/unregisteredPlayers?id=" + lobby_id
					fmt.Println(idUrl)
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
					fmt.Printf("Player IDs: %v\n", player_ids)
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
		// the following is for the client joining the lobby
		// get lobby id from text input
		lobby_id = game.standardTextInput.GetText()
		response, err := httpClient.Get(getSignalingURL() + "/lobby/join?id=" + lobby_id)
		if err != nil {
			panic(err)
		}
		var player_data PlayerData
		err = json.NewDecoder(response.Body).Decode(&player_data)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Player ID: %v\n", player_data)
		// Create a datachannel with label 'data'
		dataChannel, err := peerConnection.CreateDataChannel("data", nil)
		if err != nil {
			panic(err)
		}

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
				fmt.Println(postUrl)
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
					fmt.Println("Tick at", t)
					fmt.Println("Polling for answer")
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
					fmt.Printf("Got answer %v\n", body.String())
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

func closeConnection() {
	if cErr := peerConnection.Close(); cErr != nil {
		fmt.Printf("cannot close peerConnection: %v\n", cErr)
	}
	// TODO: this doesn't work, fix this
	if isHost {
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

	// load images for button states: idle, hover, and pressed
	buttonImage, _ := loadButtonImage()

	// load button text font
	face, _ := loadFont(20)

	// construct a new container that serves as the root of the UI hierarchy
	rootContainer := widget.NewContainer(
		// the container will use a plain color as its background
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{0x13, 0x1a, 0x22, 0xff})),

		// the container will use an anchor layout to layout its single child widget
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)
	game := Game{}
	// construct the UI
	game.ui = &ebitenui.UI{
		Container: rootContainer,
	}

	// Creating button variable first so that it is usable in callbacks
	game.hostButton = widget.NewButton(
		// set general widget options
		widget.ButtonOpts.WidgetOpts(
			// instruct the container's anchor layout to center the button both horizontally and vertically
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionStart,
			}),
		),
		// specify the images to use
		widget.ButtonOpts.Image(buttonImage),

		// specify the button's text, the font face, and the color
		//widget.ButtonOpts.Text("Hello, World!", face, &widget.ButtonTextColor{
		widget.ButtonOpts.Text("Host Game", face, &widget.ButtonTextColor{
			Idle:    color.NRGBA{0xdf, 0xf4, 0xff, 0xff},
			Hover:   color.NRGBA{0, 255, 128, 255},
			Pressed: color.NRGBA{255, 0, 0, 255},
		}),
		widget.ButtonOpts.TextProcessBBCode(true),
		// specify that the button's text needs some padding for correct display
		widget.ButtonOpts.TextPadding(widget.Insets{
			Left:   30,
			Right:  30,
			Top:    5,
			Bottom: 5,
		}),

		// add a handler that reacts to clicking the button
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			fmt.Println(game.standardTextInput.GetText())
			isHost = true
			startConnection(&game)
		}),

		// Indicate that this button should not be submitted when enter or space are pressed
		widget.ButtonOpts.DisableDefaultKeys(),
	)

	// add the button as a child of the container
	rootContainer.AddChild(game.hostButton)

	// Creating button variable first so that it is usable in callbacks
	game.joinButton = widget.NewButton(
		// set general widget options
		widget.ButtonOpts.WidgetOpts(
			// instruct the container's anchor layout to center the button both horizontally and vertically
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionEnd,
			}),
		),
		// specify the images to use
		widget.ButtonOpts.Image(buttonImage),

		// specify the button's text, the font face, and the color
		//widget.ButtonOpts.Text("Hello, World!", face, &widget.ButtonTextColor{
		widget.ButtonOpts.Text("Join Lobby", face, &widget.ButtonTextColor{
			Idle:    color.NRGBA{0xdf, 0xf4, 0xff, 0xff},
			Hover:   color.NRGBA{0, 255, 128, 255},
			Pressed: color.NRGBA{255, 0, 0, 255},
		}),
		widget.ButtonOpts.TextProcessBBCode(true),
		// specify that the button's text needs some padding for correct display
		widget.ButtonOpts.TextPadding(widget.Insets{
			Left:   30,
			Right:  30,
			Top:    5,
			Bottom: 5,
		}),

		// add a handler that reacts to clicking the button
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			fmt.Println(game.standardTextInput.GetText())
			isHost = false
			startConnection(&game)
		}),

		// Indicate that this button should not be submitted when enter or space are pressed
		widget.ButtonOpts.DisableDefaultKeys(),
	)

	// add the button as a child of the container
	rootContainer.AddChild(game.joinButton)

	// construct a new container to contain textboxes
	textBoxContainer := widget.NewContainer(
		// the container will use a plain color as its background
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{0x13, 0x1a, 0x22, 0xff})),

		// the container will use an anchor layout to layout its single child widget
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),

		widget.ContainerOpts.WidgetOpts(
			//Set the layout information to center the container in the parent
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
			widget.WidgetOpts.MinSize(150, 150),
		),
	)

	rootContainer.AddChild(textBoxContainer)

	// construct a standard textinput widget for signaling server ip
	signalingTextInput := widget.NewTextInput(
		widget.TextInputOpts.WidgetOpts(
			//Set the layout information to center the textbox in the parent
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionStart,
			}),
			widget.WidgetOpts.MinSize(150, 30),
		),

		//Set the Idle and Disabled background image for the text input
		//If the NineSlice image has a minimum size, the widget will use that or
		// widget.WidgetOpts.MinSize; whichever is greater
		widget.TextInputOpts.Image(&widget.TextInputImage{
			Idle:     image.NewNineSliceColor(color.NRGBA{R: 100, G: 100, B: 100, A: 255}),
			Disabled: image.NewNineSliceColor(color.NRGBA{R: 100, G: 100, B: 100, A: 255}),
		}),

		//Set the font face and size for the widget
		widget.TextInputOpts.Face(face),

		//Set the colors for the text and caret
		widget.TextInputOpts.Color(&widget.TextInputColor{
			Idle:          color.NRGBA{254, 255, 255, 255},
			Disabled:      color.NRGBA{R: 200, G: 200, B: 200, A: 255},
			Caret:         color.NRGBA{254, 255, 255, 255},
			DisabledCaret: color.NRGBA{R: 200, G: 200, B: 200, A: 255},
		}),

		//Set how much padding there is between the edge of the input and the text
		widget.TextInputOpts.Padding(widget.NewInsetsSimple(5)),

		//Set the font and width of the caret
		widget.TextInputOpts.CaretOpts(
			widget.CaretOpts.Size(face, 2),
		),

		//This text is displayed if the input is empty
		widget.TextInputOpts.Placeholder("Signaling Server IP"),

		//This is called whenever there is a change to the text
		widget.TextInputOpts.ChangedHandler(func(args *widget.TextInputChangedEventArgs) {
			fmt.Println("Text Changed: ", args.InputText)
			signalingIP = args.InputText
		}),
	)

	signalingTextInput.SetText(signalingIP)

	textBoxContainer.AddChild(signalingTextInput)

	// construct a standard textinput widget for lobby id
	game.standardTextInput = widget.NewTextInput(
		widget.TextInputOpts.WidgetOpts(
			//Set the layout information to center the textbox in the parent
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionEnd,
			}),
			widget.WidgetOpts.MinSize(150, 30),
		),

		//Set the Idle and Disabled background image for the text input
		//If the NineSlice image has a minimum size, the widget will use that or
		// widget.WidgetOpts.MinSize; whichever is greater
		widget.TextInputOpts.Image(&widget.TextInputImage{
			Idle:     image.NewNineSliceColor(color.NRGBA{R: 100, G: 100, B: 100, A: 255}),
			Disabled: image.NewNineSliceColor(color.NRGBA{R: 100, G: 100, B: 100, A: 255}),
		}),

		//Set the font face and size for the widget
		widget.TextInputOpts.Face(face),

		//Set the colors for the text and caret
		widget.TextInputOpts.Color(&widget.TextInputColor{
			Idle:          color.NRGBA{254, 255, 255, 255},
			Disabled:      color.NRGBA{R: 200, G: 200, B: 200, A: 255},
			Caret:         color.NRGBA{254, 255, 255, 255},
			DisabledCaret: color.NRGBA{R: 200, G: 200, B: 200, A: 255},
		}),

		//Set how much padding there is between the edge of the input and the text
		widget.TextInputOpts.Padding(widget.NewInsetsSimple(5)),

		//Set the font and width of the caret
		widget.TextInputOpts.CaretOpts(
			widget.CaretOpts.Size(face, 2),
		),

		//This text is displayed if the input is empty
		widget.TextInputOpts.Placeholder("Lobby ID"),

		//This is called when the user hits the "Enter" key.
		//There are other options that can configure this behavior
		widget.TextInputOpts.SubmitHandler(func(args *widget.TextInputChangedEventArgs) {
			fmt.Println("Text Submitted: ", args.InputText)
		}),

		//This is called whenever there is a change to the text
		widget.TextInputOpts.ChangedHandler(func(args *widget.TextInputChangedEventArgs) {
			fmt.Println("Text Changed: ", args.InputText)
		}),
	)

	textBoxContainer.AddChild(game.standardTextInput)

	// triggers the game loop to actually start up
	// if we run into an error, log what it is
	if err := ebiten.RunGame(&game); err != nil {
		log.Fatal(err)
	}

	// close the connection when the game ends
	closeConnection()
}

type Packet struct {
	Pos_x float64
	Pos_y float64
}

// ReadLoop shows how to read from the datachannel directly
func ReadLoop(d io.Reader) {
	for {
		buffer := make([]byte, messageSize)
		_, err := io.ReadFull(d, buffer)
		if err != nil {
			fmt.Println("Datachannel closed; Exit the readloop:", err)
			return
		}

		var packet Packet
		err = binary.Unmarshal(buffer, &packet)
		if err != nil {
			panic(err)
		}

		remote_pos_x = packet.Pos_x
		remote_pos_y = packet.Pos_y

		fmt.Printf("Message from DataChannel: %f %f\n", packet.Pos_x, packet.Pos_y)
	}
}

// WriteLoop shows how to write to the datachannel directly
func WriteLoop(d io.Writer) {
	ticker := time.NewTicker(time.Millisecond * 20)
	defer ticker.Stop()
	for range ticker.C {
		packet := &Packet{pos_x, pos_y}
		fmt.Printf("Sending x:%f y:%f\n", packet.Pos_x, packet.Pos_y)
		encoded, err := binary.Marshal(packet)
		if err != nil {
			panic(err)
		}

		if _, err := d.Write(encoded); err != nil {
			panic(err)
		}
	}
}

func loadButtonImage() (*widget.ButtonImage, error) {
	idle := image.NewNineSliceColor(color.NRGBA{R: 170, G: 170, B: 180, A: 255})

	hover := image.NewNineSliceColor(color.NRGBA{R: 130, G: 130, B: 150, A: 255})

	pressed := image.NewNineSliceColor(color.NRGBA{R: 100, G: 100, B: 120, A: 255})

	return &widget.ButtonImage{
		Idle:    idle,
		Hover:   hover,
		Pressed: pressed,
	}, nil
}

func loadFont(size float64) (text.Face, error) {
	s, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return &text.GoTextFace{
		Source: s,
		Size:   size,
	}, nil
}
