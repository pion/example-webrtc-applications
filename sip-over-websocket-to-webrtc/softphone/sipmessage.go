// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package softphone

import (
	"fmt"
	"regexp"
	"strings"
)

// SIPMessage ...
type SIPMessage struct {
	method  string
	address string
	subject string
	headers map[string]string
	Body    string
}

func (sm *SIPMessage) addAuthorization(softphone Softphone, nonce, method string) *SIPMessage {
	sm.headers["Authorization"] = generateAuthorization(softphone.sipInfo, method, nonce)

	return sm
}

func (sm *SIPMessage) addProxyAuthorization(softphone Softphone, nonce, user, method string) *SIPMessage {
	sm.headers["Proxy-Authorization"] = generateProxyAuthorization(softphone.sipInfo, method, user, nonce)

	return sm
}

func (sm *SIPMessage) newViaBranch() {
	if val, ok := sm.headers["Via"]; ok {
		sm.headers["Via"] = regexp.MustCompile(";branch=z9hG4bK.+?$").ReplaceAllString(val, ";branch="+branch())
	}
}

func (sm *SIPMessage) addCseq(softphone *Softphone) *SIPMessage {
	sm.headers["CSeq"] = fmt.Sprintf("%d %s", softphone.cseq, sm.method)
	softphone.cseq++

	return sm
}

func (sm *SIPMessage) addCallID(softphone Softphone) *SIPMessage {
	sm.headers["Call-ID"] = softphone.callID

	return sm
}

func (sm *SIPMessage) addUserAgent() {
	sm.headers["User-Agent"] = "Pion WebRTC SIP Client"
}

// ToString ...
func (sm SIPMessage) ToString() string {
	arr := []string{fmt.Sprintf("%s sip:%s SIP/2.0", sm.method, sm.address)}
	for k, v := range sm.headers {
		arr = append(arr, fmt.Sprintf("%s: %s", k, v))
	}

	arr = append(arr, fmt.Sprintf("Content-Length: %d", len(sm.Body)))
	arr = append(arr, "Max-Forwards: 70")
	arr = append(arr, "", sm.Body)

	return strings.Join(arr, "\r\n")
}

// FromString ...
func (sm SIPMessage) FromString(s string) SIPMessage {
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

// Response ...
func (sm SIPMessage) Response(softphone Softphone, statusCode int, headers map[string]string, body string) string {
	arr := []string{fmt.Sprintf("SIP/2.0 %d %s", statusCode, responseCodes[statusCode])}
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
