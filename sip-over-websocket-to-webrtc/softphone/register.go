package softphone

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"

	"github.com/gorilla/websocket"
)

func (softphone *Softphone) register() {
	url := url.URL{Scheme: strings.ToLower(softphone.sipInfo.Transport), Host: softphone.sipInfo.OutboundProxy, Path: ""}
	dialer := websocket.DefaultDialer
	dialer.Subprotocols = []string{"sip"}
	dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint

	conn, _, err := dialer.Dial(url.String(), nil)
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
			// log.Print("↓↓↓\n", message)

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
	sipMessage.headers["Organization"] = "ACE MEDIAS TOOLS"
	sipMessage.headers["Supported"] = "path,ice"
	sipMessage.addCseq(softphone).addCallID(*softphone).addUserAgent()

	registered, registeredFunc := context.WithCancel(context.Background())

	softphone.request(sipMessage, func(message string) bool {
		authenticateHeader := SIPMessage{}.FromString(message).headers["WWW-Authenticate"]
		regex := regexp.MustCompile(`, nonce="(.+?)"`)
		nonce := regex.FindStringSubmatch(authenticateHeader)[1]

		sipMessage.addAuthorization(*softphone, nonce, "REGISTER").addCseq(softphone).newViaBranch()
		softphone.request(sipMessage, func(msg string) bool {
			registeredFunc()

			return false
		})

		return true
	})

	<-registered.Done()
}
