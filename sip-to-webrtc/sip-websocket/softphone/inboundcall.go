package softphone

import (
	"encoding/xml"
	"fmt"
	"strings"
)

func (softphone *Softphone) OpenToInvite() {
	softphone.inviteKey = softphone.addMessageListener(func(message string) {
		if strings.HasPrefix(message, "INVITE sip:") {
			inviteMessage := SipMessage{}.FromString(message)

			dict := map[string]string{"Contact": fmt.Sprintf(`<sip:%s;transport=ws>`, softphone.fakeDomain)}
			responseMsg := inviteMessage.Response(*softphone, 180, dict, "")
			softphone.Response(responseMsg)

			var msg Msg
			xml.Unmarshal([]byte(inviteMessage.headers["P-rc"]), &msg)
			sipMessage := SipMessage{}
			sipMessage.method = "MESSAGE"
			sipMessage.address = msg.Hdr.From
			sipMessage.headers = make(map[string]string)
			sipMessage.headers["Via"] = fmt.Sprintf("SIP/2.0/WSS %s;branch=%s", softphone.fakeDomain, branch())
			sipMessage.headers["From"] = fmt.Sprintf("<sip:%s@%s>;tag=%s", softphone.sipInfo.Username, softphone.sipInfo.Domain, softphone.fromTag)
			sipMessage.headers["To"] = fmt.Sprintf("<sip:%s>", msg.Hdr.From)
			sipMessage.headers["Content-Type"] = "x-rc/agent"
			sipMessage.addCseq(softphone).addCallId(*softphone).addUserAgent()
			sipMessage.Body = fmt.Sprintf(`<Msg><Hdr SID="%s" Req="%s" From="%s" To="%s" Cmd="17"/><Bdy Cln="%s"/></Msg>`, msg.Hdr.SID, msg.Hdr.Req, msg.Hdr.To, msg.Hdr.From, softphone.sipInfo.AuthorizationId)
			softphone.request(sipMessage, nil)

			softphone.OnInvite(inviteMessage)
		}
	})
}

func (softphone *Softphone) CloseToInvite() {
	softphone.removeMessageListener(softphone.inviteKey)
}
