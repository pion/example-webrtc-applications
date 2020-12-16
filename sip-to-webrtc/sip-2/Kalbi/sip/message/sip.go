package message

import (
	"bytes"
	"strconv"
	"strings"

	"Kalbi/sdp"
)

var sipType = 0
var keepSrc = true

//SipMsg is a representation of a SIP message
type SipMsg struct {
	Req      SipReq
	From     SipFrom
	To       SipTo
	Contact  SipContact
	Via      []SipVia
	Cseq     SipCseq
	Auth     SipAuth
	Ua       SipVal
	Exp      SipVal
	MaxFwd   SipVal
	CallId   SipVal
	ContType SipVal
	ContLen  SipVal
	Src      []byte
	Body     []byte
	Sdp      sdp.SdpMsg
}

//SetAuthHeader sets auth header
func (sm *SipMsg) SetAuthHeader(auth *SipAuth) {
	sm.Auth = *auth
}

//GetStatusCode returns responses status code
func (sm *SipMsg) GetStatusCode() int {
	code, err := strconv.Atoi(string(sm.Req.StatusCode))
	if err != nil {
		return 0
	}
	return code
}

//CopyHeaders copys headers from one SIP message to another
func (sm *SipMsg) CopyHeaders(msg *SipMsg) {
	sm.Via = msg.Via
	sm.From = msg.From
	sm.To = msg.To
	sm.Contact = msg.Contact
	sm.CallId = msg.CallId
	sm.ContType = msg.ContType
	sm.Cseq = msg.Cseq
	sm.MaxFwd = msg.MaxFwd
	sm.ContLen = msg.ContLen
}

//CopySdp copys SDP from one SIP message to another
func (sm *SipMsg) CopySdp(msg *SipMsg) {
	sm.Sdp = msg.Sdp
}

//Export returns SIP message as string
func (sm *SipMsg) String() string {
	sipmsg := ""
	sipmsg += sm.Req.String() + "\r\n"
	sipmsg += sm.Via[0].String() + "\r\n"
	sipmsg += sm.From.String() + "\r\n"
	sipmsg += sm.To.String() + "\r\n"
	sipmsg += sm.Contact.String() + "\r\n"
	sipmsg += sm.Cseq.String() + "\r\n"
	if sm.ContType.Value != nil {
		sipmsg += "Content-Type: " + sm.ContType.String() + "\r\n"
	}
	if sm.Auth.Response != nil {
        sipmsg += "Authorization: " + sm.Auth.String() + "\r\n"
	}else if sm.Auth.Nonce != nil{
		sipmsg += "WWW-Authenticate: " + sm.Auth.String() + "\r\n"
	}
	sipmsg += "Call-ID: " + sm.CallId.String() + "\r\n"
	sipmsg += "Max-Forwards: " + sm.MaxFwd.String() + "\r\n"
	sipmsg += "Content-Length: " + sm.ContLen.String() + "\r\n"
	sipmsg += "\r\n"

	if sm.Body != nil {
		sipmsg += string(sm.Body)
	}

	return sipmsg
}




//SipVal is the value of a simple SIP Header e.g. Max Forwards
type SipVal struct {
	Value []byte // Sip Value
	Src   []byte // Full source if needed
}

//SetValue sets the value of a simple SIP Header e.g. Max Forwards
func (sv *SipVal) SetValue(value string) {
	sv.Value = []byte(value)
}

//Export returns SIP value as string
func (sv *SipVal) String() string {
	return string(sv.Value)
}

// Parse routine, passes by value
func Parse(v []byte) (output SipMsg) {
	output.Src = v
	// Allow multiple vias and media Attribs
	viaIdx := 0
	output.Via = make([]SipVia, 0, 8)

	// Split SIP & Body
	bodysplit := bytes.Split(v, []byte("\r\n\r\n"))
	if len(bodysplit) > 1 {
		output.Body = bodysplit[1]
	}

	lines := bytes.Split(v, []byte("\r\n"))

	for i, line := range lines {
		//fmt.Println(i, string(line))
		line = bytes.TrimSpace(line)
		if i == 0 {
			// For the first line parse the request
			ParseSipReq(line, &output.Req)
		} else {
			// For subsequent lines split in sep (: for sip, = for sdp)
			spos, stype := indexSep(line)
			if spos > 0 && stype == ':' {
				// SIP: Break up into header and value
				lhdr := strings.ToLower(string(line[0:spos]))
				lval := bytes.TrimSpace(line[spos+1:])

				// Switch on the line header
				//fmt.Println(i, string(lhdr), string(lval))
				switch {
				case lhdr == "f" || lhdr == "from":
					ParseSipFrom(lval, &output.From)
				case lhdr == "t" || lhdr == "to":
					ParseSipTo(lval, &output.To)
				case lhdr == "m" || lhdr == "contact":
					ParseSipContact(lval, &output.Contact)
				case lhdr == "v" || lhdr == "via":
					var tmpVia SipVia
					output.Via = append(output.Via, tmpVia)
					ParseSipVia(lval, &output.Via[viaIdx])
					viaIdx++
				case lhdr == "i" || lhdr == "call-id":
					output.CallId.Value = lval
				case lhdr == "c" || lhdr == "content-type":
					output.ContType.Value = lval
				case lhdr == "content-length":
					output.ContLen.Value = lval
				case lhdr == "user-agent":
					output.Ua.Value = lval
				case lhdr == "expires":
					output.Exp.Value = lval
				case lhdr == "max-forwards":
					output.MaxFwd.Value = lval
				case lhdr == "cseq":
					ParseSipCseq(lval, &output.Cseq)
				case lhdr == "authorization" || lhdr == "www-authenticate":
					ParseSipAuth(lval, &output.Auth)
				} // End of Switch
			}

		}
	}

	return
}

