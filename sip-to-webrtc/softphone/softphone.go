// Package softphone provides abstractions for SIP over Websocket
package softphone

import (
	"log"
	"math/rand"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Softphone ...
type Softphone struct {
	OnInvite         func(inviteMessage SIPMessage) // sipmessage.go
	sipInfo          SIPInfoResponse                // util.go
	wsConn           *websocket.Conn
	fakeDomain       string
	FakeEmail        string
	fromTag          string
	toTag            string
	callID           string
	cseq             int
	messageListeners map[string]func(string)
	inviteKey        string
	messages         chan string
}

// NewSoftPhone ...
func NewSoftPhone(sipInfo SIPInfoResponse) *Softphone {
	softphone := Softphone{}
	softphone.OnInvite = func(inviteMessage SIPMessage) {}
	softphone.fakeDomain = uuid.New().String() + ".invalid"
	softphone.FakeEmail = uuid.New().String() + "@" + softphone.fakeDomain
	softphone.fromTag = uuid.New().String()
	softphone.toTag = uuid.New().String()
	softphone.callID = uuid.New().String()
	softphone.cseq = rand.Intn(10000) + 1 //nolint
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

func (softphone Softphone) request2(sipMessage SIPMessage, expectedResp string) string {


	if err := softphone.wsConn.WriteMessage(1, []byte(sipMessage.ToString())); err != nil {
		log.Panic(err)
	}

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

func (softphone *Softphone) request(sipMessage SIPMessage, responseHandler func(string) bool) {
	// log.Print("↑↑↑\n", sipMessage.ToString())
	if responseHandler != nil {
		var key string
		key = softphone.addMessageListener(func(message string) {
			done := responseHandler(message)
			if done {
				softphone.removeMessageListener(key)
			}
		})
	}

	if err := softphone.wsConn.WriteMessage(1, []byte(sipMessage.ToString())); err != nil {
		log.Fatal(err)
	}
}

func (softphone *Softphone) response(message string) {
//	log.Print("↑↑↑\n", message)

	if err := softphone.wsConn.WriteMessage(1, []byte(message)); err != nil {
		log.Fatal(err)
	}
}
