package message

/*
 RFC 3261 - https://www.ietf.org/rfc/rfc3261.txt - 8.1.1.5 CSeq

  The CSeq header field serves as a way to identify and order
  transactions.  It consists of a sequence number and a method.  The
  method MUST match that of the request.

  Example:

     CSeq: 4711 INVITE

*/

//SipCseq SIP CSeq Header
type SipCseq struct {
	Id     []byte // Cseq ID
	Method []byte // Cseq Method
	Src    []byte // Full source if needed
}

//SetID sets CSeq ID
func (sc *SipCseq) SetID(id string) {
	sc.Id = []byte(id)
}

//SetMethod sets method
func (sc *SipCseq) SetMethod(method string) {
	sc.Method = []byte(method)
}

//String returns Header as String
func (sc *SipCseq) String() string {
	line := "Cseq: "
	line += string(sc.Id) + " " + string(sc.Method)
	return line
}

//ParseSipCseq parses CSeq Header
func ParseSipCseq(v []byte, out *SipCseq) {
	pos := 0
	state := fieldID

	// Init the output area
	out.Id = nil
	out.Method = nil
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
		case fieldID:
			if v[pos] == ' ' {
				state = fieldMethod
				pos++
				continue
			}
			out.Id = append(out.Id, v[pos])

		case fieldMethod:
			out.Method = append(out.Method, v[pos])
		}
		pos++
	}
}
