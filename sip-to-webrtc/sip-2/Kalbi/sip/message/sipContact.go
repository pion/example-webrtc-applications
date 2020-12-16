package message

/*

RFC 3261 - https://www.ietf.org/rfc/rfc3261.txt - 8.1.1.8 Contact

   The Contact header field provides a SIP or SIPS URI that can be used
   to contact that specific instance of the UA for subsequent requests.
   The Contact header field MUST be present and contain exactly one SIP
   or SIPS URI in any request that can result in the establishment of a
   dialog.

Examples:

   Contact: "Mr. Watson" <sip:watson@worcester.bell-telephone.com>
      ;q=0.7; expires=3600,
      "Mr. Watson" <mailto:watson@bell-telephone.com> ;q=0.1
   m: <sips:bob@192.0.2.4>;expires=60

*/

//SipContact SIP Contact Header
type SipContact struct {
	UriType string // Type of URI sip, sips, tel etc
	Name    []byte // Named portion of URI
	User    []byte // User part
	Host    []byte // Host part
	Port    []byte // Port number
	Tran    []byte // Transport
	Qval    []byte // Q Value
	Expires []byte // Expires
	Src     []byte // Full source if needed
}

//SetUriType sets the uri type e.g. sip , sips
func (sc *SipContact) SetUriType(uriType string) {
	sc.UriType = uriType
}

//SetUser sets user part of the uri
func (sc *SipContact) SetUser(user string) {
	sc.User = []byte(user)
}

//SetHost sets host part of uri
func (sc *SipContact) SetHost(host string) {
	sc.Host = []byte(host)
}

//SetPort sets port of uri
func (sc *SipContact) SetPort(port string) {
	sc.Port = []byte(port)
}

//SetName stes name portion of header
func (sc *SipContact) SetName(name string) {
	sc.Name = []byte(name)

}

//String returns Header as String
func (sc *SipContact) String() string {
	line := "Contact: "
	if sc.Name != nil {
		line += string(sc.Name) + " "
	}
	line += "<" + sc.UriType + ":" + string(sc.User) + "@" + string(sc.Host)
	if sc.Port != nil {
		line += ":" + string(sc.Port) + ">"
	} else {
		line += ">"
	}
	if sc.Tran != nil {
		line += ";transport=" + string(sc.Tran)
	}
	if sc.Qval != nil {
		line += ";q=" + string(sc.Qval)
	}
	if sc.Expires != nil {
		line += ";expires=" + string(sc.Expires)
	}
	return line
}

//ParseSipContact parses SIP contact header
func ParseSipContact(v []byte, out *SipContact) {

	pos := 0
	state := fieldBase

	// Init the output area
	out.UriType = ""
	out.Name = nil
	out.User = nil
	out.Host = nil
	out.Port = nil
	out.Tran = nil
	out.Qval = nil
	out.Expires = nil
	out.Src = nil

	// Keep the source line if needed
	if keepSrc {
		out.Src = v
	}

	// Loop through the bytes making up the line
	for pos < len(v) {
		// FSM
		//fmt.Println("POS:", pos, "CHR:", string(v[pos]), "STATE:", state)
		switch state {
		case fieldBase:
			if v[pos] == '"' && out.UriType == "" {
				state = fieldNameQ
				pos++
				continue
			}
			if v[pos] != ' ' {
				// Not a space so check for uri types
				if getString(v, pos, pos+4) == "sip:" {
					state = fieldUser
					pos = pos + 4
					out.UriType = "sip"
					continue
				}
				if getString(v, pos, pos+5) == "sips:" {
					state = fieldUser
					pos = pos + 5
					out.UriType = "sips"
					continue
				}
				if getString(v, pos, pos+4) == "tel:" {
					state = fieldUser
					pos = pos + 4
					out.UriType = "tel"
					continue
				}
				// Look for a Q identifier
				if getString(v, pos, pos+2) == "q=" {
					state = fieldQ
					pos = pos + 2
					continue
				}
				// Look for a Expires identifier
				if getString(v, pos, pos+8) == "expires=" {
					state = fieldExpires
					pos = pos + 8
					continue
				}
				// Look for a transport identifier
				if getString(v, pos, pos+10) == "transport=" {
					state = fieldTran
					pos = pos + 10
					continue
				}
				// Look for other identifiers and ignore
				if v[pos] == '=' {
					state = fieldIgnore
					pos = pos + 1
					continue
				}
				// Check for other chrs
				if v[pos] != '<' && v[pos] != '>' && v[pos] != ';' && out.UriType == "" {
					state = fieldName
					continue
				}
			}

		case fieldNameQ:
			if v[pos] == '"' {
				state = fieldBase
				pos++
				continue
			}
			out.Name = append(out.Name, v[pos])

		case fieldName:
			if v[pos] == '<' || v[pos] == ' ' {
				state = fieldBase
				pos++
				continue
			}
			out.Name = append(out.Name, v[pos])

		case fieldUser:
			if v[pos] == '@' {
				state = fieldUserHost
				pos++
				continue
			}
			out.User = append(out.User, v[pos])

		case fieldUserHost:
			if v[pos] == ':' {
				state = fieldPort
				pos++
				continue
			}
			if v[pos] == ';' || v[pos] == '>' {
				state = fieldBase
				pos++
				continue
			}
			out.Host = append(out.Host, v[pos])

		case fieldPort:
			if v[pos] == ';' || v[pos] == '>' || v[pos] == ' ' {
				state = fieldBase
				pos++
				continue
			}
			out.Port = append(out.Port, v[pos])

		case fieldTran:
			if v[pos] == ';' || v[pos] == '>' || v[pos] == ' ' {
				state = fieldBase
				pos++
				continue
			}
			out.Tran = append(out.Tran, v[pos])

		case fieldQ:
			if v[pos] == ';' || v[pos] == '>' || v[pos] == ' ' {
				state = fieldBase
				pos++
				continue
			}
			out.Qval = append(out.Qval, v[pos])

		case fieldExpires:
			if v[pos] == ';' || v[pos] == '>' || v[pos] == ' ' {
				state = fieldBase
				pos++
				continue
			}
			out.Expires = append(out.Expires, v[pos])

		case fieldIgnore:
			if v[pos] == ';' || v[pos] == '>' {
				state = fieldBase
				pos++
				continue
			}

		}
		pos++
	}
}
