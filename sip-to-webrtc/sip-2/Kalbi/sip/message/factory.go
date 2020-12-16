package message

//NewRequestLine creates new request line
func NewRequestLine(method string, uriType string, user string, host string, port string) *SipReq {
	requestLine := new(SipReq)
	requestLine.SetMethod(method)
	requestLine.SetUriType(uriType)
	requestLine.SetUser(user)
	requestLine.SetHost(host)
	requestLine.SetPort(port)
	return requestLine
}

//NewResponseLine creates new request line
func NewResponseLine(statuscode int, statusdesc string) *SipReq {
	requestLine := new(SipReq)
	requestLine.SetStatusCode(statuscode)
	requestLine.SetStatusDesc(statusdesc)
	return requestLine
}

//NewViaHeader creates new Via Header
func NewViaHeader(transport string, host string, port string) *SipVia {
	via := new(SipVia)
	via.SetTransport(transport)
	via.SetHost(host)
	via.SetPort(port)
	return via
}

//NewFromHeader creates new From header
func NewFromHeader(user string, uriType string, host string, port string) *SipFrom {
	from := new(SipFrom)
	from.SetUriType(uriType)
	from.SetUser(user)
	from.SetHost(host)
	from.SetPort(port)
	return from
}

//NewToHeader creates new To header
func NewToHeader(user string, uriType string, host string, port string) *SipTo {
	to := new(SipTo)
	to.SetUriType(uriType)
	to.SetUser(user)
	to.SetHost(host)
	to.SetPort(port)
	return to
}

//NewContactHeader creates new Contact Header
func NewContactHeader(uriType string, user string, host string) *SipContact {
	contact := new(SipContact)
	contact.SetUriType(uriType)
	contact.SetUser(user)
	contact.SetHost(host)
	return contact
}

//NewCallID creates new CallID Header
func NewCallID(value string) *SipVal {
	callid := new(SipVal)
	callid.SetValue(value)
	return callid
}

//NewCSeq creates new CSeq header
func NewCSeq(id string, method string) *SipCseq {
	cseq := new(SipCseq)
	cseq.SetMethod(method)
	cseq.SetID(id)
	return cseq
}

//NewMaxForwards creates new Max Forwards header
func NewMaxForwards(value string) *SipVal {
	maxFor := new(SipVal)
	maxFor.SetValue(value)
	return maxFor
}

//NewContentLength creates new Content Length Header
func NewContentLength(value string) *SipVal {
	contlen := new(SipVal)
	contlen.SetValue(value)
	return contlen
}


