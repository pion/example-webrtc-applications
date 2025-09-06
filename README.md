<h1 align="center">
  Examples WebRTC Applications
</h1>

The following are a collection of example applications built by Pion users.  These applications show real world usage of Pion,
and should serve as a good starting point for your next project. For more minimal examples check out [examples](https://github.com/pion/webrtc/tree/master/examples) in the Pion WebRTC repository

If you have a request please make an issue, we also love contributions more examples are always welcome.

Have any questions? Join [the Slack channel](https://pion.ly/slack) to follow development and speak with the maintainers.

## Examples
* [GoCV Receive](gocv-receive): Example gocv-receive shows how to receive media using Pion and then do motion detection with GoCV.
* [Gstreamer Receive](gstreamer-receive): Example gstreamer-receive shows how to receive media from the browser and play it live. This example uses GStreamer for rendering.
* [Gstreamer Send](gstreamer-send): Example gstreamer-send shows how to send video to your browser. This example uses GStreamer to process the video.
* [Gstreamer Send Offer](gstreamer-send-offer): Example gstreamer-send-offer is a variant of gstreamer-send that initiates the WebRTC connection by sending an offer.
* [Janus Gateway](janus-gateway): Example janus-gateway is a collection of examples showing how to use Pion WebRTC with [janus-gateway](https://github.com/meetecho/janus-gateway).
* [SFU Websocket](sfu-ws): The SFU example demonstrates a conference system that uses WebSocket for signaling. It also includes a flutter client for Android, iOS and Native.
* [Save to WebM](save-to-webm): Example save-to-webm shows how to receive audio and video using Pion and then save to WebM container.
* [Twitch](twitch): Example twitch shows how to send audio/video from WebRTC to https://www.twitch.tv/ via RTMP.
* [C DataChannels](c-data-channels) Example c-data-channels shows how you can use Pion WebRTC from a C program
* [Snapshot](snapshot) Example snapshot shows how you can convert incoming video frames to jpeg and serve them via HTTP.
* [SIP to WebRTC](sip-to-webrtc) Example sip-to-webrtc shows how to bridge WebRTC and SIP traffic.
* [GoCV to WebRTC](gocv-to-webrtc): Example gocv-to-webrtc captures webcam and performs motion detection in Go, it then sends results to view in the browser.


### Usage
We've made it easy to run the browser based examples on your local machine.

1. Build and run the example server:
    ``` sh
    go get github.com/pion/example-webrtc-applications
    cd $GOPATH/src/github.com/pion/example-webrtc-applications
    go run examples.go
    ```

2. Browse to [localhost](http://localhost) to browse through the examples.

Note that you can change the port of the server using the ``--address`` flag.

### Contributing
Check out the **[contributing wiki](https://github.com/pion/webrtc/wiki/Contributing)** to join the group of amazing people making this project possible

### License
MIT License - see [LICENSE](LICENSE) for full text
