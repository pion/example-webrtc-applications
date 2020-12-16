package message

import (
	"strconv"
)

/*
 RFC 3261 - https://www.ietf.org/rfc/rfc3261.txt

INVITE sip:01798300765@87.252.61.202;user=phone SIP/2.0
SIP/2.0 200 OK

*/

//SipReq is the initial line of a SIP message
type SipReq struct {
	Method     []byte // Sip Method eg INVITE etc
	UriType    string // Type of URI sip, sips, tel etc
	StatusCode []byte // Status Code eg 100
	StatusDesc []byte // Status Code Description eg trying
	User       []byte // User part
	Host       []byte // Host part
	Port       []byte // Port number
	UserType   []byte // User Type
	Src        []byte // Full source if needed
}

//SetMethod gives the ability to set method e.g INVITE, REGISTER
func (sr *SipReq) SetMethod(method string) {
	sr.Method = []byte(method)
}

//SetUriType gives the ability to set URI type e.g. sip:, sips:
func (sr *SipReq) SetUriType(uriType string) {
	sr.UriType = uriType
}

//SetStatusCode gives the ability to set SIP status codes e.g. 200, 100, 302
func (sr *SipReq) SetStatusCode(code int) {
	sr.StatusCode = []byte(strconv.Itoa(code))
}

//SetStatusDesc gives the ability to set status descriptions e.g. OK, Trying, Moved Temporarily
func (sr *SipReq) SetStatusDesc(desc string) {
	sr.StatusDesc = []byte(desc)
}

//SetUser set user portion of uri
func (sr *SipReq) SetUser(user string) {
	sr.User = []byte(user)
}

//SetHost set host protion of uri
func (sr *SipReq) SetHost(host string) {
	sr.Host = []byte(host)
}

//SetPort set port portion of uri
func (sr *SipReq) SetPort(port string) {
	sr.Port = []byte(port)
}

//SetUserType sets user type
func (sr *SipReq) SetUserType(userType string) {
	sr.UserType = []byte(userType)
}

//String returns header as string
func (sr *SipReq) String() string {
	requestline := ""
	if sr.Method != nil {
		requestline += string(sr.Method)
		return requestline + " " + sr.UriType + ":" + string(sr.User) + "@" + string(sr.Host) + " " + string(sr.UserType) + "SIP/2.0"

	}
	return requestline + "SIP/2.0 " + string(sr.StatusCode) + " " + string(sr.StatusDesc)

}

//ParseSipReq parse initial request line
func ParseSipReq(v []byte, out *SipReq) {

	pos := 0
	state := 0

	// Init the output area
	out.UriType = ""
	out.Method = nil
	out.StatusCode = nil
	out.User = nil
	out.Host = nil
	out.Port = nil
	out.UserType = nil
	out.Src = nil

	// Keep the source line if needed
	if keepSrc {
		out.Src = v
	}

	// Loop through the bytes making up the line
	for pos < len(v) {
		// FSM
		switch state {
		case fieldNull:
			if v[pos] >= 'A' && v[pos] <= 'S' && pos == 0 {
				state = fieldMethod
				continue
			}

		case fieldMethod:
			if v[pos] == ' ' || pos > 9 {
				if string(out.Method) == "SIP/2.0" {
					state = fieldStatus
					out.Method = []byte{}
				} else {
					state = fieldBase
				}
				pos++
				continue
			}
			out.Method = append(out.Method, v[pos])

		case fieldBase:
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
				if getString(v, pos, pos+5) == "user=" {
					state = fieldUserType
					pos = pos + 5
					continue
				}
				if v[pos] == '@' {
					state = fieldUserHost
					out.User = out.Host // Move host to user
					out.Host = nil      // Clear the host
					pos++
					continue
				}
			}
		case fieldUser:
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
			if v[pos] == '@' {
				state = fieldUserHost
				out.User = out.Host // Move host to user
				out.Host = nil      // Clear the host
				pos++
				continue
			}
			out.Host = append(out.Host, v[pos]) // Append to host for now

		case fieldUserHost:
			if v[pos] == ':' {
				state = fieldPort
				pos++
				continue
			}
			if v[pos] == ';' || v[pos] == '>' || v[pos] == ' ' {
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

		case fieldStatus:
			if v[pos] == ';' || v[pos] == '>' {
				state = fieldBase
				pos++
				continue
			}
			if v[pos] == ' ' {
				state = fieldStatusDesc
				pos++
				continue
			}
			out.StatusCode = append(out.StatusCode, v[pos])

		case fieldStatusDesc:
			if v[pos] == ';' || v[pos] == '>' {
				state = fieldBase
				pos++
				continue
			}
			out.StatusDesc = append(out.StatusDesc, v[pos])

		}
		pos++
	}
}
