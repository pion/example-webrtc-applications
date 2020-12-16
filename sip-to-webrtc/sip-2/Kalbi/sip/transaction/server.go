package transaction

/*
Author - Aaron Parfitt
Date - 11th October 2020

RFC3261 - SIP: Session Initiation Protocol
https://tools.ietf.org/html/rfc3261#section-17.2

Server Transaction

   The server transaction is responsible for the delivery of requests to
   the TU and the reliable transmission of responses.  It accomplishes
   this through a state machine.  Server transactions are created by the
   core when a request is received, and transaction handling is desired
   for that request (this is not always the case).

   As with the client transactions, the state machine depends on whether
   the received request is an INVITE request.
*/

import (
	//"fmt"
	"Kalbi/interfaces"
	"Kalbi/log"
	"Kalbi/sip/message"
	"Kalbi/sip/method"
	"github.com/looplab/fsm"
)

const (
	serverInputRequest      = "server_input_request"
	serverInputAck          = "server_input_ack"
	serverInputUser1xx      = "server_input_user_1xx"
	serverInputUser2xx      = "server_input_user_2xx"
	serverInputUser300Plus  = "server_input_user_300_plus"
	serverInputTimerG       = "server_input_timer_g"
	serverInputTimerH       = "server_input_timer_h"
	serverInputTimerI       = "server_input_timer_i"
	serverInputTransportErr = "server_input_transport_err"
	serverInputDelete       = "server_input_delete"
)

//ServerTransaction is a representation of a Server Transaction references in RFC3261
type ServerTransaction struct {
	ID             string
	BranchID       string
	TransManager   *TransactionManager
	Origin         *message.SipMsg
	FSM            *fsm.FSM
	msgHistory     []*message.SipMsg
	ListeningPoint interfaces.ListeningPoint
	Host           string
	Port           string
	LastMessage    *message.SipMsg
}

//InitFSM initializes the finite state machine within the client transaction
func (st *ServerTransaction) InitFSM(msg *message.SipMsg) {

	switch string(msg.Req.Method) {
	case method.INVITE:
		st.FSM = fsm.NewFSM("", fsm.Events{
			{Name: serverInputRequest, Src: []string{""}, Dst: "Proceeding"},
			{Name: serverInputUser1xx, Src: []string{"Proceeding"}, Dst: "Proceeding"},
			{Name: serverInputUser300Plus, Src: []string{"Proceeding"}, Dst: "Completed"},
			{Name: serverInputAck, Src: []string{"Completed"}, Dst: "Confirmed"},
			{Name: serverInputUser2xx, Src: []string{"Proceeding"}, Dst: "Terminated"},
		}, fsm.Callbacks{serverInputUser1xx: st.actRespond,
			serverInputTransportErr: st.actTransErr,
			serverInputUser2xx:      st.actRespondDelete,
			serverInputUser300Plus:  st.actRespond})
	default:
		st.FSM = fsm.NewFSM("", fsm.Events{
			{Name: serverInputRequest, Src: []string{""}, Dst: "Proceeding"},
			{Name: serverInputUser1xx, Src: []string{"Proceeding"}, Dst: "Proceeding"},
			{Name: serverInputUser300Plus, Src: []string{"Proceeding"}, Dst: "Completed"},
			{Name: serverInputAck, Src: []string{"Completed"}, Dst: "Confirmed"},
			{Name: serverInputUser2xx, Src: []string{"Proceeding"}, Dst: "Terminated"},
		}, fsm.Callbacks{serverInputUser1xx: st.actRespond,
			serverInputTransportErr: st.actTransErr,
			serverInputUser2xx:      st.actRespondDelete,
			serverInputUser300Plus:  st.actRespond})
	}
}

//SetListeningPoint sets a listening point to the client transaction
func (st *ServerTransaction) SetListeningPoint(lp interfaces.ListeningPoint) {
	st.ListeningPoint = lp
}

//GetListeningPoint returns current listening point
func (st *ServerTransaction) GetListeningPoint() interfaces.ListeningPoint {
	return st.ListeningPoint
}

//GetBranchID returns branchId which is the identifier of a transaction
func (st *ServerTransaction) GetBranchID() string {
	return st.BranchID
}

//GetOrigin returns the SIP message that initiated this transaction
func (st *ServerTransaction) GetOrigin() *message.SipMsg {
	return st.Origin
}

//GetLastMessage returns the last received SIP message to this transaction
func (st *ServerTransaction) GetLastMessage() *message.SipMsg {
	return st.LastMessage
}

//SetLastMessage sets the last message received
func (st *ServerTransaction) SetLastMessage(msg *message.SipMsg) {
	st.LastMessage = msg
}

//Receive takes in the SIP message from the transport layer
func (st *ServerTransaction) Receive(msg *message.SipMsg) {
	st.LastMessage = msg
	log.Log.Info("Message Received for transactionId " + st.BranchID + ": \n" + string(msg.Src))
	log.Log.Info(message.MessageDetails(msg))
	if msg.Req.Method != nil || string(msg.Req.Method) != method.ACK {
		err := st.FSM.Event(serverInputRequest)
		if err != nil {
			log.Log.Error(err)
		}
	}

}

//Respond is used to process response from transport layer
func (st *ServerTransaction) Respond(msg *message.SipMsg) {
	//TODO: this will change due to issue https://Kalbi/issues/20
	log.Log.Info("Message Sent for transactionId " + st.BranchID + ": \n" + message.MessageDetails(msg))
	if msg.GetStatusCode() < 200 {
		err := st.FSM.Event(serverInputUser1xx)
		if err != nil {
			log.Log.Error(err)
		}
	} else if msg.GetStatusCode() < 300 {
		err := st.FSM.Event(serverInputUser2xx)
		if err != nil {
			log.Log.Error(err)
		}
	} else {
		err := st.FSM.Event(serverInputUser300Plus)
		if err != nil {
			log.Log.Error(err)
		}
	}

}

//GetServerTransactionID  returns Server transaction ID
func (st *ServerTransaction) GetServerTransactionID() string {
	return st.GetBranchID()
}

//Send used to send SIP message to specified host
func (st *ServerTransaction) Send(msg *message.SipMsg, host string, port string) {
	st.LastMessage = msg
	st.Host = host
	st.Port = port
	st.Respond(msg)
}

func (st *ServerTransaction) actRespond(event *fsm.Event) {
	err := st.ListeningPoint.Send(st.Host, st.Port, st.LastMessage.String())
	if err != nil {
		err2 := st.FSM.Event(serverInputTransportErr)
		if err2 != nil {
			log.Log.Error(err)
		}
	}

}

func (st *ServerTransaction) actRespondDelete(event *fsm.Event) {
	err := st.ListeningPoint.Send(st.Host, st.Port, st.LastMessage.String())
	if err != nil {
		err2 := st.FSM.Event(serverInputTransportErr)
		if err2 != nil {
			log.Log.Error(err)
		}
	}
	st.TransManager.DeleteServerTransaction(st.BranchID)
}

func (st *ServerTransaction) actTransErr(event *fsm.Event) {
	log.Log.Error("Transport error for transactionID : " + st.BranchID)
	err := st.FSM.Event(serverInputDelete)
	if err != nil {
		log.Log.Error(err)
	}
}

func (st *ServerTransaction) actDelete(event *fsm.Event) {
	st.TransManager.DeleteServerTransaction(st.BranchID)
}
