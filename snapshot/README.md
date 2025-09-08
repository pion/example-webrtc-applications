# snapshot
snapshot shows how you can convert incoming video frames to jpeg and serve them via HTTP.

## Instructions

### Download snapshot
This example requires you to clone the repo since it is serving static HTML.

```
mkdir -p $GOPATH/src/github.com/pion
cd $GOPATH/src/github.com/pion
git clone https://github.com/pion/example-webrtc-applications.git
cd example-webrtc-applications/snapshot
```

### Run snapshot
Execute `go run *.go`

### Open the Web UI
Open [http://localhost:8080](http://localhost:8080) to publish your video and generate snapshots. You can open this page twice to generate snapshots and publish from different tabs, or you can do it in the same tab.

First press `Publish Video` to start pushing video to your the Pion WebRTC backend. When you are ready to generate a snapshot press `Generate Snapshot`

You can access the snapshot generator directly at [http://localhost:8080/snapshot](http://localhost:8080/snapshot)

Congrats, you have used Pion WebRTC! Now start building something cool
