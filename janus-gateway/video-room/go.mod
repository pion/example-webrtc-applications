module videoroom

go 1.12

replace gstreamer-src v0.0.0 => ../../internal/gstreamer-src

require (
	github.com/gorilla/websocket v1.4.0 // indirect
	github.com/notedit/janus-go v0.0.0-20180821162543-a152adf0cb7b
	github.com/pion/webrtc/v2 v2.0.17
	gstreamer-src v0.0.0
)
