# play-from-disk-mkv
play-from-disk-mkv demonstrates how to send video and/or audio to your browser from a MKV file

This example has the same structure as [play-from-disk](https://github.com/pion/webrtc/tree/master/examples/play-from-disk) but instead reads from a MKV file

## Instructions
### Create a MKV with a H264 + Opus track
```
ffmpeg -i $INPUT_FILE -c:v libx264 -b:v 2M -max_delay 0 -bf 0 -g 30 -c:a libopus -page_duration 20000 output.mkv
```

### Download play-from-disk-mkv
```
go install github.com/pion/example-webrtc-applications/v3/play-from-disk-mkv@latest
```

### Open play-from-disk-mkv example page
[jsfiddle.net](https://jsfiddle.net/8qvzh6ue/) you should see two text-areas and a 'Start Session' button

### Run play-from-disk-mkv with your browsers SessionDescription as stdin
The `output.mkv` you created should be in the same directory as `play-from-disk-mkv`. In the jsfiddle the top textarea is your browser, copy that and:

#### Linux/macOS
Run `echo $BROWSER_SDP | play-from-disk-mkv`
#### Windows
1. Paste the SessionDescription into a file.
1. Run `play-from-disk-mkv < my_file`

### Input play-from-disk-mkv's SessionDescription into your browser
Copy the text that `play-from-disk-mkv` just emitted and copy into second text area

### Hit 'Start Session' in jsfiddle, enjoy your video!
A video should start playing in your browser above the input boxes. `play-from-disk-mkv` will exit when the file reaches the end

Congrats, you have used Pion WebRTC! Now start building something cool
