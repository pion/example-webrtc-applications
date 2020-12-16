package status

const (
	// 1xx Provisional Responses

	// Trying Extended search being performed may take a significant time so a forking proxy must send a 100 Trying response.
	// RFC_3261
	Trying = 100 // : "Trying"
	// Ringing Destination user agent received INVITE, and is alerting user of call.
	// RFC_3261
	Ringing = 180 // : "Ringing",
	// CallIsBeingForwarded Servers can optionally send this response to indicate a call is being forwarded.
	// RFC_3261
	CallIsBeingForwarded = 181 // : "CallIsBeingForwarded",
	// Queued Indicates that the destination was temporarily unavailable, so the server has queued the call until the destination is available.
	// A server may send multiple 182 responses to update progress of the queue.
	// RFC_3261
	Queued = 182 // : "Queued",
	// SessionProgress This response may be used to send extra information for a call which is still being set up.
	// RFC_3261
	SessionProgress = 183 // : "SessionProgress",
	// EarlyDialogTerminated Can be used by User Agent Server to indicate to upstream SIP entities (including the User Agent Client (UAC)) that an early dialog has been terminated.
	// RFC_6228
	EarlyDialogTerminated = 199 // : "EarlyDialogTerminated",

	// 2xx Successful Responses

	// OK Indicates that the request was successful.
	// RFC_3261
	OK = 200 // : "OK",
	// Accepted Indicates that the request has been accepted for processing, but the processing has not been completed.
	// RFC_3265
	// RFC_2616 10.2.3
	// Deprecated RFC_6665
	Accepted = 202 // : "Accepted",
	// NoNotification Indicates the request was successful, but the corresponding response will not be received.
	// RFC_5839 7.1
	NoNotification = 204 // : "NoNotification",

	// 3xx Redirection Responses

	// MultipleChoices The address resolved to one of several options for the user or client to choose between,
	// which are listed in the message body or the message's Contact fields.
	// RFC_3261
	MultipleChoices = 300 // : "MultipleChoices",
	// MovedPermanently The original Request-URI is no longer valid, the new address is given in the Contact header field,
	// and the client should update any records of the original Request-URI with the new value.
	// RFC_3261
	MovedPermanently = 301 // : "MovedPermanently",
	// MovedTemporarily The client should try at the address in the Contact field.
	// If an Expires field is present, the client may cache the result for that period of time.
	// RFC_3261
	MovedTemporarily = 302 // : "MovedTemporarily",
	// UseProxy The Contact field details a proxy that must be used to access the requested destination.
	// RFC_3261
	UseProxy = 305 // : "UseProxy",
	// AlternativeService The call failed, but alternatives are detailed in the message body.
	// RFC_3261
	AlternativeService = 380 // : "AlternativeService",

	// 4xx Client Failure Responses

	// BadRequest The request could not be understood due to malformed syntax.
	// RFC_3261
	BadRequest = 400 // : "BadRequest",
	// Unauthorized The request requires user authentication. This response is issued by UASs and registrars.
	// RFC_3261
	Unauthorized = 401 // : "Unauthorized",
	// PaymentRequired Reserved for future use.
	// RFC_3261
	PaymentRequired = 402 // : "PaymentRequired",
	// Forbidden The server understood the request, but is refusing to fulfill it.
	// RFC_3261
	// Sometimes (but not always) this means the call has been rejected by the receiver.
	Forbidden = 403 // : "Forbidden",
	// NotFound The server has definitive information that the user does not exist at the domain specified in the Request-URI.
	// This status is also returned if the domain in the Request-URI does not match any of the domains handled by the recipient of the request.
	// RFC_3261
	NotFound = 404 // : "NotFound",
	// MethodNotAllowed The method specified in the Request-Line is understood, but not allowed for the address identified by the Request-URI.
	// RFC_3261
	MethodNotAllowed = 405 // : "MethodNotAllowed",
	// NotAcceptable406 The resource identified by the request is only capable of generating response entities that have content characteristics but
	// not acceptable according to the Accept header field sent in the request.
	// RFC_3261
	NotAcceptable406 = 406 // : "NotAcceptable406",
	// ProxyAuthenticationRequired The request requires user authentication. This response is issued by proxys.
	// RFC_3261
	ProxyAuthenticationRequired = 407 // : "ProxyAuthenticationRequired",
	// RequestTimeout Couldn't find the user in time.
	// The server could not produce a response within a suitable amount of time,
	// for example, if it could not determine the location of the user in time.
	// The client MAY repeat the request without modifications at any later time.
	// RFC_3261
	RequestTimeout = 408 // : "RequestTimeout",
	// Conflict User already registered.
	// RFC_2543
	// Deprecated by omission from later RFCs RFC_3261 and by non-registration with the IANA.
	Conflict = 409 // : "Conflict",
	// Gone The user existed once, but is not available here any more.
	// RFC_3261
	Gone = 410 // : "Gone",
	// LengthRequired The server will not accept the request without a valid Content-Length. RFC_2543
	// Deprecated by omission from later RFCs RFC_3261 and by non-registration with the IANA.
	LengthRequired = 411 // : "LengthRequired",
	// ConditionalRequestFailed The given precondition has not been met.
	// RFC_3903 11.2.1
	ConditionalRequestFailed = 412 // : "ConditionalRequestFailed",
	// RequestEntityTooLarge Request body too large.
	// RFC_3261
	RequestEntityTooLarge = 413 // : "RequestEntityTooLarge",
	// RequestURITooLong The server is refusing to service the request because the Request-URI is longer than the server is willing to interpret.
	// RFC_3261
	RequestURITooLong = 414 // : "RequestURITooLong",
	// UnsupportedMediaType Request body in a format not supported.
	// RFC_3261
	UnsupportedMediaType = 415 // : "UnsupportedMediaType",
	// UnsupportedURIScheme Request-URI is unknown to the server.
	// RFC_3261
	UnsupportedURIScheme = 416 // : "UnsupportedURIScheme",
	// UnknownResourcePriority There was a resource-priority option tag, but no Resource-Priority header.
	// RFC_4412 4.6.2
	UnknownResourcePriority = 417 // : "UnknownResourcePriority",
	// BadExtension Bad SIP Protocol Extension used, not understood by the server.
	// RFC_3261
	BadExtension = 420 // : "BadExtension",
	// ExtensionRequired The server needs a specific extension not listed in the Supported header.
	// RFC_3261
	ExtensionRequired = 421 // : "ExtensionRequired",
	// SessionIntervalTooSmall The received request contains a Session-Expires header field with a duration below the minimum timer.
	// RFC_4028_sec_6
	SessionIntervalTooSmall = 422 // : "SessionIntervalTooSmall",
	// IntervalTooBrief Expiration time of the resource is too short.
	// RFC_3261
	IntervalTooBrief = 423 // : "IntervalTooBrief",
	// BadLocationInformation The request's location content was malformed or otherwise unsatisfactory.
	// RFC_6422
	// Location Conveyance for the Session Initiation Protocol RFC_6442 4.3
	BadLocationInformation = 424 // : "BadLocationInformation",
	// UseIdentityHeader The server policy requires an Identity header, and one has not been provided.
	// RFC_4474
	UseIdentityHeader = 428 // : "UseIdentityHeader",
	// ProvideReferrerIdentity The server did not receive a valid Referred-By token on the request.
	// RFC_3892 5
	ProvideReferrerIdentity = 429 // : "ProvideReferrerIdentity",
	// FlowFailed A specific flow to a user agent has failed, although other flows may succeed.
	// This response is intended for use between proxy devices,
	// and should not be seen by an endpoint (and if it is seen by one, should be treated as a [[#400|400 Bad Request]] response).
	// RFC_5626
	FlowFailed = 430 // : "FlowFailed",
	// AnonymityDisallowed The request has been rejected because it was anonymous.
	// RFC_5079 5
	AnonymityDisallowed = 433 // : "AnonymityDisallowed",
	// BadIdentityInfo The request has an Identity-Info header, and the URI scheme in that header cannot be dereferenced.
	// RFC_4474
	BadIdentityInfo = 436 // : "BadIdentityInfo",
	// UnsupportedCertificate The server was unable to validate a certificate for the domain that signed the request.<ref name="RFC_4474"/>{{rp|p11}}
	UnsupportedCertificate = 437 // : "UnsupportedCertificate",
	// InvalidIdentityHeader The server obtained a valid certificate that the request claimed was used to sign the request, but was unable to verify that signature.<ref name="RFC_4474"/>{{rp|p12}}
	InvalidIdentityHeader = 438 // : "InvalidIdentityHeader",
	// FirstHopLacksOutboundSupport The first outbound [[Session Initiation Protocol#Proxy server|proxy]] the user is attempting to register through does not
	// support the "outbound" feature of RFC 5626, although the [[Session Initiation Protocol#Registrar|registrar]] does.<ref name="RFC_5626"/>{{rp|§11.6}}
	FirstHopLacksOutboundSupport = 439 // : "FirstHopLacksOutboundSupport",
	// MaxBreadthExceeded If a SIP proxy determines a response context has insufficient Incoming Max-Breadth to carry out a desired parallel fork,
	// and the proxy is unwilling/unable to compensate by forking serially or sending a redirect, that proxy MUST return a 440 response.
	// A client receiving a 440 response can infer that its request did not reach all possible destinations.
	// RFC_5393
	MaxBreadthExceeded = 440 // : "MaxBreadthExceeded",
	// BadInfoPackage If a SIP UA receives an INFO request associated with an Info Package that the UA has not indicated willingness to receive,
	// the UA MUST send a 469 response, which contains a Recv-Info header field with Info Packages for which the UA is willing to receive INFO requests.
	// 6086
	BadInfoPackage = 469 // : "BadInfoPackage",
	// ConsentNeeded The source of the request did not have the permission of the recipient to make such a request.<ref name="RFC_5360">{{cite IETF
	//| section      = 5.9.2
	ConsentNeeded = 470 // : "ConsentNeeded",
	// TemporarilyUnavailable Callee currently unavailable.<ref name="RFC_3261"/>{{rp|§21.4.18}}
	TemporarilyUnavailable = 480 // : "TemporarilyUnavailable",
	// TransactionDoesNotExist Server received a request that does not match any dialog or transaction.<ref name="RFC_3261"/>{{rp|§21.4.19}}
	TransactionDoesNotExist = 481 // : "TransactionDoesNotExist",
	// LoopDetected Server has detected a loop.<ref name="RFC_3261"/>{{rp|§21.4.20}}
	LoopDetected = 482 // : "LoopDetected",
	// TooManyHops Max-Forwards header has reached the value '0'.<ref name="RFC_3261"/>{{rp|§21.4.21}}
	TooManyHops = 483 // : "TooManyHops",
	// AddressIncomplete Request-URI incomplete.<ref name="RFC_3261"/>{{rp|§21.4.22}}
	AddressIncomplete = 484 // : "AddressIncomplete",
	// Ambiguous Request-URI is ambiguous.<ref name="RFC_3261"/>{{rp|§21.4.23}}
	Ambiguous = 485 // : "Ambiguous",
	// BusyHere Callee is busy.<ref name="RFC_3261"/>{{rp|§21.4.24}}
	BusyHere = 486 // : "BusyHere",
	// RequestTerminated Request has terminated by bye or cancel.<ref name="RFC_3261"/>{{rp|§21.4.25}}
	RequestTerminated = 487 // : "RequestTerminated",
	// NotAcceptableHere Some aspect of the session description or the Request-URI is not acceptable.<ref name="RFC_3261"/>{{rp|§21.4.26}}
	NotAcceptableHere = 488 // : "NotAcceptableHere",
	// BadEvent The server did not understand an event package specified in an Event header field.<ref name="RFC_3265"/>{{rp|§7.3.2}}<ref name="RFC_6665"/>{{rp|§8.3.2}}
	BadEvent = 489 // : "BadEvent",
	// RequestPending Server has some pending request from the same dialog.<ref name="RFC_3261"/>{{rp|§21.4.27}}
	RequestPending = 491 // : "RequestPending",
	// Undecipherable Request contains an encrypted MIME body, which recipient can not decrypt.<ref name="RFC_3261"/>{{rp|§21.4.28}}
	Undecipherable = 493 // : "Undecipherable",
	// SecurityAgreementRequired The server has received a request that requires a negotiated security mechanism,
	// and the response contains a list of suitable security mechanisms for the requester to choose between, RFC_3329 or a [[digest authentication]] challenge.
	// RFC_3329
	SecurityAgreementRequired = 494 // : "SecurityAgreementRequired",

	// 5xx Server Failure Responses

	// InternalServerError The server could not fulfill the request due to some unexpected condition.<ref name="RFC_3261"/>{{rp|§21.5.1}}
	InternalServerError = 500 // : "InternalServerError",
	// NotImplemented The server does not have the ability to fulfill the request, such as because it does not recognize the request method.
	// (Compare with [[#405|405 Method Not Allowed]], where the server recognizes the method but does not allow or support it.)
	// RFC_3261
	NotImplemented = 501 // : "NotImplemented",
	// BadGateway The server is acting as a [[Gateway (computer program)|gateway]] or [[Proxy server|proxy]], and received an invalid response from a downstream server while attempting to fulfill the request.<ref name="RFC_3261"/>{{rp|§21.5.3}}
	BadGateway = 502 // : "BadGateway",
	// ServiceUnavailable The server is undergoing maintenance or is temporarily overloaded and so cannot process the request. A "Retry-After" header field may specify when the client may reattempt its request.<ref name="RFC_3261"/>{{rp|§21.5.4}}
	ServiceUnavailable = 503 // : "ServiceUnavailable",
	// ServerTimeout The server attempted to access another server in attempting to process the request, and did not receive a prompt response.<ref name="RFC_3261"/>{{rp|§21.5.5}}
	ServerTimeout = 504 // : "ServerTimeout",
	// VersionNotSupported The SIP protocol version in the request is not supported by the server.<ref name="RFC_3261"/>{{rp|§21.5.6}}
	VersionNotSupported = 505 // : "VersionNotSupported",
	// MessageTooLarge The request message length is longer than the server can process.<ref name="RFC_3261"/>{{rp|§21.5.7}}
	MessageTooLarge = 513 // : "MessageTooLarge",
	// PushNotificationServiceNotSupported The server does not support the push notification service identified in a 'pn-provider' SIP URI parameter<ref name="RFC_8599">{{cite IETF
	PushNotificationServiceNotSupported = 555 // : "PushNotificationServiceNotSupported",
	// PreconditionFailure The server is unable or unwilling to meet some constraints specified in the offer.
	// RFC_3312 8
	PreconditionFailure = 580 // : "PreconditionFailure",

	// 6xx Global Failure Responses

	// BusyEverywhere All possible destinations are busy. Unlike the 486 response, this response indicates the destination knows there are no alternative destinations (such as a voicemail server) able to accept the call.<ref name="RFC_3261"/>{{rp|§21.6.1}}
	BusyEverywhere = 600 // : "BusyEverywhere",
	// Decline The destination does not wish to participate in the call, or cannot do so, and additionally the destination knows there are no alternative destinations (such as a voicemail server) willing to accept the call.<ref name="RFC_3261"/>{{rp|§21.6.2}} The response may indicate a better time to call in the Retry-After header field.
	Decline = 603 // : "Decline",
	// DoesNotExistAnywhere The server has authoritative information that the requested user does not exist anywhere.<ref name="RFC_3261"/>{{rp|§21.6.3}}
	DoesNotExistAnywhere = 604 // : "DoesNotExistAnywhere",
	// NotAcceptable606 The user's agent was contacted successfully but some aspects of the session description such as the requested media, bandwidth, or addressing style were not acceptable.<ref name="RFC_3261"/>{{rp|§21.6.4}}
	NotAcceptable606 = 606 // : "NotAcceptable606",
	// Unwanted The called party did not want this call from the calling party. Future attempts from the calling party are likely to be similarly rejected.<ref name="RFC_8197">{{cite IETF
	Unwanted = 607 // : "Unwanted",
	// Rejected An intermediary machine or process rejected the call attempt.
	// RFC_8688
	// This contrasts with the 607 (Unwanted) SIP response code in which the called party rejected the call.
	// The response may include the contact entities that blocked the call in Call-Info header containing.
	// This provides a remediation mechanism for legal callers that find their calls blocked.
	Rejected = 608 // : "Rejected",
)

