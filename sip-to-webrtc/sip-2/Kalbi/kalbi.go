package kalbi

import (
	"Kalbi/interfaces"
	"Kalbi/log"
	"Kalbi/sip/dialog"
	"Kalbi/sip/message"
	"Kalbi/sip/transaction"
	"Kalbi/transport"
)

//NewSipStack  creates new sip stack
func NewSipStack(Name string) *SipStack {
	stack := new(SipStack)
	stack.Name = Name
	stack.TransManager = transaction.NewTransactionManager()
	stack.TransportChannel = make(chan interfaces.SipEventObject)

	return stack
}

//SipStack has multiple protocol listning points
type SipStack struct {
	Name             string
	ListeningPoints  []interfaces.ListeningPoint
	OutputPoint      chan message.SipMsg
	InputPoint       chan message.SipMsg
	Alive            bool
	TransManager     *transaction.TransactionManager
	Dialogs          []dialog.Dialog
	TransportChannel chan interfaces.SipEventObject
	sipListener      interfaces.SipListener
}

//GetTransactionManager returns TransactionManager
func (ed *SipStack) GetTransactionManager() *transaction.TransactionManager {
	return ed.TransManager
}

//CreateListenPoint creates listening point to the event dispatcher
func (ed *SipStack) CreateListenPoint(protocol string, host string, port int) interfaces.ListeningPoint {
	listenpoint := transport.NewTransportListenPoint(protocol, host, port)
	listenpoint.SetTransportChannel(ed.TransportChannel)
	ed.ListeningPoints = append(ed.ListeningPoints, listenpoint)
	return listenpoint
}

//SetSipListener sets a struct that follows the SipListener interface
func (ed *SipStack) SetSipListener(listener interfaces.SipListener) {
	ed.sipListener = listener

}

//IsAlive check if SipStack is alive
func (ed *SipStack) IsAlive() bool {
	return ed.Alive
}

//Stop stops SipStack execution
func (ed *SipStack) Stop() {
	log.Log.Info("Stopping SIPStack...")
	ed.Alive = false
}

//Start starts the sip stack
func (ed *SipStack) Start() {
	log.Log.Info("Starting SIPStack...")
	ed.TransManager.ListeningPoint = ed.ListeningPoints[0]
	ed.Alive = true
	for _, listeningPoint := range ed.ListeningPoints {
		go listeningPoint.Start()
	}

	for ed.Alive == true {
		msg := <-ed.TransportChannel
		event := ed.TransManager.Handle(msg)
		message := event.GetSipMessage()
		if message.Req.StatusCode != nil {
			go ed.sipListener.HandleResponses(event)
		} else if message.Req.Method != nil {
			go ed.sipListener.HandleRequests(event)
		}
		//TODO: Handle failed SIP parse send 400 Bad Request

	}
}
