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
Check out the **[contributing wiki](https://github.com/pion/webrtc/wiki/Contributing)** to join the group of amazing people making this project possible:

* [John Bradley](https://github.com/kc5nra) - *Original Author*
* [Raphael Randschau](https://github.com/nicolai86) - *STUN*
* [Sean DuBois](https://github.com/Sean-Der) - *Original Author*
* [Michiel De Backker](https://github.com/backkem) - *SDP, Public API, Project Management*
* [Konstantin Itskov](https://github.com/trivigy) - *SDP Parsing*
* [Max Hawkins](https://github.com/maxhawkins) - *RTCP*
* [Justin Okamoto](https://github.com/justinokamoto) - *Fix Docs*
* [leeoxiang](https://github.com/notedit) - *Implement Janus examples*
* [Michael MacDonald](https://github.com/mjmac)
* [Woodrow Douglass](https://github.com/wdouglass) *RTCP, RTP improvements, G.722 support, Bugfixes*
* [Rob Deutsch](https://github.com/rob-deutsch) *RTPReceiver graceful shutdown*
* [Jin Lei](https://github.com/jinleileiking) - *SFU example use http*
* [Antoine Baché](https://github.com/Antonito) - *OGG Opus export*
* [frank](https://github.com/feixiao) - *Building examples on OSX*
* [adwpc](https://github.com/adwpc) - *SFU example with websocket*
* [imalic3](https://github.com/imalic3) - *SFU websocket example with datachannel broadcast*
* [Simonacca Fotokite](https://github.com/simonacca-fotokite)
* [Steve Denman](https://github.com/stevedenman)
* [RunningMan](https://github.com/xsbchen)
* [mchlrhw](https://github.com/mchlrhw)
* [CloudWebRTC|湖北捷智云技术有限公司](https://github.com/cloudwebrtc) - *Flutter example for SFU-WS*
* [Atsushi Watanabe](https://github.com/at-wat) - *WebM muxer example*
* [Jadon Bennett](https://github.com/jsjb)
* [Lukas Herman](https://github.com/lherman-cs) - *C Data Channels example*
* [EricSong](https://github.com/xsephiroth) - *Implement GstV4l2Alsa example*
* [Tristan Matthews](https://github.com/tmatth)
* [Alexey Kravtsov](https://github.com/alexey-kravtsov) - *GStreamer encoder tune*
* [Tarrence van As](https://github.com/tarrencev) - *Webm saver fix*
* [Cameron Elliott](https://github.com/cameronelliott) - *Small race bug fix*
* [Jamie Good](https://github.com/jamiegood) - *Bug fix in jsfiddle example*
* [PhVHoang](https://github.com/PhVHoang)
* [Pascal Benoit](https://github.com/pascal-ace)
* [Jin Gong](https://github.com/cgojin)
* [harkirat singh](https://github.com/hkirat)
* [oasangqi](https://github.com/oasangqi)
* [Shahin Sabooni](https://github.com/longlonghands)

### License
MIT License - see [LICENSE](LICENSE) for full text
