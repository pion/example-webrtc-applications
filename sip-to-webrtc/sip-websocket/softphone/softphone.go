package softphone

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"strings"
	log "github.com/sirupsen/logrus"
	"math/rand"
)

type Softphone struct {
	OnInvite         func(inviteMessage SipMessage) //sipmessage.go
	sipInfo          SIPInfoResponse //util.go
	wsConn           *websocket.Conn
	fakeDomain       string
	FakeEmail        string
	fromTag          string
	toTag            string
	callId           string
	cseq             int
	messageListeners map[string]func(string)
	inviteKey        string
	messages         chan string
}

func NewSoftPhone(sipInfo SIPInfoResponse) *Softphone {
	configureLog()
	softphone := Softphone{}
	softphone.OnInvite = func(inviteMessage SipMessage) {}
	softphone.fakeDomain = uuid.New().String() + ".invalid"
	softphone.FakeEmail = uuid.New().String() + "@" + softphone.fakeDomain
	softphone.fromTag = uuid.New().String()
	softphone.toTag = uuid.New().String()
	softphone.callId = uuid.New().String()
	softphone.cseq = rand.Intn(10000) + 1
	softphone.messageListeners = make(map[string]func(string))
	softphone.sipInfo = sipInfo
	softphone.register()
	return &softphone
}

func (softphone *Softphone) addMessageListener(messageListener func(string)) string {
	key := uuid.New().String()
	softphone.messageListeners[key] = messageListener
	return key
}
func (softphone *Softphone) removeMessageListener(key string) {
	delete(softphone.messageListeners, key)
}

func (softphone Softphone) request2(sipMessage SipMessage, expectedResp string) string {
	println(sipMessage.ToString())
	softphone.wsConn.WriteMessage(1, []byte(sipMessage.ToString()))
	if expectedResp != "" {
		for {
			message := <-softphone.messages
			if strings.Contains(message, expectedResp) {
				return message
			}
		}
	}
	return ""
}

func (softphone *Softphone) request(sipMessage SipMessage, responseHandler func(string) bool) {
	//fmt.Printf("%+v\n", sipMessage.ToString())
	log.Debug("↑↑↑\n", sipMessage.ToString())
	if responseHandler != nil {
		var key string
		key = softphone.addMessageListener(func(message string) {
			done := responseHandler(message)
			if done {
				softphone.removeMessageListener(key)
			}
		})
	}
	err := softphone.wsConn.WriteMessage(1, []byte(sipMessage.ToString()))
	if err != nil {
		log.Fatal(err)
	}
}

func (softphone *Softphone) Response(message string) {
	log.Debug("↑↑↑\n", message)
	err := softphone.wsConn.WriteMessage(1, []byte(message))
	if err != nil {
		log.Fatal(err)
	}
}
