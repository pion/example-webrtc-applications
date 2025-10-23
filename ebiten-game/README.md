## Ebitengine Game!

This is a simple cross-platform game demo built with **Pion** and **Ebitengine**.

You can run one client in the browser and another on the desktop.  
Both clients connect to the same signaling server to communicate in real time.

---

### 🚀 How to Run
To run this demo, you must start the signaling server first, followed by at least two game clients (one web, one desktop, or both the same).

#### 1. Start the signaling server
The signaling server is necessary for establishing WebRTC connections.
```
cd signaling-server
go run . 
```
The signaling server will start on http://localhost:3000

#### 2. Run the Game (Desktop)

```
cd game
go run .
```
This launches the desktop version of the game.

#### 3. Run the Game (Web)
To build and serve the web version:
```
./build_wasm.sh # or ./build_wasm.ps1 on windows
python3 -m http.server 8080
```
Then open your browser and go to:http://localhost:8080

(see [this tutorial for more information on how to build for WebAssembly](https://ebitengine.org/en/documents/webassembly.html))

 You can use any simple static file server.
 
On Windows, building the WASM version requires Git Bash.


## 🕹️ How to Play

#### 1.Connection Setup

* One player clicks “Host Game” to create a lobby.
* Share the Lobby ID with other players.
* Other players enter that Lobby ID to join.

#### 2.Gameplay:

* Use the arrow keys to move around.

Currently, the lobby supports up to two players.

# Notes

Make sure the signaling server is running before starting the game.

The web client and the desktop client must be connected to the same signaling server.

This example is part of the pion/example-webrtc-applications
repository.


*This demo uses [ebitengine/debugui](https://github.com/ebitengine/debugui) for UI elements.*


