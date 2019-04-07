<h1 align="center">
  Examples WebRTC Applications
</h1>

The following are a collection of example applications built by Pion users.  These applications show real world usage of Pion,
and should serve as a good starting point for your next project. For more minimal examples check out [examples](https://github.com/pion/webrtc/tree/master/examples) in the Pion WebRTC repository

If you have a request please make an issue, we also love contributions more examples are always welcome.

Have any questions? Join [the Slack channel](https://pion.ly/slack) to follow development and speak with the maintainers.

## Examples
* [Gstreamer Receive](gstreamer-receive): The gstreamer-receive example shows how to receive media from the browser and play it live. This example uses GStreamer for rendering.
* [Gstreamer Send](gstreamer-send): Example gstreamer-send shows how to send video to your browser. This example uses GStreamer to process the video.
* [Gstreamer Send Offer](gstreamer-send-offer): Example gstreamer-send-offer is a variant of gstreamer-send that initiates the WebRTC connection by sending an offer.
* [Janus Gateway](janus-gateway): Example janus-gateway is a collection of examples showing how to use Pion WebRTC with [janus-gateway](https://github.com/meetecho/janus-gateway).
* [SFU Websocket](sfu-ws): The SFU example demonstrates how to broadcast a video to multiple peers. A broadcaster uploads the video once and the server forwards it to all other peers.


### Usage
We've made it easy to run the browser based examples on your local machine.

1. Build and run the example server:
    ``` sh
    go get github.com/webrtc-example-applications
    cd $GOPATH/src/github.com/webrtc-example-applications
    go run examples.go
    ```

2. Browse to [localhost](http://localhost) to browse through the examples.

Note that you can change the port of the server using the ``--address`` flag.

### Contributing
Check out the **[contributing wiki](https://github.com/pion/webrtc/wiki/Contributing)** to join the group of amazing people making this project possible:

### License
MIT License - see [LICENSE](LICENSE) for full text
