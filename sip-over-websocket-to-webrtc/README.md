# sip-over-websocket-to-webrtc
SIP Signaling via WebSocket is defined in [RFC 7118](https://www.rfc-editor.org/rfc/rfc7118.html).
If you want to connect to a SIP server via UDP/TCP see [sip-to-webrtc](../sip-to-webrtc)

sip-over-websocket-to-webrtc demonstrates how to connect to a SIP Server via Websocket.
This example connects to an extension and saves the audio to a ogg file.

## Instructions
### Setup FreeSWITCH (or SIP over WebSocket Server)
With a fresh install of FreeSWITCH all you need to do is

* Enable `ws-binding`
* Set a `default_password` to something you know

### Run `sip-to-webrtc`
Run `go run *.go -h` to see the arguments of the program. If everything is working
this is the output you will see.

```
$ go run *.go -host 172.17.0.2 -password Aelo1ievoh2oopooTh2paijaeNaidiek
  Connection State has changed checking
  Connection State has changed connected
  Got Opus track, saving to disk as output.ogg
  Connection State has changed disconnected
```

### Play the audio file
ffmpeg's in-tree Opus decoder isn't able to play the default audio file from FreeSWITCH. Use the following command to force libopus.

`ffplay -acodec libopus output.ogg`
