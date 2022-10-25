# gocv-receive
gocv-receive is a simple application that shows how to receive media using Pion and then do motion detection with GoCV.


This example borrows heavily from GoCV's [motion-detect](https://github.com/hybridgroup/gocv/blob/master/cmd/motion-detect/main.go) example.
You could easily implement many other GoCV applications following the same pattern

## Instructions
### Install Dependencies
This example requires you have GoCV and ffmpeg installed, these are the supported platforms
#### Debian/Ubuntu
* Follow the setup instructions for [GoCV](https://github.com/hybridgroup/gocv)
* `sudo apt-get install ffmpeg`
#### macOS
* `brew install ffmpeg opencv`

### Build gocv-receive
```
go build -tags gocv
```

### Open gocv-receive example page
[jsfiddle.net](https://jsfiddle.net/tfmLq8jw/) you should see your Webcam, two text-areas and a 'Start Session' button

### Run gocv-receive with your browsers SessionDescription as stdin
In the jsfiddle the top textarea is your browser, copy that and:
#### Linux/macOS
Run `echo $BROWSER_SDP | gocv-receive`
#### Windows
1. Paste the SessionDescription into a file.
1. Run `gocv-receive < my_file`

### Input gocv-receive's SessionDescription into your browser
Copy the text that `gocv-receive` just emitted and copy into second text area

### Hit 'Start Session' in jsfiddle, enjoy your media!
Your video and/or audio should popup automatically, and will continue playing until you close the application.

Congrats, you have used pion-WebRTC! Now start building something cool
