# Ebitengine Game!

This is a pretty nifty demo on how to use [ebitengine](https://ebitengine.org/) and [pion](https://github.com/pion/webrtc) to pull off a cross platform game!

You can have a client running on the browser and one running on a desktop and they can talk to each other, provided they are connected to the same signaling server

Requires the signaling server to be running. To do so, just go inside the folder /signaling-server and do ``go run .``

you can then run the game by going in /game and doing either

``go run .`` for running the game on desktop

(see [this tutorial for more information on how to build for WebAssembly](https://ebitengine.org/en/documents/webassembly.html))

Click "Host Game" to get the lobby id, and then share that with the other clients to get connected

To play: Just move around with the arrow keys once you have connected!

Right now this only supports two clients in the same lobby
