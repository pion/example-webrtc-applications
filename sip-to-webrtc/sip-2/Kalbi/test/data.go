package test

const msg = "INVITE sip:1234@127.0.0.1;transport=UDP SIP/2.0\r\n" +
	"Via: SIP/2.0/UDP 127.0.0.1:43842;branch=z9hG4bK-524287-1---c71235afbdc0efe8;rport\r\n" +
	"Max-Forwards: 70\r\n" +
	"Contact: <sip:22@127.0.0.1:43842;transport=UDP>\r\n" +
	"To: <sip:1234@127.0.0.1>\r\n" +
	"From: <sip:22@127.0.0.1;transport=UDP>;tag=14c6d77c\r\n" +
	"Call-ID: 1qAAdc_RECnDVggZPZkacw..\r\n" +
	"CSeq: 1 INVITE\r\n" +
	"Allow: INVITE, ACK, CANCEL, BYE, NOTIFY, REFER, MESSAGE, OPTIONS, INFO, SUBSCRIBE\r\n" +
	"Content-Type: application/sdp\r\n" +
	"Authorization: DIGEST username=\"BrentTC20301Ext\", realm=\"wjking.co.uk\", nonce=\"BroadWorksXkgau305aTjibt11BW\", qop=auth, cnonce=\"KgKrItTGZ9xXJyC\", nc=00000001, uri=\"sip:wjking.co.uk\", response=\"5b6330fe13556fff0b0f91fbfabe5076\", algorithm=MD5\r\n" +
	//"WWW-Authenticate: DIGEST qop=\"auth\",nonce=\"BroadWorksXkgau305aTjibt11BW\",realm=\"wjking.co.uk\",algorithm=MD5\r\n"+
	"User-Agent: Z 5.4.8 rv2.10.11.4\r\n" +
	"Allow-Events: presence, kpml, talk\r\n" +
	"Content-Length: 574\r\n" +
	"\r\n\r\n" +
	"v=0\r\n" +
	"o=Z 1602241428579 1 IN IP4 127.0.0.1\r\n" +
	"s=Z\r\n" +
	"c=IN IP4 127.0.0.1\r\n" +
	"t=0 0\r\n" +
	"m=audio 8000 RTP/AVP 106 9 3 111 0 8 97 110 112 98 101 100 99 102\r\n" +
	"a=rtpmap:106 opus/48000/2\r\n" +
	"a=fmtp:106 minptime=20; useinbandfec=1\r\n" +
	"a=rtpmap:111 speex/16000\r\n" +
	"a=rtpmap:97 iLBC/8000\r\n" +
	"a=fmtp:97 mode=20\r\n" +
	"a=rtpmap:110 speex/8000\r\n" +
	"a=rtpmap:112 speex/32000\r\n" +
	"a=rtpmap:98 telephone-event/48000\r\n" +
	"a=fmtp:98 0-16\r\n" +
	"a=rtpmap:101 telephone-event/8000\r\n" +
	"a=fmtp:101 0-16\r\n" +
	"a=rtpmap:100 telephone-event/16000\r\n" +
	"a=fmtp:100 0-16\r\n" +
	"a=rtpmap:99 telephone-event/32000\r\n" +
	"a=fmtp:99 0-16\r\n" +
	"a=rtpmap:102 G726-32/8000\r\n" +
	"a=sendrecv\r\n'"
