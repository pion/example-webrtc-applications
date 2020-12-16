package method

const (
	//REGISTER Register the URI listed in the To-header field with a location server and associates it with the network address given in a Contact header field. RFC_3261
	REGISTER = "REGISTER"
	//INVITE Initiate a dialog for establishing a call. The request is sent by a user agent client to a user agent server.
	INVITE = "INVITE"
	//ACK Confirm that an entity has received a final response to an INVITE request.
	ACK = "ACK"
	//BYE Signal termination of a dialog and end a call.
	BYE = "BYE"
	//CANCEL Cancel any pending request.
	CANCEL = "CANCEL"
	//UPDATE Modify the state of a session without changing the state of the dialog.
	UPDATE = "UPDATE"
	//REFER Ask recipient to issue a request for the purpose of call transfer.
	REFER = "REFER"
	//PRACK Provisional acknowledgement.
	PRACK = "PRACK"
	//SUBSCRIBE Initiates a subscription for notification of events from a notifier.
	SUBSCRIBE = "SUBSCRIBE"
	//NOTIFY Inform a subscriber of notifications of a new event.
	NOTIFY = "NOTIFY"
	//PUBLISH Publish an event to a notification server.
	PUBLISH = "PUBLISH"
	//MESSAGE Deliver a text message.
	MESSAGE = "MESSAGE"
	//INFO Send mid-session information that does not modify the session state.
	INFO = "INFO"
	//OPTIONS Query the capabilities of an endpoint.
	OPTIONS = "OPTIONS"
)
