# c-data-channels
c-data-channels is a Pion WebRTC application that shows how you can send/recv DataChannel messages from a web browser that's
mostly identical to the pure Go implementation, https://github.com/pion/webrtc/tree/master/examples/data-channels.
The main difference is that the OnDataChannel is fully implemented in C.

## Instructions
### Clone c-data-channels
```
git clone https://github.com/pion/example-webrtc-applications.git
```

### Go into the c-data-channels example directory
```
cd example-webrtc-applications/c-data-channels
```

### Build it
```
make
```

### Open data-channels example page
[jsfiddle.net](https://jsfiddle.net/9tsx15mg/90/)

### Run data-channels, with your browsers SessionDescription as stdin
In the jsfiddle the top textarea is your browser's session description, copy that and:
#### Linux/macOS
Run `echo $BROWSER_SDP | c-data-channels`
#### Windows
1. Paste the SessionDescription into a file.
1. Run `c-data-channels < my_file`

### Input data-channels's SessionDescription into your browser
Copy the text that `c-data-channels` just emitted and copy into second text area

### Hit 'Start Session' in jsfiddle
Under Start Session you should see 'Checking' as it starts connecting. If everything worked you should see `New DataChannel foo 1`

Now you can put whatever you want in the `Message` textarea, and when you hit `Send Message` it should appear in your browser!

You can also type in your terminal, and when you hit enter it will appear in your web browser.

Congrats, you have used Pion WebRTC! Now start building something cool

## Organization

### bridge.go
This file contains all of the bridging between Go and C. This is the only file that contains cgo stuff.

### main.go
This file is pure Go. It is mostly identical to the original data-channel example.

### Reference
* https://github.com/golang/go/issues/20639
* https://github.com/golang/go/issues/25832
* https://github.com/pion/webrtc/tree/master/examples/data-channels/jsfiddle - jsfiddle source codes
