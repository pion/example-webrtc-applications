// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package softphone

import (
	"encoding/xml"
	"fmt"
	"log"
	"strings"
)

// OpenToInvite adds a handler that responds to any incoming invites.
func (softphone *Softphone) OpenToInvite() {
	softphone.inviteKey = softphone.addMessageListener(func(message string) {
		if strings.HasPrefix(message, "INVITE sip:") {
			inviteMessage := SIPMessage{}.FromString(message)

			dict := map[string]string{"Contact": fmt.Sprintf(`<sip:%s;transport=ws>`, softphone.fakeDomain)}
			responseMsg := inviteMessage.Response(*softphone, 180, dict, "")
			softphone.response(responseMsg)

			var msg Msg
			if err := xml.Unmarshal([]byte(inviteMessage.headers["P-rc"]), &msg); err != nil {
				log.Panic(err) // nolint
			}
			sipMessage := SIPMessage{}
			sipMessage.method = "MESSAGE"
			sipMessage.address = msg.Hdr.From
			sipMessage.headers = make(map[string]string)
			sipMessage.headers["Via"] = fmt.Sprintf("SIP/2.0/WSS %s;branch=%s", softphone.fakeDomain, branch())
			sipMessage.headers["From"] = fmt.Sprintf("<sip:%s@%s>;tag=%s", softphone.sipInfo.Username, softphone.sipInfo.Domain, softphone.fromTag) // nolint
			sipMessage.headers["To"] = fmt.Sprintf("<sip:%s>", msg.Hdr.From)
			sipMessage.headers["Content-Type"] = "x-rc/agent"
			sipMessage.addCseq(softphone).addCallID(*softphone).addUserAgent()
			sipMessage.Body = fmt.Sprintf(`<Msg><Hdr SID="%s" Req="%s" From="%s" To="%s" Cmd="17"/><Bdy Cln="%s"/></Msg>`, msg.Hdr.SID, msg.Hdr.Req, msg.Hdr.To, msg.Hdr.From, softphone.sipInfo.AuthorizationID) // nolint
			softphone.request(sipMessage, nil)

			softphone.OnInvite(inviteMessage)
		}
	})
}

// CloseToInvite removes the previously set invite listener.
func (softphone *Softphone) CloseToInvite() {
	softphone.removeMessageListener(softphone.inviteKey)
}

// OnOK adds a handler that responds to any incoming ok events.
func (softphone *Softphone) OnOK(hdlr func(string)) {
	softphone.addMessageListener(func(message string) {
		if strings.HasPrefix(message, "SIP/2.0 200 OK") {
			parsed := SIPMessage{}.FromString(message)
			hdlr(parsed.Body)
		}
	})
}
