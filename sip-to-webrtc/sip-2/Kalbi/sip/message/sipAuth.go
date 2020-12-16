package message

/*

20.7 Authorization

   The Authorization header field contains authentication credentials of
   a UA.  Section 22.2 overviews the use of the Authorization header
   field, and Section 22.4 describes the syntax and semantics when used
   with HTTP authentication.

   This header field, along with Proxy-Authorization, breaks the general
   rules about multiple header field values.  Although not a comma-
   separated list, this header field name may be present multiple times,
   and MUST NOT be combined into a single header line using the usual
   rules described in Section 7.3.

*/

//SipAuth SIP Authorization Header
type SipAuth struct {
	Username  []byte
	Realm     []byte
	Nonce     []byte
	CNonce    []byte
	QoP       []byte
	Algorithm []byte
	Nc        []byte
	URI       []byte
	Response  []byte
	Src       []byte
}

//SetUsername sets username 
func (sa *SipAuth) SetUsername(value string) {
    sa.Username = []byte(value)
}

//GetUsername returns username
func (sa *SipAuth) GetUsername() string {
    return string(sa.Username)
}

//SetRealm sets realm
func (sa *SipAuth) SetRealm(value string) {
    sa.Realm = []byte(value)
}

//GetRealm returns realm
func (sa *SipAuth) GetRealm() string {
    return string(sa.Realm)
}

//SetNonce sets nonce
func (sa *SipAuth) SetNonce(value string) {
	sa.Nonce = []byte(value)
}

//GetNonce returns nonce
func (sa *SipAuth) GetNonce() string {
    return string(sa.Nonce)
}

//SetCNonce sets cnonce
func (sa *SipAuth) SetCNonce(value string) {
	sa.CNonce = []byte(value)
}

//GetCNonce returns cnonce
func (sa *SipAuth) GetCNonce() string {
    return string(sa.CNonce)
}

//SetQoP sets qop
func (sa *SipAuth) SetQoP(value string) {
    sa.QoP = []byte(value)
}

//GetQoP returns qop
func (sa *SipAuth) GetQoP() string {
    return string(sa.QoP)
}

//SetAlgorithm sets algorithm
func (sa *SipAuth) SetAlgorithm(value string) {
    sa.Algorithm = []byte(value)
}

//GetAlgorithm returns algorithm
func (sa *SipAuth) GetAlgorithm() string {
    return string(sa.Algorithm)
}

//SetNc sets nc
func (sa *SipAuth) SetNc(value string) {
    sa.Nc = []byte(value)
}

//GetAlgorithm returns nc
func (sa *SipAuth) GetNc() string {
    return string(sa.Nc)
}

//SetURI sets nc
func (sa *SipAuth) SetURI(value string) {
    sa.URI = []byte(value)
}

//GetURI returns nc
func (sa *SipAuth) GetURI() string {
    return string(sa.URI)
}

//SetResponse sets response
func (sa *SipAuth) SetResponse(value string) {
    sa.Response = []byte(value)
}

//GetResponse returns response
func (sa *SipAuth) GetResponse() string {
    return string(sa.Response)
}

func (sa *SipAuth) String() string {
	line := "DIGEST "
	if sa.QoP != nil {
		line += "qop=" + string(sa.QoP) + " "
	} 
	if sa.Nonce != nil {
		line += ",nonce=\"" + string(sa.Nonce) + "\" " 
	}
	if sa.Realm != nil {
        line += ",realm=\"" + string(sa.Realm) + "\" "
	}
	if sa.Algorithm != nil {
        line += ",algorithm=" + string(sa.Algorithm) + " "
	}
	if sa.Username != nil {
        line += ",username=\"" + string(sa.Username) + "\" "
	}
	if sa.URI != nil {
        line += ",uri=\"" + string(sa.URI) + "\" "
	}
	if sa.Nc != nil {
		line += ",nc="+ string(sa.Nc) + " "
	}
	if sa.Response != nil {
		line += ",response=\"" + string(sa.Response) + "\" "
	}
	if sa.CNonce != nil {
		line += ",cnonce=\"" + string(sa.CNonce) + "\" "
	}
    return line
}


