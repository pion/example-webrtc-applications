package message

import (
	"strings"
)

/*
 RFC 3261 - https://www.ietf.org/rfc/rfc3261.txt - 8.1.1.7 Via

 The Via header field indicates the transport used for the transaction
and identifies the location where the response is to be sent.  A Via
header field value is added only after the transport that will be
used to reach the next hop has been selected (which may involve the
usage of the procedures in [4]).

*/

//SipVia SIP Via Header
type SipVia struct {
	Trans  string // Type of Transport udp, tcp, tls, sctp etc
	Host   []byte // Host part
	Port   []byte // Port number
	Branch []byte //
	Rport  []byte //
	Maddr  []byte //
	Ttl    []byte //
	Rcvd   []byte //
	Src    []byte // Full source if needed
}

//String returns Header as String
func (sv *SipVia) String() string {
	return "Via: SIP/2.0/" + strings.ToUpper(sv.Trans) + " " + string(sv.Host) + ":" + string(sv.Port) + ";branch=" + string(sv.Branch)
}

//SetTransport sets transport in via header
func (sv *SipVia) SetTransport(trans string) {
	sv.Trans = strings.ToUpper(trans)
}

//SetHost set host portion of uri
func (sv *SipVia) SetHost(value string) {
	sv.Host = []byte(value)
}

//SetPort sets port portion of uri
func (sv *SipVia) SetPort(value string) {
	sv.Port = []byte(value)
}

//SetBranch sets branch
func (sv *SipVia) SetBranch(value string) {
	sv.Branch = []byte(value)
}

//ParseSipVia parses SIP Via Header
func ParseSipVia(v []byte, out *SipVia) {

	pos := 0
	state := fieldBase

	// Init the output area
	out.Trans = ""
	out.Host = nil
	out.Port = nil
	out.Branch = nil
	out.Rport = nil
	out.Maddr = nil
	out.Ttl = nil
	out.Rcvd = nil
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
			if v[pos] != ' ' {
				// Not a space
				if getString(v, pos, pos+8) == "SIP/2.0/" {
					// Transport type
					state = fieldUserHost
					pos = pos + 8
					if getString(v, pos, pos+3) == "UDP" {
						out.Trans = "udp"
						pos = pos + 3
						continue
					}
					if getString(v, pos, pos+3) == "TCP" {
						out.Trans = "tcp"
						pos = pos + 3
						continue
					}
					if getString(v, pos, pos+3) == "TLS" {
						out.Trans = "tls"
						pos = pos + 3
						continue
					}
					if getString(v, pos, pos+4) == "SCTP" {
						out.Trans = "sctp"
						pos = pos + 4
						continue
					}
				}
				// Look for a Branch identifier
				if getString(v, pos, pos+7) == "branch=" {
					state = fieldBranch
					pos = pos + 7
					continue
				}
				// Look for a Rport identifier
				if getString(v, pos, pos+6) == "rport=" {
					state = fieldRport
					pos = pos + 6
					continue
				}
				// Look for a maddr identifier
				if getString(v, pos, pos+6) == "maddr=" {
					state = fieldMaddr
					pos = pos + 6
					continue
				}
				// Look for a ttl identifier
				if getString(v, pos, pos+4) == "ttl=" {
					state = fieldTTL
					pos = pos + 4
					continue
				}
				// Look for a recevived identifier
				if getString(v, pos, pos+9) == "received=" {
					state = fieldRec
					pos = pos + 9
					continue
				}
			}

		case fieldUserHost:
			if v[pos] == ':' {
				state = fieldPort
				pos++
				continue
			}
			if v[pos] == ';' {
				state = fieldBase
				pos++
				continue
			}
			if v[pos] == ' ' {
				pos++
				continue
			}
			out.Host = append(out.Host, v[pos])

		case fieldPort:
			if v[pos] == ';' {
				state = fieldBase
				pos++
				continue
			}
			out.Port = append(out.Port, v[pos])

		case fieldBranch:
			if v[pos] == ';' {
				state = fieldBase
				pos++
				continue
			}
			out.Branch = append(out.Branch, v[pos])

		case fieldRport:
			if v[pos] == ';' {
				state = fieldBase
				pos++
				continue
			}
			out.Rport = append(out.Rport, v[pos])

		case fieldMaddr:
			if v[pos] == ';' {
				state = fieldBase
				pos++
				continue
			}
			out.Maddr = append(out.Maddr, v[pos])

		case fieldTTL:
			if v[pos] == ';' {
				state = fieldBase
				pos++
				continue
			}
			out.Ttl = append(out.Ttl, v[pos])

		case fieldRec:
			if v[pos] == ';' {
				state = fieldBase
				pos++
				continue
			}
			out.Rcvd = append(out.Rcvd, v[pos])
		}
		pos++
	}
}
