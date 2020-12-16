package softphone


import (
	"fmt"
	"regexp"
	"strings"
)

type SipMessage struct {
	method  string
	address string
	subject string
	headers map[string]string
	Body    string
}

func (sm *SipMessage) addAuthorization(softphone Softphone, nonce string) *SipMessage {
	sm.headers["Authorization"] = generateAuthorization(softphone.sipInfo, "REGISTER", nonce)
	return sm
}

func (sm *SipMessage) newViaBranch() *SipMessage {
	if val, ok := sm.headers["Via"]; ok {
		sm.headers["Via"] = regexp.MustCompile(";branch=z9hG4bK.+?$").ReplaceAllString(val, ";branch="+branch())
	}
	return sm
}

func (sm *SipMessage) addCseq(softphone *Softphone) *SipMessage {
	sm.headers["CSeq"] = fmt.Sprintf("%d %s", softphone.cseq, sm.method)
	softphone.cseq += 1
	return sm
}

func (sm *SipMessage) addCallId(softphone Softphone) *SipMessage {
	sm.headers["Call-ID"] = softphone.callId
	return sm
}

func (sm *SipMessage) addUserAgent() *SipMessage {
	sm.headers["User-Agent"] = "ACE MEDIAS TOOLS / INES-SIP"
	return sm
}

func (sm SipMessage) ToString() string {
	arr := []string{fmt.Sprintf("%s sip:%s SIP/2.0", sm.method, sm.address)}
	for k, v := range sm.headers {
		arr = append(arr, fmt.Sprintf("%s: %s", k, v))
	}
	arr = append(arr, fmt.Sprintf("Content-Length: %d", len(sm.Body)))
	arr = append(arr, "Max-Forwards: 70")
	arr = append(arr, "", sm.Body)
	return strings.Join(arr, "\r\n")
}

func (sm SipMessage) FromString(s string) SipMessage {
	parts := strings.Split(s, "\r\n\r\n")
	sm.Body = strings.Join(parts[1:], "\r\n\r\n")
	parts = strings.Split(parts[0], "\r\n")
	sm.subject = parts[0]
	sm.headers = make(map[string]string)
	for _, line := range parts[1:] {
		tokens := strings.Split(line, ": ")
		sm.headers[tokens[0]] = tokens[1]
	}
	return sm
}

func (sm SipMessage) Response(softphone Softphone, statusCode int, headers map[string]string, body string) string {
	arr := []string{fmt.Sprintf("SIP/2.0 %d %s", statusCode, ResponseCodes[statusCode])}
	for _, key := range []string{"Via", "From", "Call-ID", "CSeq"} {
		arr = append(arr, fmt.Sprintf("%s: %s", key, sm.headers[key]))
	}
	for k, v := range headers {
		arr = append(arr, fmt.Sprintf("%s: %s", k, v))
	}
	arr = append(arr, "Supported: outbound", fmt.Sprintf("To: %s;tag=%s", sm.headers["To"], softphone.toTag))
	arr = append(arr, fmt.Sprintf("Content-Length: %d", len(body)))
	arr = append(arr, "", body)
	return strings.Join(arr, "\r\n")
}
