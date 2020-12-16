package transaction

import (
	"Kalbi/interfaces"
	"Kalbi/log"
	"Kalbi/sip/message"
	"Kalbi/sip/method"
	"github.com/sirupsen/logrus"
	"sync"
)

//NewTransactionManager returns a new TransactionManager
func NewTransactionManager() *TransactionManager {
	txmng := new(TransactionManager)
	txmng.ClientTX = make(map[string]interfaces.Transaction)
	txmng.ServerTX = make(map[string]interfaces.Transaction)
	txmng.txLock = &sync.RWMutex{}

	return txmng
}

//TransactionManager handles SIP transactions
type TransactionManager struct {
	ServerTX        map[string]interfaces.Transaction
	ClientTX        map[string]interfaces.Transaction
	RequestChannel  chan interfaces.Transaction
	ResponseChannel chan interfaces.Transaction
	ListeningPoint  interfaces.ListeningPoint
	txLock          *sync.RWMutex
}

// Handle runs TransManager
func (tm *TransactionManager) Handle(event interfaces.SipEventObject) interfaces.SipEventObject {

	message := event.GetSipMessage()

	if message.Req.StatusCode != nil {
		log.Log.Info("Client transaction")

		tx, exists := tm.FindClientTransaction(message)
		if exists {
			tx.SetLastMessage(message)
			log.Log.Info("Client Transaction already exists")
		} else {
			tx = tm.NewClientTransaction(message)
		}

		tx.Receive(message)
		event.SetTransaction(tx)
		return event

	} else if message.Req.Method != nil {
		log.Log.Info("Server transaction")
		tx, exists := tm.FindServerTransaction(message)

		if exists {
			tx.SetLastMessage(message)
			log.Log.Info("Server Transaction already exists")

		} else {
			tx = tm.NewServerTransaction(message)
		}

		tx.Receive(message)

		event.SetTransaction(tx)

		return event
	}
	return event
}

//FindServerTransaction finds transaction by SipMsg
func (tm *TransactionManager) FindServerTransaction(msg *message.SipMsg) (interfaces.Transaction, bool) {
	//key := tm.MakeKey(*msg)
	tm.txLock.RLock()
	tx, exists := tm.ServerTX[string(msg.Via[0].Branch)]
	tm.txLock.RUnlock()
	return tx, exists
}

//FindClientTransaction finds transaction by SipMsg
func (tm *TransactionManager) FindClientTransaction(msg *message.SipMsg) (interfaces.Transaction, bool) {
	//key := tm.MakeKey(*msg)
	tm.txLock.RLock()
	tx, exists := tm.ClientTX[string(msg.Via[0].Branch)]
	tm.txLock.RUnlock()
	return tx, exists
}

//FindServerTransactionByID finds transaction by id
func (tm *TransactionManager) FindServerTransactionByID(value string) (interfaces.Transaction, bool) {
	//key := tm.MakeKey(*msg)
	tm.txLock.RLock()
	tx, exists := tm.ServerTX[value]
	tm.txLock.RUnlock()
	return tx, exists
}

//FindClientTransactionByID finds transaction by id
func (tm *TransactionManager) FindClientTransactionByID(value string) (interfaces.Transaction, bool) {
	//key := tm.MakeKey(*msg)
	tm.txLock.RLock()
	tx, exists := tm.ClientTX[value]
	tm.txLock.RUnlock()
	return tx, exists
}

//PutServerTransaction stores a transaction
func (tm *TransactionManager) PutServerTransaction(tx interfaces.Transaction) {
	tm.txLock.Lock()
	tm.ServerTX[tx.GetBranchID()] = tx
	tm.txLock.Unlock()
}

//PutClientTransaction stores a transaction
func (tm *TransactionManager) PutClientTransaction(tx interfaces.Transaction) {
	tm.txLock.Lock()
	tm.ClientTX[tx.GetBranchID()] = tx
	tm.txLock.Unlock()
}

//DeleteServerTransaction removes a stored transaction
func (tm *TransactionManager) DeleteServerTransaction(branch string) {
	log.Log.Info("Deleting transaction with ID: " + branch)
	log.Log.WithFields(logrus.Fields{"transactions": len(tm.ServerTX)}).Debug("Current transaction count before DeleteTransaction() is called")
	tm.txLock.Lock()
	delete(tm.ServerTX, branch)
	tm.txLock.Unlock()
	log.Log.WithFields(logrus.Fields{"transactions": len(tm.ServerTX)}).Debug("Current transaction count after DeleteTransaction() is called")
}

//DeleteClientTransaction removes a stored transaction
func (tm *TransactionManager) DeleteClientTransaction(branch string) {
	log.Log.Info("Deleting transaction with ID: " + branch)
	log.Log.WithFields(logrus.Fields{"transactions": len(tm.ClientTX)}).Debug("Current transaction count before DeleteTransaction() is called")
	tm.txLock.Lock()
	delete(tm.ClientTX, branch)
	tm.txLock.Unlock()
	log.Log.WithFields(logrus.Fields{"transactions": len(tm.ClientTX)}).Debug("Current transaction count after DeleteTransaction() is called")
}

//MakeKey creates new transaction identifier
func (tm *TransactionManager) MakeKey(msg message.SipMsg) string {

	key := string(msg.Via[0].Branch)
	var _method string
	if msg.Req.Method != nil {
		if string(msg.Req.Method) == method.ACK {
			_method = method.INVITE
		} else {
			_method = string(msg.Req.Method)
		}
	} else {
		_method = string(msg.Cseq.Method)
	}

	key += _method
	return key
}

//NewClientTransaction builds new CLientTransaction
func (tm *TransactionManager) NewClientTransaction(msg *message.SipMsg) *ClientTransaction {

	tx := new(ClientTransaction)
	tx.SetListeningPoint(tm.ListeningPoint)
	tx.TransManager = tm

	tx.InitFSM(msg)

	tx.BranchID = string(msg.Via[0].Branch)
	tx.Origin = msg

	tm.PutClientTransaction(tx)
	return tx

}

//NewServerTransaction builds new ServerTransaction
func (tm *TransactionManager) NewServerTransaction(msg *message.SipMsg) *ServerTransaction {

	tx := new(ServerTransaction)
	tx.SetListeningPoint(tm.ListeningPoint)

	tx.TransManager = tm

	tx.InitFSM(msg)

	tx.BranchID = string(msg.Via[0].Branch)
	tx.Origin = msg

	tm.PutServerTransaction(tx)
	return tx

}
