package message

/*
Parses a single line that is in the format of a from line, v
Also requires a pointer to a struct of type sipFrom to write output to

RFC 3261 - https://www.ietf.org/rfc/rfc3261.txt - 8.1.1.3 From

*/

//SipFrom SIP From Header
type SipFrom struct {
	UriType  string // Type of URI sip, sips, tel etc
	Name     []byte // Named portion of URI
	User     []byte // User part
	Host     []byte // Host part
	Port     []byte // Port number
	Tag      []byte // Tag
	UserType []byte // User Type
	Src      []byte // Full source if needed
}

//SetUriType sets the uri type e.g. sip , sips
func (sf *SipFrom) SetUriType(uriType string) {
	sf.UriType = uriType
}

//SetUser sets user part of the uri
func (sf *SipFrom) SetUser(user string) {
	sf.User = []byte(user)
}

//SetHost sets host part of uri
func (sf *SipFrom) SetHost(host string) {
	sf.Host = []byte(host)
}

//SetPort sets port of uri
func (sf *SipFrom) SetPort(port string) {
	sf.Port = []byte(port)
}

//SetUserType sets UserType
func (sf *SipFrom) SetUserType(userType string) {
	sf.UserType = []byte(userType)
}

//SetTag sets From Tag
func (sf *SipFrom) SetTag(tag string) {
	sf.Tag = []byte(tag)
}

//String returns Header as String
func (sf *SipFrom) String() string {
	requestline := "From: "
	requestline += "<" + sf.UriType + ":" + string(sf.User) + "@" + string(sf.Host) + ">"

	if sf.Tag != nil {
		requestline += ";tag=" + string(sf.Tag)
	}
	return requestline
}

//ParseSipFrom parser for SIP from Header
func ParseSipFrom(v []byte, out *SipFrom) {

	pos := 0
	state := fieldBase

	// Init the output area
	out.UriType = ""
	out.Name = nil
	out.User = nil
	out.Host = nil
	out.Port = nil
	out.Tag = nil
	out.UserType = nil
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
				// Look for a Tag identifier
				if getString(v, pos, pos+4) == "tag=" {
					state = fieldTag
					pos = pos + 4
					continue
				}
				// Look for other identifiers and ignore
				if v[pos] == '=' {
					state = fieldIgnore
					pos = pos + 1
					continue
				}
				// Look for a User Type identifier
				if getString(v, pos, pos+5) == "user=" {
					state = fieldUserType
					pos = pos + 5
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

		case fieldUserType:
			if v[pos] == ';' || v[pos] == '>' || v[pos] == ' ' {
				state = fieldBase
				pos++
				continue
			}
			out.UserType = append(out.UserType, v[pos])

		case fieldTag:
			if v[pos] == ';' || v[pos] == '>' || v[pos] == ' ' {
				state = fieldBase
				pos++
				continue
			}
			out.Tag = append(out.Tag, v[pos])

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
