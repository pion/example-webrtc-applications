package test

import (
	"fmt"
	"Kalbi/sip/message"
	"Kalbi/authentication"
	"testing"
)

func TestSIPParser(t *testing.T) {
	byteMsg := []byte(msg)
	x := message.Parse(byteMsg)

	fmt.Println(authentication.MD5Challange("02922401513", "thevoicefactory.co.uk", "Chuckie93@", "sip:thevoicefactory.co.uk", "BroadWorksXiv8la38lT5rbw3uBW", "slmssmsf", "00000001", "auth", "REGISTER"))
	
	if string(x.Req.Method) != "INVITE" {
		t.Error("Method line not parsed")
	}
}
