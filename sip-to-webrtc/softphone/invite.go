package softphone

import (
	"fmt"
)

// Invite ...
func (softphone *Softphone) Invite(extension, offer string) {
	sipMessage := SIPMessage{headers: map[string]string{}}

	sipMessage.method = "INVITE"
	sipMessage.address = softphone.sipInfo.Domain

	sipMessage.headers["Contact"] = fmt.Sprintf("<sip:%s;transport=ws>;expires=200", softphone.FakeEmail)
	sipMessage.headers["To"] = fmt.Sprintf("<sip:%s@%s>", extension, softphone.sipInfo.Domain)
	sipMessage.headers["Via"] = fmt.Sprintf("SIP/2.0/WS %s;branch=%s", softphone.fakeDomain, branch())
	sipMessage.headers["From"] = fmt.Sprintf("<sip:%s@%s>;tag=%s", softphone.sipInfo.Username, softphone.sipInfo.Domain, softphone.fromTag)
	sipMessage.headers["Supported"] = "replaces, outbound,ice"
	sipMessage.addCseq(softphone).addCallID(*softphone).addUserAgent()

	sipMessage.headers["Content-Type"] = "application/sdp"
	sipMessage.Body = offer

	softphone.request(sipMessage, func(message string) bool {
		authenticateHeader := SIPMessage{}.FromString(message).headers["Proxy-Authenticate"]
		ai :=  GetAuthInfo(authenticateHeader)
		ai.AuthType = "Proxy-Authorization"
		ai.Uri = "sip:"+ extension + "@"+ softphone.sipInfo.Domain
		ai.Method = "INVITE"
		sipMessage.addAuthorization(*softphone, ai).addCseq(softphone).newViaBranch()
		softphone.request(sipMessage, func(msg string) bool {
			return false
		})

		return true
	})
}
