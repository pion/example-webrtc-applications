# sip-to-webrtc
SIP is an example of accepting inbounding SIP traffic (Invites) and bridging it with WebRTC.
This is the most common way to connect phone calls with your WebRTC application.
This is possible because of the excellent [emiago/sipgo](https://github.com/emiago/sipgo) library.

This example demonstrates accepting SIP audio and playing it in the browser. If you wish to implement multiple participants
in a call you will need to have audio mixing. See [zaf/g711](https://github.com/zaf/g711) for decoding + encoding. After you have
decoded sum all the samples and then encode it.

## Instructions
### Open sip-to-webrtc example page
[jsfiddle.net](https://jsfiddle.net/gds05mc3/) you should see a audio player, two text-areas and a 'Start Session' button

## Run sip-to-webrtc with your browsers SessionDescription as stdin
In the jsfiddle the top textarea is your browser, copy that and:
#### Linux/macOS
Run `echo $BROWSER_SDP | sip-to-webrtc`
#### Windows
1. Paste the SessionDescription into a file.
1. Run `sip-to-webrtc < my_file`

### Input sip-to-webrtc's SessionDescription into your browser
Copy the text that `sip-to-webrtc` just emitted and copy into second text area

### Hit 'Start Session' to connect
If a WebRTC session was successfully established you will get log messages
about ICEConnectionState going to `connected`. In your browser and terminal.

**Browser**
```
checking
connected
```

**Terminal**
```
Connection State has changed checking
Connection State has changed connected

Starting SIP Listener
```

If everything worked now it is time to make a SIP Invite.

### Make your phone call
`sip-to-webrtc` is now listening on `:5060` and will accept all invites.
When an Invite has been accepted you will see a log message like this.

```
Accepting SIP Invite: From: "+15550100001" <sip:+15550100001@192.168.1.93>;tag=nc8uzmZUHUTbqH0v
```

### Done
You should hear the audio of the phone call in your browser.

Congrats, you have used Pion WebRTC! Now start building something cool
