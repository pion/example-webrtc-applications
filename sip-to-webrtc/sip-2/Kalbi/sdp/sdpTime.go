package sdp

/*Author - Aaron Parfitt

RFC 4566 - https://tools.ietf.org/html/rfc4566#section-5.9

Timing ("t=")

      t=<start-time> <stop-time>

   The "t=" lines specify the start and stop times for a session.
   Multiple "t=" lines MAY be used if a session is active at multiple
   irregularly spaced times; each additional "t=" line specifies an
   additional period of time for which the session will be active.  If
   the session is active at regular times, an "r=" line (see below)
   should be used in addition to, and following, a "t=" line -- in which
   case the "t=" line specifies the start and stop times of the repeat
   sequence.

   The first and second sub-fields give the start and stop times,
   respectively, for the session.  These values are the decimal
   representation of Network Time Protocol (NTP) time values in seconds
   since 1900 [13].  To convert these values to UNIX time, subtract
   decimal 2208988800.

   NTP timestamps are elsewhere represented by 64-bit values, which wrap
   sometime in the year 2036.  Since SDP uses an arbitrary length
   decimal representation, this should not cause an issue (SDP
   timestamps MUST continue counting seconds since 1900, NTP will use
   the value modulo the 64-bit limit).

   If the <stop-time> is set to zero, then the session is not bounded,
   though it will not become active until after the <start-time>.  If
   the <start-time> is also zero, the session is regarded as permanent.

   User interfaces SHOULD strongly discourage the creation of unbounded
   and permanent sessions as they give no information about when the
   session is actually going to terminate, and so make scheduling
   difficult.

   The general assumption may be made, when displaying unbounded
   sessions that have not timed out to the user, that an unbounded
   session will only be active until half an hour from the current time
   or the session start time, whichever is the later.  If behaviour
   other than this is required, an end-time SHOULD be given and modified
   as appropriate when new information becomes available about when the
   session should really end.

   Permanent sessions may be shown to the user as never being active
   unless there are associated repeat times that state precisely when
   the session will be active.

*/

type sdpTime struct {
	TimeStart []byte
	TimeStop  []byte
	Src       []byte
}

//Export returns object as string
func (st *sdpTime) String() string {
	line := "t="
	line += string(st.TimeStart) + " "
	line += string(st.TimeStop)
	return line
}

//ParserSdpTime parses SDP time header
func ParserSdpTime(v []byte, out *sdpTime) {
	pos := 0
	state := fieldTimeStart

	// Init the output area
	out.TimeStart = nil
	out.TimeStop = nil
	out.Src = nil

	// Keep the source line if needed
	if keepSrc {
		out.Src = v
	}

	// Loop through the bytes making up the line
	for pos < len(v) {
		switch state {

		case fieldTimeStart:
			if v[pos] == ' ' {
				state = fieldTimeStop
				pos++
				continue
			}
			out.TimeStart = append(out.TimeStart, v[pos])

		case fieldTimeStop:
			out.TimeStop = append(out.TimeStop, v[pos])
		}
		pos++
	}

}
