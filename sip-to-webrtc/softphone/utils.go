package softphone

import (
	"crypto/md5" //nolint
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
)

// SIPInfoResponse ...
type SIPInfoResponse struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Domain string `json:"domain"`
	WebsocketURL string `json:"websocket_url"`
	Certificate string `json:"certificate"`
	SwitchBackInterval int `json:"switchBackInterval"`
}

type AuthInfo struct {
	AuthType string  `json:"auth_type"`
	Realm string `json:"realm"`
	Nonce string `json:"nonce"`
	Uri string `json:"uri"`
	Algorithm string `json:"algorithm"`
	Qop string `json:"qop"`
	Method string `json:"method"`
	Cnonce string `json:"cnonce"`
	NonceCount string `json:"nonce_count"`
}
func generateAuthorization(sipInfo SIPInfoResponse, ai AuthInfo) (ret string) {
	var HA1, HA2, response [16]byte
	switch ai.Algorithm {
	case "MD5":
		//HA1 = MD5(username:realm:password)
		HA1 = md5.Sum([]byte(fmt.Sprintf("%s:%s:%s", sipInfo.Username, ai.Realm, sipInfo.Password)))
		break
	case "MD5-sess":
		//HA1 = MD5(MD5(username:realm:password):nonce:cnonce)
		HA1 = md5.Sum([]byte(fmt.Sprintf("%s:%s:%s", sipInfo.Username, ai.Realm, sipInfo.Password)))
		HA1 = md5.Sum([]byte(fmt.Sprintf("%s:%s:%s", HA1, ai.Nonce, ai.Cnonce)))
		break
	default:
		panic("NO ALGORITHM FOUND ! ")
	}
	switch ai.Qop {
	case "auth":
		// HA2 = MD5(A2) = MD5(method:digestURI).
		// Response = MD5(HA1:nonce:nonceCount:credentialsNonce:qop:HA2).
		HA2 = md5.Sum([]byte(fmt.Sprintf("%s:%s", ai.Method, ai.Uri)))
		response = md5.Sum([]byte(fmt.Sprintf("%x:%s:%s:%s:%s:%x", HA1, ai.Nonce,ai.NonceCount, ai.Cnonce,ai.Qop , HA2)))
		ret = fmt.Sprintf(
			`Digest username="%s", realm="%s", nonce="%s", uri="%s", response="%x", algorithm=%s, cnonce="%s", qop=%s, nc=%s`,
			sipInfo.Username, ai.Realm, ai.Nonce, ai.Uri, response,ai.Algorithm, ai.Cnonce, ai.Qop, ai.NonceCount,
		)
		break
	case "auth-int":
		//TODO : DO THIS PART !
		// HA2 = MD5(A2) = MD5(method:digestURI:MD5(entityBody)).
		// Response = MD5(HA1:nonce:nonceCount:credentialsNonce:qop:HA2).
		break
	default:
		// HA2 = MD5(A2) = MD5(method:digestURI).
		// Response = MD5(HA1:nonce:HA2).
		HA2 = md5.Sum([]byte(fmt.Sprintf("%s:%s", ai.Method, ai.Uri)))
		response = md5.Sum([]byte(fmt.Sprintf("%x:%s:%x", HA1, ai.Nonce, HA2)))
		ret = fmt.Sprintf(
			`Digest username="%s", realm="%s", nonce="%s", uri="%s", response="%x", algorithm=%s`,
			sipInfo.Username, ai.Realm, ai.Nonce, ai.Uri, response,ai.Algorithm,
		)
		break
	}

	return
}

func GetAuthInfo(wwwAuthenticate string) (ai AuthInfo) {
	ai.Cnonce = RandomString() // Generate RANDOM PLEASE
	ai.NonceCount = "00000001"
	splitedWA := strings.Split(wwwAuthenticate, ",")
	for i := 0; i < len(splitedWA); i++ {
		if strings.Contains(splitedWA[i], "realm") {
			ai.Realm = strings.ReplaceAll(strings.Split(splitedWA[i], "=")[1],"\"", "")
		} else if strings.Contains(splitedWA[i], "nonce") {
			ai.Nonce = strings.ReplaceAll(strings.Split(splitedWA[i], "=")[1],"\"", "")
		} else if strings.Contains(splitedWA[i], "algorithm") {
			ai.Algorithm = strings.ReplaceAll(strings.Split(splitedWA[i], "=")[1],"\"", "")
		} else if strings.Contains(splitedWA[i], "qop") {
			ai.Qop = strings.ReplaceAll(strings.Split(splitedWA[i], "=")[1],"\"", "")
		}
	}
	return
}

func RandomString() string { // GENERATE A RANDOM CNONCE
	rand.Seed(time.Now().UnixNano())
	min := 5
	max := 30
	n := rand.Intn(max - min + 1) + min
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}


func branch() string {
	return "z9hG4bK" + uuid.New().String()
}