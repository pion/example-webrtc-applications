package softphone

import (
	"crypto/tls"

	"fmt"
	log "github.com/sirupsen/logrus"
	"net/url"
	"regexp"
	"strings"

	"github.com/gorilla/websocket"

)

func (softphone *Softphone) register() {
	//bytes := softphone.rc.Post("/restapi/v1.0/client-info/sip-provision", strings.NewReader(`{"sipInfo":[{"transport":"WSS"}]}`))
	//fmt.Printf("MY SIP INFO : %+v\n", softphone.sipInfo)
	url := url.URL{Scheme: strings.ToLower(softphone.sipInfo.Transport), Host: softphone.sipInfo.OutboundProxy, Path: ""}
	dialer := websocket.DefaultDialer
	dialer.Subprotocols = []string{"sip"}
	dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
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
			//fmt.Printf("%+v", message)
			log.Debug("↓↓↓\n", message)
			for _, ml := range softphone.messageListeners {
				go ml(message)
			}
		}
	}()
//Authorization: Digest username="101",realm="192.168.1.30",nonce="2a43702f-ef21-463c-95f1-565ac4f439bd",uri="sip:192.168.1.30",response="2f5a99ed666f28a3ed3afae7c289205d"",algorithm=MD5,cnonce="0e6758e1adfccffbd0ad9ffdde3ef655",qop=auth,nc=00000001
//Authorization: Digest username="102",realm="192.168.1.30",nonce="687f7a9d-8d53-477f-b23e-92bc59daa081",uri="sip:192.168.1.30",response="2a1e079b4ccd9d2abf6185ffa0eecf1c",algorithm=MD5,cnonce="dfe910a916adb292027e926280325a2c",qop=auth,nc=00000001
	sipMessage := SipMessage{}
	sipMessage.method = "REGISTER"
	sipMessage.address = softphone.sipInfo.Domain
	sipMessage.headers = make(map[string]string)
	sipMessage.headers["Contact"] = fmt.Sprintf("<sip:%s;transport=ws>;expires=200", softphone.FakeEmail)
	sipMessage.headers["Via"] = fmt.Sprintf("SIP/2.0/WS %s;branch=%s", softphone.fakeDomain, branch())
	sipMessage.headers["From"] = fmt.Sprintf("<sip:%s@%s>;tag=%s", softphone.sipInfo.Username, softphone.sipInfo.Domain, softphone.fromTag)
	sipMessage.headers["To"] = fmt.Sprintf("<sip:%s@%s>", softphone.sipInfo.Username, softphone.sipInfo.Domain)
	sipMessage.headers["Organization"] = fmt.Sprintf("ACE MEDIAS TOOLS")
	sipMessage.headers["Supported"] = fmt.Sprintf("path")
	sipMessage.addCseq(softphone).addCallId(*softphone).addUserAgent()
	softphone.request(sipMessage, func(message string) bool {

		if strings.Contains(message, "WWW-Authenticate: Digest") {
			authenticateHeader := SipMessage{}.FromString(message).headers["WWW-Authenticate"]
			regex := regexp.MustCompile(`, nonce="(.+?)"`)
			nonce := regex.FindStringSubmatch(authenticateHeader)[1]
			sipMessage.addAuthorization(*softphone, nonce).addCseq(softphone).newViaBranch()
			softphone.request(sipMessage, nil)
			return true
		}
		return false
	})
}
//a1 := md5Hex(username + ":" + hostname + ":" + account.password)
//	ha2 := md5Hex(sipnet.MethodRegister + ":" + authArgs.Get("uri"))
//	response := md5Hex(ha1 + ":" + session.nonce + ":" + authArgs.Get("nc") +
//		":" + authArgs.Get("cnonce") + ":auth:" + ha2)