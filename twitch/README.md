# This example is deprecated

This example is hard to debug/opaque in how it works so we have deprecated it.
A better alternative is [rtp-forwarder](https://github.com/pion/webrtc/tree/master/examples/rtp-forwarder#twitchrtmp).
`rtp-forwarder` shows how to forward WebRTC into ffmpeg. When you have the media inside
ffmpeg it is much easier to process and send to Twitch by invoking the process yourself.

------

# twitch
Twitch demonstrates how to capture your webcam/microphone via WebRTC and send to Twitch.

## Instructions
### Install ffmpeg
This example requires you have ffmpeg installed, these are the supported platforms
#### Debian/Ubuntu
`sudo apt-get install ffmpeg`
#### Windows MinGW64/MSYS2
`pacman -S ffmpeg`
#### macOS
` brew install ffmpeg`

### Download twitch
```
export GO111MODULE=on
go get github.com/pion/example-webrtc-applications/v3/twitch
```

### Open twitch example page
[jsfiddle.net](https://jsfiddle.net/cqavdpj8/1/) you should see your Webcam, two text-areas and a 'Start Session' button

### Run twitch with your browsers SessionDescription as stdin and stream-key as an argument
In the jsfiddle the top textarea is your browser, copy that and:
#### Linux/macOS
Run `echo $BROWSER_SDP | twitch $STREAM_KEY`
#### Windows
1. Paste the SessionDescription into a file.
1. Run `twitch $STREAM_KEY < my_file`

### Input twitch's SessionDescription into your browser
Copy the text that `twitch` just emitted and copy into second text area

### Hit 'Start Session' in jsfiddle, enjoy your media!
The output from `ffmpeg` will be printed to your console, and if your stream-key is correct you will see it on Twitch soon!

Congrats, you have used Pion WebRTC! Now start building something cool
