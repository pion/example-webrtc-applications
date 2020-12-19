package main

/*
#include <stddef.h>
#include <stdint.h>
typedef struct GoDataChannelMessage
{
    // bool in cgo is really weird, so it's simpler to just use classic int as bool
    int is_string;
    void *data;
    size_t data_len;
} GoDataChannelMessage;

typedef uint16_t GoDataChannel;
typedef void (*GoOnDataChannelFunc)(GoDataChannel);
typedef void (*GoOnOpenFunc)(GoDataChannel);
typedef void (*GoOnMessageFunc)(GoDataChannel, GoDataChannelMessage);

// Calling C function pointers is currently not supported, however you can
// declare Go variables which hold C function pointers and pass them back
// and forth between Go and C. C code may call function pointers received from Go.
// Reference: https://golang.org/cmd/cgo/#hdr-Go_references_to_C
inline void bridge_on_data_channel(GoOnDataChannelFunc cb, GoDataChannel d)
{
    cb(d);
}

inline void bridge_on_open(GoOnOpenFunc cb, GoDataChannel d)
{
	cb(d);
}

inline void bridge_on_message(GoOnMessageFunc cb, GoDataChannel d, GoDataChannelMessage msg)
{
	cb(d, msg);
}
*/
import "C"

import (
	"github.com/pion/webrtc/v3"
)

var store = map[C.GoDataChannel]*webrtc.DataChannel{}

//export GoRun
func GoRun(f C.GoOnDataChannelFunc) {
	Run(func(d *webrtc.DataChannel) {
		// Since cgo doesn't allow storing Go pointers in C, we need to store some data in C
		// that can tell Go how to get webrtc.DataChannel later. So, here we simply use data channel's
		// id, which is just a simple 16 unsigned int that we can pass easily from/to C.
		id := C.GoDataChannel(*d.ID())
		store[id] = d
		C.bridge_on_data_channel(f, id)
	})
}

//export GoOnOpen
func GoOnOpen(d C.GoDataChannel, f C.GoOnOpenFunc) {
	// get the actual DataChannel using a unique id
	dc := store[d]
	dc.OnOpen(func() {
		C.bridge_on_open(f, d)
	})
}

//export GoOnMessage
func GoOnMessage(d C.GoDataChannel, f C.GoOnMessageFunc) {
	dc := store[d]
	dc.OnMessage(func(msg webrtc.DataChannelMessage) {
		var isString int
		// Since C interprets non-zero to be true, we can simply set isString to be 1
		// or any non-zero value to make C to think that isString is true
		if msg.IsString {
			isString = 1
		}

		cMsg := C.GoDataChannelMessage{
			is_string: C.int(isString),
			data:      C.CBytes(msg.Data),
			data_len:  C.ulong(len(msg.Data)),
		}
		C.bridge_on_message(f, d, cMsg)
	})
}

//export GoSendText
func GoSendText(d C.GoDataChannel, t *C.char) {
	dc := store[d]
	dc.SendText(C.GoString(t))
}

//export GoLabel
func GoLabel(d C.GoDataChannel) *C.char {
	dc := store[d]
	return C.CString(dc.Label())
}
