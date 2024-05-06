module github.com/pion/example-webrtc-applications/v3

go 1.19

require (
	github.com/at-wat/ebml-go v0.17.1
	github.com/emiago/sipgo v0.17.0
	github.com/google/uuid v1.6.0
	github.com/gorilla/websocket v1.5.1
	github.com/notedit/janus-go v0.0.0-20210115013133-fdce1b146d0e
	github.com/pion/rtcp v1.2.14
	github.com/pion/rtp v1.8.6
	github.com/pion/sdp/v3 v3.0.9
	github.com/pion/webrtc/v3 v3.2.40
	gocv.io/x/gocv v0.35.0
	golang.org/x/image v0.15.0
	golang.org/x/net v0.24.0
)

require (
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/gobwas/ws v1.2.1 // indirect
	github.com/icholy/digest v0.1.22 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.16 // indirect
	github.com/pion/datachannel v1.5.6 // indirect
	github.com/pion/dtls/v2 v2.2.10 // indirect
	github.com/pion/ice/v3 v3.0.7 // indirect
	github.com/pion/interceptor v0.1.29 // indirect
	github.com/pion/logging v0.2.2 // indirect
	github.com/pion/mdns/v2 v2.0.7 // indirect
	github.com/pion/randutil v0.1.0 // indirect
	github.com/pion/sctp v1.8.16 // indirect
	github.com/pion/srtp/v3 v3.0.1 // indirect
	github.com/pion/stun/v2 v2.0.0 // indirect
	github.com/pion/transport/v2 v2.2.4 // indirect
	github.com/pion/transport/v3 v3.0.2 // indirect
	github.com/pion/turn/v3 v3.0.3 // indirect
	github.com/pion/webrtc/v4 v4.0.0-beta.19 // indirect
	github.com/rs/xid v1.4.0 // indirect
	github.com/rs/zerolog v1.28.0 // indirect
	github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b // indirect
	golang.org/x/crypto v0.22.0 // indirect
	golang.org/x/sys v0.19.0 // indirect
)

replace github.com/pion/webrtc/v3 => ../pion/webrtc

replace github.com/pion/interceptor => ../pion/interceptor
