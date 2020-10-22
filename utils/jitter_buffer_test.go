package utils

import (
	"fmt"
	"strings"
	"testing"

	"github.com/pion/rtp"
)

type testAdd struct {
	seq       int64
	payload   string
	wantAdded bool
}

type testBuffer struct {
	wantNextStart int64
	wantEnd       int64
	wantPayload   string
}

func TestJitterBuffer(t *testing.T) {
	b := NewJitterBuffer(8)

	wantAdd(t, b, testAdd{seq: 100, payload: "0", wantAdded: true})
	wantBuffer(t, b, testBuffer{wantNextStart: 100, wantEnd: 100, wantPayload: "-,-,-,-,0,-,-,-"})
	wantAdd(t, b, testAdd{seq: 101, payload: "1", wantAdded: true})
	wantBuffer(t, b, testBuffer{wantNextStart: 100, wantEnd: 101, wantPayload: "-,-,-,-,0,1,-,-"})
	wantAdd(t, b, testAdd{seq: 103, payload: "3", wantAdded: true})
	wantBuffer(t, b, testBuffer{wantNextStart: 100, wantEnd: 103, wantPayload: "-,-,-,-,0,1,-,3"})
	wantAdd(t, b, testAdd{seq: 104, payload: "4", wantAdded: true})
	wantBuffer(t, b, testBuffer{wantNextStart: 100, wantEnd: 104, wantPayload: "4,-,-,-,0,1,-,3"})
	wantAdd(t, b, testAdd{seq: 106, payload: "6", wantAdded: true})
	wantBuffer(t, b, testBuffer{wantNextStart: 100, wantEnd: 106, wantPayload: "4,-,6,-,0,1,-,3"})
	wantAdd(t, b, testAdd{seq: 105, payload: "5", wantAdded: true})
	wantBuffer(t, b, testBuffer{wantNextStart: 100, wantEnd: 106, wantPayload: "4,5,6,-,0,1,-,3"})
	wantAdd(t, b, testAdd{seq: 107, payload: "7", wantAdded: true})
	wantBuffer(t, b, testBuffer{wantNextStart: 100, wantEnd: 107, wantPayload: "4,5,6,7,0,1,-,3"})

	wantNextPackets(t, b, "0,1")
	wantBuffer(t, b, testBuffer{wantNextStart: 102, wantEnd: 107, wantPayload: "4,5,6,7,-,-,-,3"})
	wantNextPackets(t, b, "")

	wantAdd(t, b, testAdd{seq: 108, payload: "8", wantAdded: true})
	wantBuffer(t, b, testBuffer{wantNextStart: 102, wantEnd: 108, wantPayload: "4,5,6,7,8,-,-,3"})
	wantAdd(t, b, testAdd{seq: 109, payload: "9", wantAdded: true})
	wantBuffer(t, b, testBuffer{wantNextStart: 102, wantEnd: 109, wantPayload: "4,5,6,7,8,9,-,3"})
	wantAdd(t, b, testAdd{seq: 110, payload: "a", wantAdded: true})
	wantBuffer(t, b, testBuffer{wantNextStart: 103, wantEnd: 110, wantPayload: "4,5,6,7,8,9,a,3"})

	wantNextPackets(t, b, "3,4,5,6,7,8,9,a")
	wantBuffer(t, b, testBuffer{wantNextStart: 111, wantEnd: 110, wantPayload: "-,-,-,-,-,-,-,-"})
	wantNextPackets(t, b, "")

	wantAdd(t, b, testAdd{seq: 2, payload: "2", wantAdded: false})
	wantBuffer(t, b, testBuffer{wantNextStart: 111, wantEnd: 110, wantPayload: "-,-,-,-,-,-,-,-"})
}

func wantNextPackets(t *testing.T, b *JitterBuffer, wantPayload string) {
	t.Helper()
	gotPayload := packetsPayload(b.NextPackets())
	errors := make([]string, 0)
	if wantPayload != gotPayload {
		errors = append(errors, fmt.Sprintf("payload want/got: %s/%s", wantPayload, gotPayload))
	}

	if len(errors) > 0 {
		t.Errorf("NextPackets %s", strings.Join(errors, " "))
	}
}

func wantAdd(t *testing.T, b *JitterBuffer, add testAdd) {
	t.Helper()
	added := b.Add(add.seq, &rtp.Packet{
		Payload: []byte(add.payload),
	},
	)
	if add.wantAdded != added {
		t.Errorf("Added (seq: %d, payload: %s) added want/got: %t/%t", add.seq, add.payload, add.wantAdded, added)
	}
}

func wantBuffer(t *testing.T, b *JitterBuffer, buffer testBuffer) {
	t.Helper()

	errors := make([]string, 0)
	if buffer.wantNextStart != b.nextStart {
		errors = append(errors, fmt.Sprintf("nextStart want/got: %d/%d", buffer.wantNextStart, b.nextStart))
	}
	if buffer.wantEnd != b.end {
		errors = append(errors, fmt.Sprintf("end want/got: %d/%d", buffer.wantEnd, b.end))
	}
	payload := packetsPayload(b.packets)
	if buffer.wantPayload != payload {
		errors = append(errors, fmt.Sprintf("payload want/got: %s/%s", buffer.wantPayload, payload))
	}

	if len(errors) > 0 {
		t.Errorf("Buffer %s", strings.Join(errors, " "))
	}
}

func packetsPayload(packets []*rtp.Packet) string {
	results := make([]string, 0, len(packets))
	for _, p := range packets {
		if p == nil {
			results = append(results, "-")
		} else {
			results = append(results, string(p.Payload))
		}
	}

	return strings.Join(results, ",")
}
