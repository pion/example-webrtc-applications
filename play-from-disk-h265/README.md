# play-from-disk-h265
play-from-disk-h265 demonstrates how to send video and/or audio to your browser from H265 files saved to disk.

This example has the same structure as [play-from-disk](https://github.com/pion/webrtc/tree/master/examples/play-from-disk) but instead of sending VP8 it sends H265 instead.

## Instructions
### Create H265 Annex-B file named `output.h265` and/or `output.ogg` that contains a Opus track
```
ffmpeg -i $INPUT_FILE -an -c:v libx265 -bsf:v hevc_mp4toannexb -b:v 2M -max_delay 0 -bf 0 output.h265
ffmpeg -i $INPUT_FILE -c:a libopus -page_duration 20000 -vn -ac 2 output.ogg
```

### Download play-from-disk-h265
```
export GO111MODULE=on
go install github.com/pion/example-webrtc-applications/v3/play-from-disk-h265@latest
```

### Open play-from-disk-h265 example page
[jsfiddle.net](https://jsfiddle.net/8qvzh6ue/) you should see two text-areas and a 'Start Session' button

### Run play-from-disk-h265 with your browsers SessionDescription as stdin
The `output.h265` and `output.ogg` you created should be in the same directory as `play-from-disk-h265`. In the jsfiddle the top textarea is your browser, copy that and:

#### Linux/macOS
Run `echo $BROWSER_SDP | play-from-disk-h265`

#### Windows
1. Paste the SessionDescription into a file.
1. Run `play-from-disk-h265 < my_file`

### Input play-from-disk-h265's SessionDescription into your browser
Copy the text that `play-from-disk-h265` just emitted and copy into second text area

### Hit 'Start Session' in jsfiddle, enjoy your video!
A video should start playing in your browser above the input boxes. `play-from-disk-h265` will exit when the file reaches the end

Congrats, you have used Pion WebRTC! Now start building something cool
