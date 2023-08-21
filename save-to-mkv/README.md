# save-to-mkv
save-to-mkv is a simple application that shows how to receive opus audio and h264 video using Pion and then save to MKV container.

## Instructions
### Open save-to-mkv example page
[jsfiddle.net](https://jsfiddle.net/07jknmed/) you should see your Webcam, two text-areas and a 'Start Session' button

### Run save-to-mkv with your browsers SessionDescription as stdin
In the jsfiddle the top textarea is your browser, copy that and:
#### Linux/macOS
Run `echo $BROWSER_SDP | save-to-mkv`
#### Windows
1. Paste the SessionDescription into a file.
1. Run `save-to-mkv < my_file`

### Input save-to-webm's SessionDescription into your browser
Copy the text that `save-to-mkv` just emitted and copy into second text area

### Hit 'Start Session' in jsfiddle, enjoy your media!
Your video and/or audio should be saved to `test.mkv`, and will continue playing until you stop the application by Ctrl+C.

Congrats, you have used pion-WebRTC! Now start building something cool