// Finds the first valid Separate or notes its type
func indexSep(s []byte) (int, byte) {

	for i := 0; i < len(s); i++ {
		if s[i] == ':' {
			return i, ':'
		}
		if s[i] == '=' {
			return i, '='
		}
	}
	return -1, ' '
}

// Get a string from a slice of bytes
// Checks the bounds to avoid any range errors
func getString(sl []byte, from, to int) string {
	// Remove negative values
	if from < 0 {
		from = 0
	}
	if to < 0 {
		to = 0
	}
	// Limit if over len
	if from > len(sl) || from > to {
		return ""
	}
	if to > len(sl) {
		return string(sl[from:])
	}
	return string(sl[from:to])
}

// Get a slice from a slice of bytes
// Checks the bounds to avoid any range errors
func getBytes(sl []byte, from, to int) []byte {
	// Remove negative values
	if from < 0 {
		from = 0
	}
	if to < 0 {
		to = 0
	}
	// Limit if over len
	if from > len(sl) || from > to {
		return nil
	}
	if to > len(sl) {
		return sl[from:]
	}
	return sl[from:to]
}

// MessageDetails prints all we know about the struct in a readable format
func MessageDetails(data *SipMsg) string {
	msg := "- SIP --------------------------------\n\n"
	msg += "[REQ]\n"
	msg += "\t[UriType] => " + string(data.Req.UriType) + "\n"
	msg += "\t[Method] =>" + string(data.Req.Method) + "\n"
	msg += "\t[StatusCode] =>" + string(data.Req.StatusCode) + "\n"
	msg += "\t[User] =>" + string(data.Req.User) + "\n"
	msg += "\t[Host] =>" + string(data.Req.Host) + "\n"
	msg += "\t[Port] =>" + string(data.Req.Port) + "\n"
	msg += "\t[UserType] =>" + string(data.Req.UserType) + "\n"
	msg += "\t[Src] =>" + string(data.Req.Src) + "\n"

	// FROM
	msg += "[FROM]" + "\n"
	msg += "\t[UriType] =>" + data.From.UriType + "\n"
	msg += "\t[Name] =>" + string(data.From.Name) + "\n"
	msg += "\t[User] =>" + string(data.From.User) + "\n"
	msg += "\t[Host] =>" + string(data.From.Host) + "\n"
	msg += "\t[Port] =>" + string(data.From.Port) + "\n"
	msg += "\t[Tag] =>" + string(data.From.Tag) + "\n"
	msg += "\t[Src] =>" + string(data.From.Src) + "\n"
	// TO
	msg += "[TO]" + "\n"
	msg += "\t[UriType] =>" + data.To.UriType + "\n"
	msg += "\t[Name] =>" + string(data.To.Name) + "\n"
	msg += "\t[User] =>" + string(data.To.User) + "\n"
	msg += "\t[Host] =>" + string(data.To.Host) + "\n"
	msg += "\t[Port] =>" + string(data.To.Port) + "\n"
	msg += "\t[Tag] =>" + string(data.To.Tag) + "\n"
	msg += "\t[UserType] =>" + string(data.To.UserType) + "\n"
	msg += "\t[Src] =>" + string(data.To.Src) + "\n"
	// TO
	msg += "[Contact]" + "\n"
	msg += "\t[UriType] =>" + data.Contact.UriType + "\n"
	msg += "\t[Name] =>" + string(data.Contact.Name) + "\n"
	msg += "\t[User] =>" + string(data.Contact.User) + "\n"
	msg += "\t[Host] =>" + string(data.Contact.Host) + "\n"
	msg += "\t[Port] =>" + string(data.Contact.Port) + "\n"
	msg += "\t[Transport] =>" + string(data.Contact.Tran) + "\n"
	msg += "\t[Q] =>" + string(data.Contact.Qval) + "\n"
	msg += "\t[Expires] =>" + string(data.Contact.Expires) + "\n"
	msg += "\t[Src] =>" + string(data.Contact.Src) + "\n"
	// UA
	/*
		fmt.Println("  [Cseq]")
		fmt.Println("    [Id] =>", string(data.Cseq.Id))
		fmt.Println("    [Method] =>", string(data.Cseq.Method))
		fmt.Println("    [Src] =>", string(data.Cseq.Src))
		// UA
		fmt.Println("  [User Agent]")
		fmt.Println("    [Value] =>", string(data.Ua.Value))
		fmt.Println("    [Src] =>", string(data.Ua.Src))
		// Exp
		fmt.Println("  [Expires]")
		fmt.Println("    [Value] =>", string(data.Exp.Value))
		fmt.Println("    [Src] =>", string(data.Exp.Src))
		// MaxFwd
		fmt.Println("  [Max Forwards]")
		fmt.Println("    [Value] =>", string(data.MaxFwd.Value))
		fmt.Println("    [Src] =>", string(data.MaxFwd.Src))
		// CallId
		fmt.Println("  [Call-ID]")
		fmt.Println("    [Value] =>", string(data.CallId.Value))
		fmt.Println("    [Src] =>", string(data.CallId.Src))
		// Content-Type
		fmt.Println("  [Content-Type]")
		fmt.Println("    [Value] =>", string(data.ContType.Value))
		fmt.Println("    [Src] =>", string(data.ContType.Src))

		// Via - Multiple
		fmt.Println("  [Via]")
		for i, via := range data.Via {
			fmt.Println("    [", i, "]")
			fmt.Println("      [Tansport] =>", via.Trans)
			fmt.Println("      [Host] =>", string(via.Host))
			fmt.Println("      [Port] =>", string(via.Port))
			fmt.Println("      [Branch] =>", string(via.Branch))
			fmt.Println("      [Rport] =>", string(via.Rport))
			fmt.Println("      [Maddr] =>", string(via.Maddr))
			fmt.Println("      [ttl] =>", string(via.Ttl))
			fmt.Println("      [Recevied] =>", string(via.Rcvd))
			fmt.Println("      [Src] =>", string(via.Src))
		}

		fmt.Println("-SDP --------------------------------")
		// Media Desc
		fmt.Println("  [MediaDesc]")
		fmt.Println("    [MediaType] =>", string(data.Sdp.MediaDesc.MediaType))
		fmt.Println("    [Port] =>", string(data.Sdp.MediaDesc.Port))
		fmt.Println("    [Proto] =>", string(data.Sdp.MediaDesc.Proto))
		fmt.Println("    [Fmt] =>", string(data.Sdp.MediaDesc.Fmt))
		fmt.Println("    [Src] =>", string(data.Sdp.MediaDesc.Src))
		// Connection Data
		fmt.Println("  [ConnData]")
		fmt.Println("    [AddrType] =>", string(data.Sdp.ConnData.AddrType))
		fmt.Println("    [ConnAddr] =>", string(data.Sdp.ConnData.ConnAddr))
		fmt.Println("    [Src] =>", string(data.Sdp.ConnData.Src))

		// Attribs - Multiple
		fmt.Println("  [Attrib]")
		for i, attr := range data.Sdp.Attrib {
			fmt.Println("    [", i, "]")
			fmt.Println("      [Cat] =>", string(attr.Cat))
			fmt.Println("      [Val] =>", string(attr.Val))
			fmt.Println("      [Src] =>", string(attr.Src))
		}*/
	msg += "-------------------------------------"
	return msg

}

const fieldNull = 0
const fieldBase = 1
const fieldValue = 2
const fieldName = 3
const fieldNameQ = 4
const fieldUser = 5
const fieldUserHost = 6
const fieldPort = 7
const fieldTag = 8
const fieldID = 9
const fieldMethod = 10
const fieldTran = 11
const fieldBranch = 12
const fieldRport = 13
const fieldMaddr = 14
const fieldTTL = 15
const fieldRec = 16
const fieldExpires = 17
const fieldQ = 18
const fieldUserType = 19
const fieldStatus = 20
const fieldStatusDesc = 21

// States for Auth Header
const fieldQop = 22
const fieldNonce = 23
const fieldRealm = 24
const fieldAlgorithm = 25
const fieldCNonce = 26
const fieldNC = 27
const fieldResponse = 28
const fieldURI = 29

//

const fieldAddrType = 40
const fieldConnAddr = 41
const fieldMedia = 42
const fieldProto = 43
const fieldFmt = 44
const fieldCat = 45

const fieldIgnore = 255
