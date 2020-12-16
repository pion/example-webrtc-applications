package transaction

/*
   Author - Aaron Parfitt
   Date - 11th October 2020

   RFC3261 - SIP: Session Initiation Protocol
   https://tools.ietf.org/html/rfc3261#section-17.1

   Client Transaction

   The client transaction provides its functionality through the
   maintenance of a state machine.

   The TU communicates with the client transaction through a simple
   interface.  When the TU wishes to initiate a new transaction, it
   creates a client transaction and passes it the SIP request to send
   and an IP address, port, and transport to which to send it.  The
   client transaction begins execution of its state machine.  Valid
   responses are passed up to the TU from the client transaction.

   There are two types of client transaction state machines, depending
   on the method of the request passed by the TU.  One handles client
   transactions for INVITE requests.  This type of machine is referred
   to as an INVITE client transaction.  Another type handles client
   transactions for all requests except INVITE and ACK.  This is
   referred to as a non-INVITE client transaction.  There is no client
   transaction for ACK.  If the TU wishes to send an ACK, it passes one
   directly to the transport layer for transmission.

   The INVITE transaction is different from those of other methods
   because of its extended duration.  Normally, human input is required
   in order to respond to an INVITE.  The long delays expected for
   sending a response argue for a three-way handshake.  On the other
   hand, requests of other methods are expected to complete rapidly.
   Because of the non-INVITE transaction's reliance on a two-way
   handshake, TUs SHOULD respond immediately to non-INVITE requests. */

import (
	"Kalbi/interfaces"
	"Kalbi/log"
	"Kalbi/sip/message"
	"Kalbi/sip/method"
	"github.com/looplab/fsm"
	"time"
)

const (
	clientInputRequest      = "client_input_request"
	clientInput1xx          = "client_input_1xx"
	clientInput2xx          = "client_input_2xx"
	clientInput300Plus      = "client_input_300_plus"
	clientInputTimerA       = "client_input_timer_a"
	clientInputTimerB       = "client_input_timer_b"
	clientInputTimerD       = "client_input_timer_d"
	clientInputTransportErr = "client_input_transport_err"
	clientInputDelete       = "client_input_transport_err"
)

// ClientTransaction represents a client transaction references in RFC3261
type ClientTransaction struct {
	ID             string
	BranchID       string
	ServerTxID     string
	TransManager   *TransactionManager
	Origin         *message.SipMsg
	FSM            *fsm.FSM
	msgHistory     []*message.SipMsg
	ListeningPoint interfaces.ListeningPoint
	Host           string
	Port           string
	LastMessage    *message.SipMsg
	timerATime     time.Duration
	timerA         *time.Timer
	timerB         *time.Timer
	timerDTime     time.Duration
	timerD         *time.Timer
}

//InitFSM initializes the finite state machine within the client transaction
func (ct *ClientTransaction) InitFSM(msg *message.SipMsg) {
	switch string(msg.Req.Method) {
	case method.INVITE:
		ct.FSM = fsm.NewFSM("", fsm.Events{
			{Name: clientInputRequest, Src: []string{""}, Dst: "Calling"},
			{Name: clientInput1xx, Src: []string{"Calling"}, Dst: "Proceeding"},
			{Name: clientInput300Plus, Src: []string{"Proceeding"}, Dst: "Completed"},
			{Name: clientInput2xx, Src: []string{"Proceeding"}, Dst: "Terminated"},
			{Name: clientInputTransportErr, Src: []string{"Calling", "Proceeding", "Completed"}, Dst: "Terminated"},
		}, fsm.Callbacks{
			clientInput1xx: ct.act100,
			clientInput2xx: ct.actDelete,
			clientInput300Plus: ct.act300,
			clientInputTimerA:  ct.actResend,
			clientInputTimerB:  ct.actTransErr,
		})
	default:
		ct.FSM = fsm.NewFSM("", fsm.Events{
			{Name: clientInputRequest, Src: []string{""}, Dst: "Calling"},
			{Name: clientInput1xx, Src: []string{"Calling"}, Dst: "Proceeding"},
			{Name: clientInput300Plus, Src: []string{"Proceeding"}, Dst: "Completed"},
			{Name: clientInput2xx, Src: []string{"Proceeding"}, Dst: "Terminated"},
		}, fsm.Callbacks{})
	}
}

//SetListeningPoint sets a listening point to the client transaction
func (ct *ClientTransaction) SetListeningPoint(lp interfaces.ListeningPoint) {
	ct.ListeningPoint = lp
}

//GetListeningPoint returns current listening point
func (ct *ClientTransaction) GetListeningPoint() interfaces.ListeningPoint {
	return ct.ListeningPoint
}

//GetBranchID returns branchId which is the identifier of a transaction
func (ct *ClientTransaction) GetBranchID() string {
	return ct.BranchID
}

