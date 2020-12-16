package sdp

import (
	"strings"
	"time"
)

// ContentType is the media type for an SDP session description.
const ContentType = "application/sdp"

// Session represents an SDP session description.
type Session struct {
	Version     int          // Protocol Version ("v=")
	Origin      *Origin      // Origin ("o=")
	Name        string       // Session Name ("s=")
	Information string       // Session Information ("i=")
	URI         string       // URI ("u=")
	Email       []string     // Email Address ("e=")
	Phone       []string     // Phone Number ("p=")
	Connection  *Connection  // Connection Data ("c=")
	Bandwidth   []*Bandwidth // Bandwidth ("b=")
	TimeZone    []*TimeZone  // TimeZone ("z=")
	Key         []*Key       // Encryption Keys ("k=")
	Timing      *Timing      // Timing ("t=")
	Repeat      []*Repeat    // Repeat Times ("r=")
	Attributes  Attributes   // Session Attributes ("a=")
	Mode        string       // Streaming mode ("sendrecv", "recvonly", "sendonly", or "inactive")
	Media       []*Media     // Media Descriptions ("m=")
}

// String returns the encoded session description as string.
func (s *Session) String() string {
	return string(s.Bytes())
}

// Bytes returns the encoded session description as buffer.
func (s *Session) Bytes() []byte {
	e := NewEncoder(nil)
	e.Encode(s)
	return e.Bytes()
}

// Origin represents an originator of the session.
type Origin struct {
	Username       string
	SessionID      int64
	SessionVersion int64
	Network        string
	Type           string
	Address        string
}

const (
	NetworkInternet = "IN"
)

const (
	TypeIPv4 = "IP4"
	TypeIPv6 = "IP6"
)

// Connection contains connection data.
type Connection struct {
	Network    string
	Type       string
	Address    string
	TTL        int
	AddressNum int
}

// Bandwidth contains session or media bandwidth information.
type Bandwidth struct {
	Type  string
	Value int
}

// TimeZone represents a time zones change information for a repeated session.
type TimeZone struct {
	Time   time.Time
	Offset time.Duration
}

// Key contains a key exchange information.
// Deprecated. Use for backwards compatibility only.
type Key struct {
	Method, Value string
}

// Timing specifies start and stop times for a session.
type Timing struct {
	Start time.Time
	Stop  time.Time
}

// Repeat specifies repeat times for a session.
type Repeat struct {
	Interval time.Duration
	Duration time.Duration
	Offsets  []time.Duration
}

// Media contains media description.
type Media struct {
	Type        string
	Port        int
	PortNum     int
	Proto       string
	Information string        // Media Information ("i=")
	Connection  []*Connection // Connection Data ("c=")
	Bandwidth   []*Bandwidth  // Bandwidth ("b=")
	Key         []*Key        // Encryption Keys ("k=")
	Attributes                // Attributes ("a=")
	Mode        string        // Streaming mode ("sendrecv", "recvonly", "sendonly", or "inactive")
	Format      []*Format     // Media Format for RTP/AVP or RTP/SAVP protocols ("rtpmap", "fmtp", "rtcp-fb")
	FormatDescr string        // Media Format for other protocols
}

// Streaming modes.
const (
	SendRecv = "sendrecv"
	SendOnly = "sendonly"
	RecvOnly = "recvonly"
	Inactive = "inactive"
)

// NegotiateMode negotiates streaming mode.
func NegotiateMode(local, remote string) string {
	switch local {
	case SendRecv:
		switch remote {
		case RecvOnly:
			return SendOnly
		case SendOnly:
			return RecvOnly
		default:
			return remote
		}
	case SendOnly:
		switch remote {
		case SendRecv, RecvOnly:
			return SendOnly
		}
	case RecvOnly:
		switch remote {
		case SendRecv, SendOnly:
			return RecvOnly
		}
	}
	return Inactive
}

// DeleteAttr removes all elements with name from attrs.
func DeleteAttr(attrs Attributes, name ...string) Attributes {
	n := 0
loop:
	for _, it := range attrs {
		for _, v := range name {
			if it.Name == v {
				continue loop
			}
		}
		attrs[n] = it
		n++
	}
	return attrs[:n]
}

// FormatByPayload returns format description by payload type.
func (m *Media) FormatByPayload(payload uint8) *Format {
	for _, f := range m.Format {
		if f.Payload == payload {
			return f
		}
	}
	return nil
}

// Format is a media format description represented by "rtpmap" attributes.
type Format struct {
	Payload   uint8
	Name      string
	ClockRate int
	Channels  int
	Feedback  []string // "rtcp-fb" attributes
	Params    []string // "fmtp" attributes
}

func (f *Format) String() string {
	return f.Name
}

var epoch = time.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC)

func isRTP(media, proto string) bool {
	switch media {
	case "audio", "video":
		return strings.Contains(proto, "RTP/AVP") || strings.Contains(proto, "RTP/SAVP")
	default:
		return false
	}
}
