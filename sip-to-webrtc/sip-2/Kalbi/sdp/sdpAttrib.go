package sdp

/*
  RFC4566 - https://tools.ietf.org/html/rfc4566#section-5.14

  6.  SDP Attributes

 The following attributes are defined.  Since application writers may
 add new attributes as they are required, this list is not exhaustive.
 Registration procedures for new attributes are defined in Section
 8.2.4.

    a=cat:<category>

       This attribute gives the dot-separated hierarchical category of
       the session.  This is to enable a receiver to filter unwanted
       sessions by category.  There is no central registry of
       categories.  It is a session-level attribute, and it is not
       dependent on charset.

 eg:
 a=ptime:20

*/

type sdpAttrib struct {
	Cat []byte // Named portion of URI
	Val []byte // Port number
	Src []byte // Full source if needed
}

func (sa *sdpAttrib) String() string {
	line := "a="
	line += string(sa.Cat)
	if sa.Val != nil {
		line += ":" + string(sa.Val)
	}
	return line
}

func parseSdpAttrib(v []byte, out *sdpAttrib) {
	pos := 0
	state := fieldCat

	// Init the output area
	out.Cat = nil
	out.Val = nil
	out.Src = nil

	// Keep the source line if needed
	if keepSrc {
		out.Src = v
	}

	// Loop through the bytes making up the line
	for pos < len(v) {
		// FSM
		switch state {
		case fieldCat:
			if v[pos] == ':' {
				state = fieldValue
				pos++
				continue
			}
			out.Cat = append(out.Cat, v[pos])

		case fieldValue:
			out.Val = append(out.Val, v[pos])
		}
		pos++
	}
}
