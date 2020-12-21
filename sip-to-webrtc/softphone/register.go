package softphone

import (
	"crypto/tls"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"strings"
)

func (softphone *Softphone) register() {
	parsedUrl, err := url.Parse(softphone.sipInfo.WebsocketURL)
	if err != nil {
		panic(err)
	}
	dialer := websocket.DefaultDialer
	dialer.Subprotocols = []string{"sip"}
	dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint

	conn, _, err := dialer.Dial(parsedUrl.String(), nil)
	if err != nil {
		log.Fatal(err)
	}

	softphone.wsConn = conn

	go func() {
		for {
			_, bytes, err := conn.ReadMessage()
			if err != nil {
				log.Fatal(err)
			}

			message := string(bytes)
			//log.Print("↓↓↓\n", message)

			for _, ml := range softphone.messageListeners {
				go ml(message)
			}
		}
	}()
	sipMessage := SIPMessage{}
	sipMessage.method = "REGISTER"
	sipMessage.address = softphone.sipInfo.Domain
	sipMessage.headers = make(map[string]string)
	sipMessage.headers["Contact"] = fmt.Sprintf("<sip:%s;transport=ws>;expires=200", softphone.FakeEmail)
	sipMessage.headers["Via"] = fmt.Sprintf("SIP/2.0/WS %s;branch=%s", softphone.fakeDomain, branch())
	sipMessage.headers["From"] = fmt.Sprintf("<sip:%s@%s>;tag=%s", softphone.sipInfo.Username, softphone.sipInfo.Domain, softphone.fromTag)
	sipMessage.headers["To"] = fmt.Sprintf("<sip:%s@%s>", softphone.sipInfo.Username, softphone.sipInfo.Domain)
	sipMessage.headers["Organization"] = "Pion WebRTC SIP Client"
	sipMessage.headers["Supported"] = "path,ice"
	sipMessage.addCseq(softphone).addCallID(*softphone).addUserAgent()

	softphone.request(sipMessage, func(message string) bool {
		authenticateHeader := SIPMessage{}.FromString(message).headers["WWW-Authenticate"]
		ai := GetAuthInfo(authenticateHeader)
		ai.AuthType = "Authorization"
		ai.Uri = "sip:" + softphone.sipInfo.Domain
		ai.Method = "REGISTER"
		sipMessage.addAuthorization(*softphone, ai).addCseq(softphone).newViaBranch()
		softphone.request(sipMessage, func(msg string) bool {

			responseStatus := strings.Split(strings.Split(msg, "\r\n")[0], " ")[1]
			textStatus := strings.Split(strings.Split(msg, "\r\n")[0], " ")[2]
			switch responseStatus {
			case "200":
				fmt.Println("### SOFTPHONE CONNECTED ###")
				break
			default:
				panic("AUTHENTICATION FAILED : " +  responseStatus  + " " + textStatus)
				break
			}

			return true
		})

		return true
	})
}
