package sdp

/*Author Aaron Parfitt
RFC 4566 - https://tools.ietf.org/html/rfc4566#section-5.1

Protocol Version ("v=")

      v=0

   The "v=" field gives the version of the Session Description Protocol.
   This memo defines version 0.  There is no minor version number.

*/

type sdpVersion struct {
	Val []byte // Version number
	Src []byte // Full source if needed
}

func (sv *sdpVersion) String() string {
	line := "v="
	line += string(sv.Val)
	return line
}
