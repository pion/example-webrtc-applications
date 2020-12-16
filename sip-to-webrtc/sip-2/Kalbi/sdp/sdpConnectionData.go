package sdp

/*
RFC4566 - https://tools.ietf.org/html/rfc4566#section-5.7

5.7.  Connection Data ("c=")

  c=<nettype> <addrtype> <connection-address>

  c=IN IP4 88.215.55.98
*/

type sdpConnData struct {
	NetType  []byte // Network Type
	AddrType []byte // Address Type
	ConnAddr []byte // Connection Address
	Src      []byte // Full source if needed
}

func (sc *sdpConnData) String() string {
	line := "c="
	line += string(sc.NetType) + " "
	line += string(sc.AddrType) + " "
	line += string(sc.ConnAddr)
	return line
}

func parseSdpConnectionData(v []byte, out *sdpConnData) {

	pos := 0
	state := fieldBase

	// Init the output area
	//out.NetType = nil
	out.AddrType = nil
	out.ConnAddr = nil
	out.Src = nil

	// Keep the source line if needed
	if keepSrc {
		out.Src = v
	}

	// Loop through the bytes making up the line
	for pos < len(v) {
		// FSM
		switch state {
		case fieldBase:
			if v[pos] == ' ' {
				state = fieldAddrType
				pos++
				continue
			}
			out.NetType = append(out.NetType, v[pos])

		case fieldAddrType:
			if v[pos] == ' ' {
				state = fieldConnAddr
				pos++
				continue
			}
			out.AddrType = append(out.AddrType, v[pos])

		case fieldConnAddr:
			if v[pos] == ' ' {
				state = fieldBase
				pos++
				continue
			}
			out.ConnAddr = append(out.ConnAddr, v[pos])
		}
		pos++
	}
}
