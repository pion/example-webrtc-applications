package sdp

import (
	"io"
	"strconv"
	"time"
)

// An Encoder writes a session description to a buffer.
type Encoder struct {
	w io.Writer
	b writer
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return NewEncoderSize(w, 1024)
}

// NewEncoderSize returns a new encoder that writes to w.
func NewEncoderSize(w io.Writer, size int) *Encoder {
	return &Encoder{w: w, b: make([]byte, 0, size)}
}

// Encode encodes the session description.
func (e *Encoder) Encode(s *Session) error {
	e.Reset()
	e.b = e.b.session(s)
	if e.w != nil {
		return e.Flush()
	}
	return nil
}

// Flush writes encoded bytes to w.
func (e *Encoder) Flush() error {
	if b := e.b; len(b) > 0 {
		n, err := e.w.Write(b)
		e.b = b[:copy(b, b[n:])]
		return err
	}
	return nil
}

// Reset resets encoder state to be empty.
func (e *Encoder) Reset() {
	e.b = e.b[:0]
	//	e.pos, e.newline = 0, false
}

// Bytes returns encoded bytes of the last session description.
// The bytes stop being valid at the next encoder call.
func (e *Encoder) Bytes() []byte {
	return e.b
}

// Bytes returns the encoded session description as str.
func (e *Encoder) String() string {
	return string(e.b)
}

type writer []byte

func (w writer) char(v byte) writer {
	return append(w, v)
}

func (w writer) add(v byte) writer {
	return append(w, '\r', '\n', v, '=')
}

func (w writer) str(v string) writer {
	if v == "" {
		return append(w, '-')
	}
	return append(w, v...)
}

func (w writer) int(v int64) writer {
	return strconv.AppendInt(w, v, 10)
}

func (w writer) sp() writer {
	return append(w, ' ')
}

func (w writer) crlf() writer {
	return append(w, '\r', '\n')
}

func (w writer) session(s *Session) writer {
	w = w.str("v=").int(int64(s.Version))
	if s.Origin != nil {
		w = w.add('o').origin(s.Origin)
	}
	w = w.add('s').str(s.Name)
	if s.Information != "" {
		w = w.add('i').str(s.Information)
	}
	if s.URI != "" {
		w = w.add('u').str(s.URI)
	}
	for _, it := range s.Email {
		w = w.add('e').str(it)
	}
	for _, it := range s.Phone {
		w = w.add('p').str(it)
	}
	if s.Connection != nil {
		w = w.add('c').connection(s.Connection)
	}
	for _, b := range s.Bandwidth {
		w = w.add('b').bandwidth(b)
	}
	if len(s.TimeZone) > 0 {
		w = w.add('z').timezone(s.TimeZone)
	}
	for _, it := range s.Key {
		w = w.add('k').key(it)
	}
	w = w.add('t').timing(s.Timing)
	for _, it := range s.Repeat {
		w = w.add('r').repeat(it)
	}
	if s.Mode != "" {
		w = w.add('a').str(s.Mode)
	}
	for _, it := range s.Attributes {
		w = w.add('a').attr(it)
	}
	for _, it := range s.Media {
		w = w.media(it)
	}
	return w.crlf()
}

func (w writer) origin(o *Origin) writer {
	return w.str(strdef(o.Username, "-")).sp().int(o.SessionID).sp().int(o.SessionVersion).sp().transport(o.Network, o.Type, o.Address)
}

func (w writer) media(m *Media) writer {
	w = w.add('m').str(m.Type).sp().int(int64(m.Port))
	if m.PortNum > 0 {
		w = w.char('/').int(int64(m.PortNum))
	}
	w = w.sp().str(m.Proto)
	if f := m.FormatDescr; f != "" {
		w = w.sp().str(f)
	} else {
		for _, it := range m.Format {
			w = w.sp().int(int64(it.Payload))
		}
	}
	if m.Information != "" {
		w = w.add('i').str(m.Information)
	}
	for _, it := range m.Connection {
		w = w.add('c').connection(it)
	}
	for _, b := range m.Bandwidth {
		w = w.add('b').bandwidth(b)
	}
	for _, it := range m.Key {
		w = w.add('k').key(it)
	}
	for _, it := range m.Format {
		w = w.format(it)
	}
	if m.Mode != "" {
		w = w.add('a').str(m.Mode)
	}
	for _, it := range m.Attributes {
		w = w.add('a').attr(it)
	}
	return w
}

func (w writer) format(f *Format) writer {
	p := int64(f.Payload)
	if f.Name != "" {
		w = w.add('a').str("rtpmap:").int(p).sp().str(f.Name).char('/').int(int64(f.ClockRate))
		if f.Channels > 0 {
			w = w.char('/').int(int64(f.Channels))
		}
	}
	for _, it := range f.Feedback {
		w = w.add('a').str("rtcp-fb:").int(p).sp().str(it)
	}
	for _, it := range f.Params {
		w = w.add('a').str("fmtp:").int(p).sp().str(it)
	}
	return w
}

func (w writer) attr(a *Attr) writer {
	if a.Value == "" {
		return w.str(a.Name)
	}
	return w.str(a.Name).char(':').str(a.Value)
}

func (w writer) timezone(z []*TimeZone) writer {
	for i, it := range z {
		if i > 0 {
			w = w.char(' ')
		}
		w = w.time(it.Time).sp().duration(it.Offset)
	}
	return w
}

func (w writer) timing(t *Timing) writer {
	if t == nil {
		return w.str("0 0")
	}
	return w.time(t.Start).sp().time(t.Stop)
}

func (w writer) repeat(r *Repeat) writer {
	w = w.duration(r.Interval).sp().duration(r.Duration)
	for _, it := range r.Offsets {
		w = w.sp().duration(it)
	}
	return w
}

func (w writer) time(t time.Time) writer {
	if t.IsZero() {
		return w.char('0')
	}
	return w.int(int64(t.Sub(epoch).Seconds()))
}

func (w writer) duration(d time.Duration) writer {
	v := int64(d.Seconds())
	switch {
	case v == 0:
		return w.char('0')
	case v%86400 == 0:
		return w.int(v / 86400).char('d')
	case v%3600 == 0:
		return w.int(v / 3600).char('h')
	case v%60 == 0:
		return w.int(v / 60).char('m')
	default:
		return w.int(v)
	}
}

func (w writer) bandwidth(b *Bandwidth) writer {
	return w.str(b.Type).char(':').int(int64(b.Value))
}

func (w writer) key(k *Key) writer {
	if k.Value == "" {
		return w.str(k.Method)
	}
	return w.str(k.Method).char(':').str(k.Value)
}

func (w writer) connection(c *Connection) writer {
	w = w.transport(c.Network, c.Type, c.Address)
	if c.TTL > 0 {
		w = w.char('/').int(int64(c.TTL))
	}
	if c.AddressNum > 1 {
		w = w.char('/').int(int64(c.AddressNum))
	}
	return w
}

func (w writer) transport(network, typ, addr string) writer {
	return w.fields(strdef(network, NetworkInternet), strdef(typ, TypeIPv4), strdef(addr, "127.0.0.1"))
}

func strdef(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

func (w writer) fields(v ...string) writer {
	for i, s := range v {
		if i > 0 {
			w = append(w, ' ')
		}
		w = append(w, s...)
	}
	return w
}
