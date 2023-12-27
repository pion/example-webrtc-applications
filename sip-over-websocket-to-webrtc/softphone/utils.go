package softphone

import (
	"crypto/md5" //nolint
	"fmt"

	"github.com/google/uuid"
)

// SIPInfoResponse ...
type SIPInfoResponse struct {
	Username           string `json:"username"`
	Password           string `json:"password"`
	AuthorizationID    string `json:"authorizationId"`
	Domain             string `json:"domain"`
	OutboundProxy      string `json:"outboundProxy"`
	Transport          string `json:"transport"`
	Certificate        string `json:"certificate"`
	SwitchBackInterval int    `json:"switchBackInterval"`
}

func generateResponse(username, password, realm, method, uri, nonce string) string { // ONLY REGISTRATION WITH QOP=AUTH !
	ha1 := md5.Sum([]byte(fmt.Sprintf("%s:%s:%s", username, realm, password)))                                                 //nolint
	ha2 := md5.Sum([]byte(fmt.Sprintf("%s:%s", method, uri)))                                                                  //nolint
	response := md5.Sum([]byte(fmt.Sprintf("%x:%s:00000001:%s:auth:%x", ha1, nonce, "0e6758e1adfccffbd0ad9ffdde3ef655", ha2))) //nolint

	return fmt.Sprintf("%x", response)
}

func generateAuthorization(sipInfo SIPInfoResponse, method, nonce string) string {
	return fmt.Sprintf(
		`Digest username="%s",realm="%s",nonce="%s",uri="sip:%s",response="%s",algorithm=MD5,cnonce="%s",qop=auth,nc=00000001`,
		sipInfo.Username, sipInfo.Domain, nonce, sipInfo.Domain,
		generateResponse(sipInfo.Username, sipInfo.Password, sipInfo.Domain, method, "sip:"+sipInfo.Domain, nonce), "0e6758e1adfccffbd0ad9ffdde3ef655",
	)
}

func generateProxyAuthorization(sipInfo SIPInfoResponse, method, targetUser, nonce string) string {
	return fmt.Sprintf(
		`Digest username="%s", realm="%s", nonce="%s", uri="sip:%s@%s", response="%s",algorithm=MD5,cnonce="%s",qop=auth,nc=00000001`,
		sipInfo.AuthorizationID, sipInfo.Domain, nonce, targetUser, sipInfo.Domain,
		generateResponse(sipInfo.AuthorizationID, sipInfo.Password, sipInfo.Domain, method, "sip:"+targetUser+"@"+sipInfo.Domain, nonce), "0e6758e1adfccffbd0ad9ffdde3ef655",
	)
}

func branch() string {
	return "z9hG4bK" + uuid.New().String()
}
