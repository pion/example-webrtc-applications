# ffmpeg-receive
ffmpeg-receive is a simple application that shows how to receive video using Pion WebRTC and decode it using FFmpeg.

## Instructions
### Install FFmpeg
This example requires you have FFmpeg installed. See the docs of [go-astiav](https://github.com/asticode/go-astiav) for
how to do this.

### Download ffmpeg-receive
```
export GO111MODULE=on
go install github.com/pion/example-webrtc-applications/v3/ffmpeg-receive@latest
```

### Open ffmpeg-receive test page
[jsfiddle.net](https://jsfiddle.net/07jknmed/) you should see your Webcam, two text-areas and a 'Start Session' button

### Run ffmpeg-receive with your browsers SessionDescription as stdin
In the jsfiddle the top textarea is your browser, copy that and:
#### Linux/macOS
Run `echo $BROWSER_SDP | ffmpeg-receive`
#### Windows
1. Paste the SessionDescription into a file.
1. Run `ffmpeg-receive < my_file`

### Input ffmpeg-receive's SessionDescription into your browser
Copy the text that `ffmpeg-receive` just emitted and copy into second text area

### Hit 'Start Session' in jsfiddle
ffmpeg-receive will decode incoming video and print frame metadata until you close the application.

Output will look like this:
```
Decoded frame: pts=36 width=640 height=480 pixel_format=yuv420p picture_type=P sample_aspect_ratio=0
Decoded frame: pts=37 width=640 height=480 pixel_format=yuv420p picture_type=P sample_aspect_ratio=0
Decoded frame: pts=38 width=640 height=480 pixel_format=yuv420p picture_type=P sample_aspect_ratio=0
Decoded frame: pts=39 width=640 height=480 pixel_format=yuv420p picture_type=P sample_aspect_ratio=0
Decoded frame: pts=40 width=640 height=480 pixel_format=yuv420p picture_type=P sample_aspect_ratio=0
Decoded frame: pts=41 width=640 height=480 pixel_format=yuv420p picture_type=P sample_aspect_ratio=0
Decoded frame: pts=42 width=640 height=480 pixel_format=yuv420p picture_type=P sample_aspect_ratio=0
Decoded frame: pts=43 width=640 height=480 pixel_format=yuv420p picture_type=P sample_aspect_ratio=0
Decoded frame: pts=44 width=640 height=480 pixel_format=yuv420p picture_type=P sample_aspect_ratio=0
Decoded frame: pts=45 width=640 height=480 pixel_format=yuv420p picture_type=P sample_aspect_ratio=0
Decoded frame: pts=46 width=640 height=480 pixel_format=yuv420p picture_type=P sample_aspect_ratio=0
Decoded frame: pts=47 width=640 height=480 pixel_format=yuv420p picture_type=P sample_aspect_ratio=0
Decoded frame: pts=48 width=640 height=480 pixel_format=yuv420p picture_type=P sample_aspect_ratio=0
```

Congrats, you have used Pion WebRTC! Now start building something cool
