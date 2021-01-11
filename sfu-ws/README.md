# sfu-ws

sfu-ws is a many-to-many websocket based SFU. This is a more advanced version of [broadcast](https://github.com/pion/webrtc/tree/master/examples/broadcast)
and demonstrates the following features.

* Trickle ICE
* Re-negotiation
* Basic RTCP
* Multiple inbound/outbound tracks per PeerConnection
* No codec restriction per call. You can have H264 and VP8 in the same conference.
* Support for multiple browsers

We also provide a flutter client that supports the following platforms

* Android, iOS
* Web
* MacOS (Windows, Linux and Fuschia in the [future](https://github.com/flutter-webrtc/flutter-webrtc#functionality))

For a production application you should also explore [simulcast](https://github.com/pion/webrtc/tree/master/examples/simulcast),
metrics and robust error handling.

## Instructions

### Download sfu-ws

This example requires you to clone the repo since it is serving static HTML.

```sh
mkdir -p $GOPATH/src/github.com/pion
cd $GOPATH/src/github.com/pion
git clone https://github.com/pion/example-webrtc-applications.git
cd example-webrtc-applications/sfu-ws
```

### Run sfu-ws

```sh
# Run sfu-ws
go run *.go

# Run sfu-ws with logs
PION_LOG_INFO=sfu-ws go run *.go

# Run sfu-ws with all logs
PION_LOG_TRACE=all go run *.go
```

### Open the Web UI

Open [http://localhost:8080](http://localhost:8080). This will automatically connect and send your video. Now join from other tabs and browsers!

Congrats, you have used Pion WebRTC! Now start building something cool