//GetOrigin returns the SIP message that initiated this transaction
func (ct *ClientTransaction) GetOrigin() *message.SipMsg {
	return ct.Origin
}

//Receive takes in the SIP message from the transport layer
func (ct *ClientTransaction) Receive(msg *message.SipMsg) {

	//fmt.Println("CURRENT STATE: " + ct.FSM.Current())
	ct.LastMessage = msg
	if msg.GetStatusCode() < 200 {
		err := ct.FSM.Event(clientInput1xx)
		if err != nil {
			log.Log.Error(err)
		}
	} else if msg.GetStatusCode() < 300 {
		err := ct.FSM.Event(clientInput2xx)
		if err != nil {
			log.Log.Error(err)
		}
	} else {
		err := ct.FSM.Event(clientInput300Plus)
		if err != nil {
			log.Log.Error(err)
		}
	}
}


func (ct *ClientTransaction) act100(event *fsm.Event) {
    ct.timerA.Stop()
}

//SetServerTransaction is used to set a Server Transaction
func (ct *ClientTransaction) SetServerTransaction(txID string) {
	ct.ServerTxID = txID
}

//GetServerTransactionID returns a ServerTransaction that has been set with SetServerTransaction()
func (ct *ClientTransaction) GetServerTransactionID() string {
	return ct.ServerTxID
}

//GetLastMessage returns the last received SIP message to this transaction
func (ct *ClientTransaction) GetLastMessage() *message.SipMsg {
	return ct.LastMessage
}

//SetLastMessage sets the last message received
func (ct *ClientTransaction) SetLastMessage(msg *message.SipMsg) {
	ct.LastMessage = msg
}

func (ct *ClientTransaction) actSend(event *fsm.Event) {
	err := ct.ListeningPoint.Send(ct.Host, ct.Port, ct.Origin.String())
	if err != nil {
		err2 := ct.FSM.Event(clientInputTransportErr)
		if err2 != nil {
			log.Log.Error("Event error in error handling for transactionID: " + ct.BranchID)
		}
	}
}

func (ct *ClientTransaction) act300(event *fsm.Event) {
	log.Log.Info("Client transaction %p, act_300", ct)
	ct.timerD = time.AfterFunc(ct.timerDTime, func() {
		err := ct.FSM.Event(clientInputTimerD)
		if err != nil {
			log.Log.Error("Event error for transactionID: " + ct.BranchID)
		}
	})
}

func (ct *ClientTransaction) actTransErr(event *fsm.Event) {
	log.Log.Error("Transport error for transactionID: " + ct.BranchID)
	err := ct.FSM.Event(clientInputDelete)
	if err != nil {
		log.Log.Error("Event error for transactionID: " + ct.BranchID)
	}
}

func (ct *ClientTransaction) actDelete(event *fsm.Event) {
	ct.TransManager.DeleteClientTransaction(string(ct.Origin.Via[0].Branch))
}

func (ct *ClientTransaction) actResend(event *fsm.Event) {
	log.Log.Info("Client transaction %p, act_resend", ct)
	ct.timerATime *= 2
	ct.timerA.Reset(ct.timerATime)
	ct.Resend()
}

//Resend is used for retransmissions
func (ct *ClientTransaction) Resend() {
	err := ct.ListeningPoint.Send(ct.Host, ct.Port, ct.Origin.String())
	if err != nil {
		err2 := ct.FSM.Event(clientInputTransportErr)
		if err2 != nil {
			log.Log.Error("Event error in error handling for transactionID: " + ct.BranchID)
		}
	}
}

//StatelessSend send a sip message without acting on the FSM
func (ct *ClientTransaction) StatelessSend(msg *message.SipMsg, host string, port string) {
	err := ct.ListeningPoint.Send(ct.Host, ct.Port, ct.Origin.String())
	if err != nil {
		log.Log.Error("Transport error for transactionID : " + ct.BranchID)
	}
}

//Send is used to send a SIP message
func (ct *ClientTransaction) Send(msg *message.SipMsg, host string, port string) {
	ct.Origin = msg
	ct.Host = host
	ct.Port = port
	ct.timerATime = T1

	//Retransmition timer
	ct.timerA = time.AfterFunc(ct.timerATime, func() {
		err := ct.FSM.Event(clientInputTimerA)
		if err != nil {
			log.Log.Error(err)
		}
	})

	//timeout timer
	ct.timerB = time.AfterFunc(64*T1, func() {
		err := ct.FSM.Event(clientInputTimerB)
		if err != nil {
			log.Log.Error(err)
		}
	})

	err := ct.ListeningPoint.Send(ct.Host, ct.Port, ct.Origin.String())
	if err != nil {
		err2 := ct.FSM.Event(serverInputTransportErr)
		if err2 != nil {
			log.Log.Error(err)
		}
	}
	ct.FSM.Event(clientInputRequest)
}
