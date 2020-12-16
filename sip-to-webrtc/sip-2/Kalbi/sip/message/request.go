package message

//NewRequest creates new SIP request
func NewRequest(request *SipReq, via *SipVia, to *SipTo, from *SipFrom, contact *SipContact, callID *SipVal, cseq *SipCseq, maxfor *SipVal, contlen *SipVal) *SipMsg {
	r := new(SipMsg)
	r.Req = *request
	r.Via = append(r.Via, *via)
	r.To = *to
	r.From = *from
	r.Contact = *contact
	r.CallId = *callID
	r.Cseq = *cseq
	r.MaxFwd = *maxfor
	r.ContLen = *contlen
	return r
}