var statusText = map[int]string{
	// 1xx Provisional Responses,
	Trying:                "Trying",
	Ringing:               "Ringing",
	CallIsBeingForwarded:  "CallIsBeingForwarded",
	Queued:                "Queued",
	SessionProgress:       "SessionProgress",
	EarlyDialogTerminated: "EarlyDialogTerminated",
	// 2xx Successful Responses
	OK:             "OK",
	Accepted:       "Accepted",
	NoNotification: "NoNotification",
	// 3xx Redirection Responses
	MultipleChoices:    "MultipleChoices",
	MovedPermanently:   "MovedPermanently",
	MovedTemporarily:   "MovedTemporarily",
	UseProxy:           "UseProxy",
	AlternativeService: "AlternativeService",
	// 4xx Client Failure Responses
	BadRequest:                   "BadRequest",
	Unauthorized:                 "Unauthorized",
	PaymentRequired:              "PaymentRequired",
	Forbidden:                    "Forbidden",
	NotFound:                     "NotFound",
	MethodNotAllowed:             "MethodNotAllowed",
	NotAcceptable406:             "NotAcceptable406",
	ProxyAuthenticationRequired:  "ProxyAuthenticationRequired",
	RequestTimeout:               "RequestTimeout",
	Conflict:                     "Conflict",
	Gone:                         "Gone",
	LengthRequired:               "LengthRequired",
	ConditionalRequestFailed:     "ConditionalRequestFailed",
	RequestEntityTooLarge:        "RequestEntityTooLarge",
	RequestURITooLong:            "RequestURITooLong",
	UnsupportedMediaType:         "UnsupportedMediaType",
	UnsupportedURIScheme:         "UnsupportedURIScheme",
	UnknownResourcePriority:      "UnknownResourcePriority",
	BadExtension:                 "BadExtension",
	ExtensionRequired:            "ExtensionRequired",
	SessionIntervalTooSmall:      "SessionIntervalTooSmall",
	IntervalTooBrief:             "IntervalTooBrief",
	BadLocationInformation:       "BadLocationInformation",
	UseIdentityHeader:            "UseIdentityHeader",
	ProvideReferrerIdentity:      "ProvideReferrerIdentity",
	FlowFailed:                   "FlowFailed",
	AnonymityDisallowed:          "AnonymityDisallowed",
	BadIdentityInfo:              "BadIdentityInfo",
	UnsupportedCertificate:       "UnsupportedCertificate",
	InvalidIdentityHeader:        "InvalidIdentityHeader",
	FirstHopLacksOutboundSupport: "FirstHopLacksOutboundSupport",
	MaxBreadthExceeded:           "MaxBreadthExceeded",
	BadInfoPackage:               "BadInfoPackage",
	ConsentNeeded:                "ConsentNeeded",
	TemporarilyUnavailable:       "TemporarilyUnavailable",
	TransactionDoesNotExist:      "TransactionDoesNotExist",
	LoopDetected:                 "LoopDetected",
	TooManyHops:                  "TooManyHops",
	AddressIncomplete:            "AddressIncomplete",
	Ambiguous:                    "Ambiguous",
	BusyHere:                     "BusyHere",
	RequestTerminated:            "RequestTerminated",
	NotAcceptableHere:            "NotAcceptableHere",
	BadEvent:                     "BadEvent",
	RequestPending:               "RequestPending",
	Undecipherable:               "Undecipherable",
	SecurityAgreementRequired:    "SecurityAgreementRequired",
	// 5xx Server Failure Responses
	InternalServerError:                 "InternalServerError",
	NotImplemented:                      "NotImplemented",
	BadGateway:                          "BadGateway",
	ServiceUnavailable:                  "ServiceUnavailable",
	ServerTimeout:                       "ServerTimeout",
	VersionNotSupported:                 "VersionNotSupported",
	MessageTooLarge:                     "MessageTooLarge",
	PushNotificationServiceNotSupported: "PushNotificationServiceNotSupported",
	PreconditionFailure:                 "PreconditionFailure",
	// 6xx Global Failure Responses,
	BusyEverywhere:       "BusyEverywhere",
	Decline:              "Decline",
	DoesNotExistAnywhere: "DoesNotExistAnywhere",
	NotAcceptable606:     "NotAcceptable606",
	Unwanted:             "Unwanted",
	Rejected:             "Rejected",
}

// StatusText returns a text for the HTTP status code. It returns the empty
// string if the code is unknown.
func StatusText(code int) string {
	return statusText[code]
}
