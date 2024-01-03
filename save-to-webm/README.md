# save-to-webm
save-to-webm is a simple application that shows how to receive audio and video using Pion and then save to WebM container.

## Instructions
### Open save-to-webm example page
[jsfiddle.net](https://jsfiddle.net/07jknmed/) you should see your Webcam, two text-areas and a 'Start Session' button

### Run save-to-webm with your browsers SessionDescription as stdin
In the jsfiddle the top textarea is your browser, copy that and:
#### Linux/macOS
Run `echo $BROWSER_SDP | save-to-webm`
#### Windows
1. Paste the SessionDescription into a file.
1. Run `save-to-webm < my_file`

### Input save-to-webm's SessionDescription into your browser
Copy the text that `save-to-webm` just emitted and copy into second text area

### Hit 'Start Session' in jsfiddle, enjoy your media!
Your video and/or audio should be saved to `test.webm`, and will continue playing until you stop the application by Ctrl+C.

Congrats, you have used Pion WebRTC! Now start building something cool
