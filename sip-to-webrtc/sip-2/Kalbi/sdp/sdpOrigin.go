package sdp

/*
   Author - Aaron Parfitt

   RFC 4566 - https://tools.ietf.org/html/rfc4566#section-5.2

   Origin ("o=")

   o=<username> <sess-id> <sess-version> <nettype> <addrtype>
     <unicast-address>

   The "o=" field gives the originator of the session (her username and
   the address of the user's host) plus a session identifier and version
   number:

   <username> is the user's login on the originating host, or it is "-"
      if the originating host does not support the concept of user IDs.
      The <username> MUST NOT contain spaces.

   <sess-id> is a numeric string such that the tuple of <username>,
      <sess-id>, <nettype>, <addrtype>, and <unicast-address> forms a
      globally unique identifier for the session.  The method of
      <sess-id> allocation is up to the creating tool, but it has been
      suggested that a Network Time Protocol (NTP) format timestamp be
      used to ensure uniqueness [13].

   <sess-version> is a version number for this session description.  Its
      usage is up to the creating tool, so long as <sess-version> is
      increased when a modification is made to the session data.  Again,
      it is RECOMMENDED that an NTP format timestamp is used.

   <nettype> is a text string giving the type of network.  Initially
      "IN" is defined to have the meaning "Internet", but other values
      MAY be registered in the future (see Section 8).

   <addrtype> is a text string giving the type of the address that
      follows.  Initially "IP4" and "IP6" are defined, but other values
      MAY be registered in the future (see Section 8).

   <unicast-address> is the address of the machine from which the
      session was created.  For an address type of IP4, this is either
      the fully qualified domain name of the machine or the dotted-
      decimal representation of the IP version 4 address of the machine.
      For an address type of IP6, this is either the fully qualified
      domain name of the machine or the compressed textual
      representation of the IP version 6 address of the machine.  For
      both IP4 and IP6, the fully qualified domain name is the form that
      SHOULD be given unless this is unavailable, in which case the
      globally unique address MAY be substituted.  A local IP address
      MUST NOT be used in any context where the SDP description might
      leave the scope in which the address is meaningful (for example, a
      local address MUST NOT be included in an application-level
      referral that might leave the scope).


      In general, the "o=" field serves as a globally unique identifier for
      this version of this session description, and the subfields excepting
      the version taken together identify the session irrespective of any
      modifications.

      For privacy reasons, it is sometimes desirable to obfuscate the
      username and IP address of the session originator.  If this is a
      concern, an arbitrary <username> and private <unicast-address> MAY be
      chosen to populate the "o=" field, provided that these are selected
      in a manner that does not affect the global uniqueness of the field.
*/

//SdpOrigin SDP origin header "o="
type SdpOrigin struct {
	Username       []byte
	SessionId      []byte
	SessionVersion []byte
	NetType        []byte
	AddrType       []byte
	UniAddr        []byte
	Src            []byte
}

//String returns header as string
func (so *SdpOrigin) String() string {
	line := "o="
	line += string(so.Username) + " "
	line += string(so.SessionId) + " "
	line += string(so.SessionVersion) + " "
	line += string(so.NetType) + " "
	line += string(so.AddrType) + " "
	line += string(so.UniAddr)
	return line
}

//ParseSdpOrigin parses SDP origin header
func ParseSdpOrigin(v []byte, out *SdpOrigin) {
	pos := 0
	state := fieldUsername

	// Init the output area
	out.Username = nil
	out.SessionId = nil
	out.SessionVersion = nil
	out.NetType = nil
	out.AddrType = nil
	out.UniAddr = nil
	out.Src = nil

	// Keep the source line if needed
	if keepSrc {
		out.Src = v
	}

	// Loop through the bytes making up the line
	for pos < len(v) {
		// FSM
		switch state {
		case fieldUsername:
			if v[pos] == ' ' {
				state = fieldSessionID
				pos++
				continue
			}
			out.Username = append(out.Username, v[pos])
		case fieldSessionID:
			if v[pos] == ' ' {
				state = fieldSessionVersion
				pos++
				continue
			}
			out.SessionId = append(out.SessionId, v[pos])

		case fieldSessionVersion:
			if v[pos] == ' ' {
				state = fieldNetType
				pos++
				continue
			}
			out.SessionVersion = append(out.SessionVersion, v[pos])

		case fieldNetType:
			if v[pos] == ' ' {
				state = fieldAddrType
				pos++
				continue
			}
			out.NetType = append(out.NetType, v[pos])
		case fieldAddrType:
			if v[pos] == ' ' {
				state = fieldUniAddr
				pos++
				continue
			}
			out.AddrType = append(out.AddrType, v[pos])
		case fieldUniAddr:
			if v[pos] == ' ' {
				state = fieldBase
				pos++
				continue
			}
			out.UniAddr = append(out.UniAddr, v[pos])
		}
		pos++
	}
}
