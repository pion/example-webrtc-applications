# rtmp-to-webrtc
rtmp-to-webrtc demonstrates how you could re-stream media from a RTMP server to WebRTC.
This example was heavily inspired by [rtp-to-webrtc](https://github.com/pion/webrtc/tree/master/examples/rtp-to-webrtc)

This example re-encodes to VP8. Pion WebRTC supports H264, but browser support is inconsistent. To switch video codecs replace all occurrences
of VP8 with H264 in `main.go`

## Instructions
### Download rtmp-to-webrtc
```
export GO111MODULE=on
go get github.com/pion/example-webrtc-applications/v3/rtmp-to-webrtc
```

### Open jsfiddle example page
[jsfiddle.net](https://jsfiddle.net/z7ms3u5r/) you should see two text-areas and a 'Start Session' button


### Run rtmp-to-webrtc with your browsers SessionDescription as stdin
In the jsfiddle the top textarea is your browser's SessionDescription, copy that and:

#### Linux/macOS
Run `echo $BROWSER_SDP | rtmp-to-webrtc`

#### Windows
1. Paste the SessionDescription into a file.
1. Run `rtmp-to-webrtc < my_file`

### Send RTP to listening socket
On startup you will get a message `Waiting for RTP Packets`, you can use any software to send VP8 packets to port 5004 and Opus packets to port 5006. We have an example using ffmpeg below

#### ffmpeg
```
ffmpeg -i '$RTMP_URL' -an -vcodec libvpx -cpu-used 5 -deadline 1 -g 10 -error-resilient 1 -auto-alt-ref 1 -f rtp rtp://127.0.0.1:5004?pkt_size=1200 -vn -c:a libopus -f rtp rtp:/127.0.0.1:5006?pkt_size=1200
```

### Input rtmp-to-webrtc's SessionDescription into your browser
Copy the text that `rtmp-to-webrtc` just emitted and copy into second text area

### Hit 'Start Session' in jsfiddle, enjoy your video!
A video should start playing in your browser above the input boxes.

Congrats, you have used Pion WebRTC! Now start building something cool
