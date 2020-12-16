package dialog

import (
	"Kalbi/interfaces"
	"Kalbi/log"
	"sync"
)

/*

RFC3261 - https://tools.ietf.org/html/rfc3261#section-12

12 Dialogs

   A key concept for a user agent is that of a dialog.  A dialog
   represents a peer-to-peer SIP relationship between two user agents
   that persists for some time.  The dialog facilitates sequencing of
   messages between the user agents and proper routing of requests
   between both of them.  The dialog represents a context in which to
   interpret SIP messages.  Section 8 discussed method independent UA
   processing for requests and responses outside of a dialog.  This
   section discusses how those requests and responses are used to
   construct a dialog, and then how subsequent requests and responses
   are sent within a dialog.

   A dialog is identified at each UA with a dialog ID, which consists of
   a Call-ID value, a local tag and a remote tag.  The dialog ID at each
   UA involved in the dialog is not the same.  Specifically, the local
   tag at one UA is identical to the remote tag at the peer UA.  The
   tags are opaque tokens that facilitate the generation of unique
   dialog IDs.

   A dialog ID is also associated with all responses and with any
   request that contains a tag in the To field.  The rules for computing
   the dialog ID of a message depend on whether the SIP element is a UAC
   or UAS.  For a UAC, the Call-ID value of the dialog ID is set to the
   Call-ID of the message, the remote tag is set to the tag in the To
   field of the message, and the local tag is set to the tag in the From

   field of the message (these rules apply to both requests and
   responses).  As one would expect for a UAS, the Call-ID value of the
   dialog ID is set to the Call-ID of the message, the remote tag is set
   to the tag in the From field of the message, and the local tag is set
   to the tag in the To field of the message.

   A dialog contains certain pieces of state needed for further message
   transmissions within the dialog.  This state consists of the dialog
   ID, a local sequence number (used to order requests from the UA to
   its peer), a remote sequence number (used to order requests from its
   peer to the UA), a local URI, a remote URI, remote target, a boolean
   flag called "secure", and a route set, which is an ordered list of
   URIs.  The route set is the list of servers that need to be traversed
   to send a request to the peer.  A dialog can also be in the "early"
   state, which occurs when it is created with a provisional response,
   and then transition to the "confirmed" state when a 2xx final
   response arrives.  For other responses, or if no response arrives at
   all on that dialog, the early dialog terminates.

12.1 Creation of a Dialog

   Dialogs are created through the generation of non-failure responses
   to requests with specific methods.  Within this specification, only
   2xx and 101-199 responses with a To tag, where the request was
   INVITE, will establish a dialog.  A dialog established by a non-final
   response to a request is in the "early" state and it is called an
   early dialog.  Extensions MAY define other means for creating
   dialogs.  Section 13 gives more details that are specific to the
   INVITE method.  Here, we describe the process for creation of dialog
   state that is not dependent on the method.

   UAs MUST assign values to the dialog ID components as described
   below.

*/

//NewDialogManager returns new Dialog Manager
func NewDialogManager() *DialogManager {
	diagMng := new(DialogManager)

	diagMng.dialogs = make(map[string]Dialog)
	diagMng.Lock = &sync.RWMutex{}
	return new(DialogManager)
}

//DialogManager hold multiple dialogs
type DialogManager struct {
	dialogs map[string]Dialog
	Lock    *sync.RWMutex
}

//GetDialog returns dialog by ID
func (dm *DialogManager) GetDialog(value string) *Dialog {
	dm.Lock.RLock()
	diag, exists := dm.dialogs[value]
	dm.Lock.RUnlock()
	if exists {
		return &diag
	}
	log.Log.Info("Dialog doesn't exist")
	return nil
}

//DeleteDialog removes dialog from Dialog Manager by ID
func (dm *DialogManager) DeleteDialog(value string) {
	log.Log.Info("Deleting Dialog ")
	delete(dm.dialogs, value)

}

//NewDialog creates a new Dialog
func (dm *DialogManager) NewDialog() *Dialog {
	diag := new(Dialog)
	diag.DialogId = GenerateDialogId()

	return diag
}

//Dialog used to store track multiple transactions
type Dialog struct {
	DialogId int32
	CallId   string
	ToTag    string
	FromTag  string
	ServerTx interfaces.Transaction
	ClientTx interfaces.Transaction
	Cseq     uint32
}
