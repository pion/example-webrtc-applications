package softphone

import (
	"crypto/md5"
	"fmt"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type SIPInfoResponse struct {
	Username string `json:"username"`
	Password string `json:"password"`
	AuthorizationId string `json:"authorizationId"`
	Domain string `json:"domain"`
	OutboundProxy string `json:"outboundProxy"`
	Transport string `json:"transport"`
	Certificate string `json:"certificate"`
	SwitchBackInterval int `json:"switchBackInterval"`
}

func generateResponse(username, password, realm, method, uri, nonce string) string { //ONLY REGISTRATION WITH QOP=AUTH !
	ha1 := md5.Sum([]byte(fmt.Sprintf("%s:%s:%s", username, realm, password)))
	ha2 := md5.Sum([]byte(fmt.Sprintf("%s:%s", method, uri)))
	//NOT MD5(HA1:nonce:HA2)
	//MD5(HA1:nonce:nonceCount:cnonce:qop:HA2)
	response := md5.Sum([]byte(fmt.Sprintf("%x:%s:00000001:%s:auth:%x", ha1, nonce, "0e6758e1adfccffbd0ad9ffdde3ef655", ha2)))
	return fmt.Sprintf("%x", response)
}
//dfe910a916adb292027e926280325a2c
//Irz887kIoab0JJDbWNDw2
func generateAuthorization(sipInfo SIPInfoResponse, method, nonce string) string {
	//2a1e079b4ccd9d2abf6185ffa0eecf1c // EXPECTED
	//4891c0a879651671c9618eb8757da3d7 // GET
	//fmt.Printf("TEST : |%s|\n", generateResponse("102", "secret", "192.168.1.30", "REGISTER", "sip:192.168.1.30","687f7a9d-8d53-477f-b23e-92bc59daa081"))
	//Digest username="102",realm="192.168.1.30",nonce="687f7a9d-8d53-477f-b23e-92bc59daa081",uri="sip:192.168.1.30",response="2a1e079b4ccd9d2abf6185ffa0eecf1c",algorithm=MD5,cnonce="dfe910a916adb292027e926280325a2c",qop=auth,nc=00000001
	return fmt.Sprintf(
		`Digest username="%s",realm="%s",nonce="%s",uri="sip:%s",response="%s",algorithm=MD5,cnonce="%s",qop=auth,nc=00000001`,
		sipInfo.Username, sipInfo.Domain, nonce, sipInfo.Domain,
		generateResponse(sipInfo.Username, sipInfo.Password, sipInfo.Domain, method, "sip:"+sipInfo.Domain, nonce),"0e6758e1adfccffbd0ad9ffdde3ef655",
	)
}

func generateProxyAuthorization(sipInfo SIPInfoResponse, method, targetUser, nonce string) string {
	return fmt.Sprintf(
		`Digest algorithm=MD5, username="%s", realm="%s", nonce="%s", uri="sip:%s@%s", response="%s"`,
		sipInfo.Username, sipInfo.Domain, nonce, targetUser, sipInfo.Domain,
		generateResponse(sipInfo.Username, sipInfo.Password, sipInfo.Domain, method, "sip:"+targetUser+"@"+sipInfo.Domain, nonce),
	)
}

func branch() string {
	return "z9hG4bK" + uuid.New().String()
}

func configureLog() {
	logLevel := "all"
	if logLevel == "all" {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.FatalLevel)
	}
	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})
}
