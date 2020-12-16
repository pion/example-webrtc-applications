package sdp

import (
	"bytes"
	"strings"
)

var keepSrc = true

//FMS States
const fieldNull = 0
const fieldBase = 1
const fieldValue = 2
const fieldPort = 3
const fieldAddrType = 4
const fieldConnAddr = 5
const fieldMedia = 6
const fieldProto = 7
const fieldFmt = 8
const fieldCat = 9
const fieldUsername = 10
const fieldSessionID = 11
const fieldSessionVersion = 12
const fieldNetType = 13
const fieldUniAddr = 14
const fieldTimeStart = 15
const fieldTimeStop = 16

const fieldIgnore = 255

//SdpMsg is representation of an SDP message
type SdpMsg struct {
	Origin    SdpOrigin
	Version   sdpVersion
	Time      sdpTime
	MediaDesc sdpMediaDesc
	Attrib    []sdpAttrib
	ConnData  sdpConnData
}

//Size returns size in bytes
func (sm *SdpMsg) Size() int {
	sdp := sm.String()
	return len([]byte(sdp))
}

func (sm *SdpMsg) String() string {
	sdp := ""
	sdp += strings.TrimSpace(sm.Version.String()) + "\r\n"
	sdp += strings.TrimSpace(sm.Origin.String()) + "\r\n"
	sdp += "s=" + strings.TrimSpace(string(sm.Origin.Username)) + "\r\n"
	sdp += strings.TrimSpace(sm.ConnData.String()) + "\r\n"
	sdp += strings.TrimSpace(sm.Time.String()) + "\r\n"
	sdp += strings.TrimSpace(sm.MediaDesc.String()) + "\r\n"
	for _, a := range sm.Attrib {
		sdp += strings.TrimSpace(a.String()) + "\r\n"
	}
	return sdp

}

func indexSep(s []byte) (int, byte) {
	for i := 0; i < len(s); i++ {
		if s[i] == '=' {
			return i, '='
		}
	}
	return -1, ' '
}

//Parse Parses SDP
func Parse(v []byte) (output SdpMsg) {
	attrIdx := 0
	output.Attrib = make([]sdpAttrib, 0, 8)

	lines := bytes.Split(v, []byte("\r\n"))

	for _, line := range lines {
		spos, stype := indexSep(line)
		//fmt.Println(i, string(line))
		line = bytes.TrimSpace(line)

		if spos == 1 && stype == '=' {
			// SDP: Break up into header and value
			lhdr := strings.ToLower(string(line[0]))
			lval := bytes.TrimSpace(line[2:])
			// Switch on the line header
			//fmt.Println(i, spos, string(lhdr), string(lval))
			switch {
			case lhdr == "v":
				output.Version = sdpVersion{Val: lval, Src: line}
			case lhdr == "o":
				ParseSdpOrigin(lval, &output.Origin)
			case lhdr == "t":
				ParserSdpTime(lval, &output.Time)
			case lhdr == "m":
				parseSdpMediaDesc(lval, &output.MediaDesc)
			case lhdr == "c":
				parseSdpConnectionData(lval, &output.ConnData)
			case lhdr == "a":
				var tmpAttrib sdpAttrib
				output.Attrib = append(output.Attrib, tmpAttrib)
				parseSdpAttrib(lval, &output.Attrib[attrIdx])
				attrIdx++

			} // End of Switch

		}
	}
	return
}
