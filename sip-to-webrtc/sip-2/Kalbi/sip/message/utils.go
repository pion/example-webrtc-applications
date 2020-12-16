package message

import (
	"fmt"
	"Kalbi/log"
	"math/rand"
)

//GenerateBranchId generates a new branch ID
func GenerateBranchId() string {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		log.Log.Fatal(err)
	}
	uuid := fmt.Sprintf("%x-%x", b[0:4], b[4:6])
	return "z9hG4bK-" + uuid
}

//GenerateNewCallID generates new Call ID
func GenerateNewCallID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		log.Log.Fatal(err)
	}
	uuid := fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	return uuid

}
