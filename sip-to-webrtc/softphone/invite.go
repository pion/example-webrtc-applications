package softphone

import (
	"fmt"
	"strings"
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
		proxyAuthenticateHeader := SIPMessage{}.FromString(message).headers["Proxy-Authenticate"]
		authenticateHeader := SIPMessage{}.FromString(message).headers["WWW-Authenticate"]

		var ai AuthInfo
		if len(authenticateHeader) > 0 { //WWW-Authenticate
			ai =  GetAuthInfo(authenticateHeader)
			ai.AuthType = "Authorization"
			ai.Uri = "sip:"+ extension + "@"+ softphone.sipInfo.Domain
			ai.Method = "INVITE"
		} else if len(proxyAuthenticateHeader) > 0 { //Proxy-Authenticate
			ai =  GetAuthInfo(proxyAuthenticateHeader)
			ai.AuthType = "Proxy-Authorization"
			ai.Uri = "sip:"+ extension + "@"+ softphone.sipInfo.Domain
			ai.Method = "INVITE"
		} else {
			panic("FAIL TO SEND INVITE")
		}
		fmt.Printf("%+v\n", ai)
		sipMessage.addAuthorization(*softphone, ai).addCseq(softphone).newViaBranch()
		softphone.request(sipMessage, func(msg string) bool {
			responseStatus := strings.Split(strings.Split(msg, "\r\n")[0], " ")[1]
			textStatus := strings.Split(strings.Split(msg, "\r\n")[0], " ")[2]
			switch responseStatus {
			case "100":
				fmt.Println(textStatus)
				return false //Continue the handler
			case "183":
				fmt.Println(textStatus)
				return false //Continue the handler

			case "200":
				fmt.Println("### INVITE SUCCESS ###")
				return true
			default: //480
				panic("INVITE FAILED : " +  responseStatus  + " " + textStatus)
			}
			return true
		})

		return true
	})
}
