package main

import (
	"fmt"
	"sip"
	"softphone"
)

func main() {
	//EXEMPLE AND LIB FROM : https://github.com/ringcentral/ringcentral-softphone-go
	//LIB WAS  ADAPTED AND CORRECTED FOR FREESWITCH !
	sipInfo := softphone.SIPInfoResponse{Username: "100", Password: "100", Domain: "192.168.1.10", Transport: "ws", OutboundProxy: "192.168.1.10:5066"}
	sp := softphone.NewSoftPhone(sipInfo)
	sp.OnInvite = func(inviteMessage softphone.SipMessage) { // NEW INCOMMING CALL
		answerSDPChan := make(chan string)
		sip.Answer(inviteMessage.Body, answerSDPChan) // BUILD  ANSWER WITH PION !
		dict := map[string]string{
			"Contact":      fmt.Sprintf("<sip:%s;transport=ws>", sp.FakeEmail),
			"Content-Type": "application/sdp",
		}
		answerSDP := <-answerSDPChan
		responseMsg := inviteMessage.Response(*sp, 200, dict, answerSDP)
		sp.Response(responseMsg)
	}
	sp.OpenToInvite()
	select {}
}
