# sfu-ws-turn
sfu-ws-turn is a many-to-many websocket based SFU. This is a more advanced version of [sfu-ws](https://github.com/pion/webrtc/tree/master/examples/sfu-ws)
and demonstrates the following features.

* Trickle ICE
* Re-negotiation
* Basic RTCP
* Multiple inbound/outbound tracks per PeerConnection
* No codec restriction per call. You can have H264 and VP8 in the same conference.
* Support for multiple browsers
* TURN
* Non-local testing

For a production application you should also explore [simulcast](https://github.com/pion/webrtc/tree/master/examples/simulcast),
metrics and robust error handling and check all TODOs for hardening

## Instructions
### Download sfu-ws-turn
This example requires you to clone the repo since it is serving static HTML.

```
mkdir -p $GOPATH/src/github.com/pion
cd $GOPATH/src/github.com/pion
git clone https://github.com/pion/example-webrtc-applications.git
cd webrtc/examples/sfu-ws-turn
```

### Create certificates
```./prep-certs```
This creates a self signed CA and a server certificate for the turn server, and the web servers.

### Fix certificate path in coturn-user-management
Tweak ./coturn-use-management/turnserver.conf as required

### Install coturn
```sudo apt install -y coturn```
This is the TURN server, for use with firewalls

### Start the turn server
```sudo turnserver -c ./coturn-user-management/turnserver.conf --daemon```

### Add relevant lines to hosts files
This will need done on all machines you want to test from
```sudo vim /etc/hosts```
and add the following, changing 127.0.0.1 to the IP of your server
```127.0.0.1 example.com
127.0.0.1 turn.example.com
```


### Run the credential server
```go build -o ./coturn-user-management/ ./coturn-user-management/main.go  && sudo ./coturn-user-management/main -cert /tmp/sfu-ws-turn-certs/server.crt -cert-key /tmp/sfu-ws-turn-certs/server.key
```
sudo is required so it has write access to the coturn db.
This enables dynamic credentials for the TURN server.

### Run sfu-ws-turn
Execute `go run *.go` with TODO flags
```go build main.go  && sudo ./main -cert /tmp/sfu-ws-turn-certs/server.crt -cert-key /tmp/sfu-ws-turn-certs/server.key -cred-URL https://example.com:8443/20987182471824882098  -insecure-reqs true
```

### Open the Web UI
Open [https://example.com](https://example.com). This will automatically connect and send your video. Now join from other tabs and browsers!

Congratulations, you have used Pion WebRTC! Now start building something cool
