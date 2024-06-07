# ffmpeg-send
ffmpeg-send is a simple application that shows how to send video to your browser using Pion WebRTC and FFmpeg.

## Instructions
### Install FFmpeg
This example requires you have FFmpeg installed. See the docs of [go-astiav](https://github.com/asticode/go-astiav) for
how to do this.

### Download ffmpeg-send
```
export GO111MODULE=on
go install github.com/pion/example-webrtc-applications/v3/ffmpeg-send@latest
```

### Open ffmpeg-send example page
[jsfiddle.net](https://jsfiddle.net/z17q28cd/) you should see two text-areas and a 'Start Session' button

### Run ffmpeg-send with your browsers SessionDescription as stdin
In the jsfiddle the top textarea is your browser, copy that and:
#### Linux/macOS
Run `echo $BROWSER_SDP | ffmpeg-send`
#### Windows
1. Paste the SessionDescription into a file.
1. Run `ffmpeg-send < my_file`

### Input ffmpeg-send's SessionDescription into your browser
Copy the text that `ffmpeg-send` just emitted and copy into second text area

### Hit 'Start Session' in jsfiddle, enjoy your video!
A video should start playing in your browser above the input boxes, and will continue playing until you close the application.

Congrats, you have used Pion WebRTC! Now start building something cool