//ParseSipAuth parse's WWW-Authenticate/Authorization headers
func ParseSipAuth(v []byte, out *SipAuth) {

	pos := 0
	state := fieldBase

	// Init the output area
	out.Username = nil
	out.Realm = nil
	out.Nonce = nil
	out.CNonce = nil
	out.QoP = nil
	out.Algorithm = nil
	out.Nc = nil
	out.Response = nil
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
			if v[pos] != ',' || v[pos] != ' ' {

				if getString(v, pos, pos+9) == "username=" {
					state = fieldUser
					if v[pos+9] == '"' {
						pos = pos + 10
					} else {
						pos = pos + 9
					}
					continue
				}

				if getString(v, pos, pos+9) == "response=" {
					state = fieldResponse
					if v[pos+9] == '"' {
						pos = pos + 10
					} else {
						pos = pos + 9
					}
					continue

				}

				if getString(v, pos, pos+4) == "qop=" {
					state = fieldQop
					if v[pos+4] == '"' {
						pos = pos + 5
					} else {
						pos = pos + 4
					}
					continue
				}

				if getString(v, pos, pos+4) == "uri=" {
					state = fieldURI
					if v[pos+4] == '"' {
						pos = pos + 5
					} else {
						pos = pos + 4
					}
					continue
				}

				if getString(v, pos, pos+3) == "nc=" {
					state = fieldNC
					pos = pos + 3
					continue
				}

				if getString(v, pos, pos+7) == "cnonce=" {
					state = fieldCNonce
					pos = pos + 8
					continue
				}

				if getString(v, pos, pos+6) == "nonce=" {
					state = fieldNonce
					pos = pos + 7
					continue
				}
				if getString(v, pos, pos+6) == "realm=" {
					state = fieldRealm
					pos = pos + 7
					continue
				}
				if getString(v, pos, pos+10) == "algorithm=" {
					state = fieldAlgorithm
					pos = pos + 10
					continue
				}

			}

		case fieldQop:
			if v[pos] == ' ' || v[pos] == ',' || v[pos] == '"' {
				state = fieldBase
				pos++
				continue
			}
			out.QoP = append(out.QoP, v[pos])

		case fieldNonce:
			if v[pos] == '"' {
				state = fieldBase
				pos++
				continue
			}
			out.Nonce = append(out.Nonce, v[pos])

		case fieldCNonce:
			if v[pos] == '"' {
				state = fieldBase
				pos++
				continue
			}
			out.CNonce = append(out.CNonce, v[pos])

		case fieldURI:
			if v[pos] == '"' {
				state = fieldBase
				pos++
				continue
			}
			out.URI = append(out.URI, v[pos])

		case fieldResponse:
			if v[pos] == '"' {
				state = fieldBase
				pos++
				continue
			}
			out.Response = append(out.Response, v[pos])

		case fieldRealm:
			if v[pos] == '"' {
				state = fieldBase
				pos++
				continue
			}
			out.Realm = append(out.Realm, v[pos])

		case fieldAlgorithm:
			if v[pos] == ' ' || v[pos] == ',' {
				state = fieldBase
				pos++
				continue
			}
			out.Algorithm = append(out.Algorithm, v[pos])

		case fieldUser:
			if v[pos] == '"' {
				state = fieldBase
				pos++
				continue
			}
			out.Username = append(out.Username, v[pos])

		case fieldNC:
			if v[pos] == ',' || v[pos] == ' ' {
				state = fieldBase
				pos++
				continue
			}
			out.Nc = append(out.Nc, v[pos])

		case fieldIgnore:
			if v[pos] == ' ' || v[pos] == ',' {
				state = fieldBase
				pos++
				continue
			}

		}
		pos++
	}
}
