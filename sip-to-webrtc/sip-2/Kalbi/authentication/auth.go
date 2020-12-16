package authentication

import ("crypto/md5"
        "encoding/hex")

//MD5Challange returns computed challenge
func MD5Challange(username string, realm string, password string, uri string, nonce string, cnonce string, nc string, qop string, method string) string {
	first := md5.Sum([]byte(username + ":" + realm + ":" + password))
	second := md5.Sum([]byte(method + ":" + uri))
	ha1 :=  hex.EncodeToString(first[:])
	ha2 :=  hex.EncodeToString(second[:])
	third := md5.Sum([]byte(ha1 + ":" + nonce + ":" + nc + ":" + cnonce + ":" + qop + ":" + ha2))
    ha3 := hex.EncodeToString(third[:])
	return ha3
}