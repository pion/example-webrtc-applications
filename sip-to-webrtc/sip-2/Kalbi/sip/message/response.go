package message

//NewRequest creates new SIP request
func NewResponse(request *SipReq, via *SipVia, to *SipTo, from *SipFrom, callID *SipVal, maxfor *SipVal) *SipMsg {
	r := new(SipMsg)
	r.Req = *request
	r.Via = append(r.Via, *via)
	r.To = *to
	r.From = *from
	r.CallId = *callID
	r.MaxFwd = *maxfor
	return r
}
