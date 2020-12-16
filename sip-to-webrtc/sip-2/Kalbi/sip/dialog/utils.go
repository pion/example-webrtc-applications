package dialog

import (
	"math/rand"
)

//GenerateDialogId creates new dialog ID
func GenerateDialogId() int32 {
	return rand.Int31()
}
