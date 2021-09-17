package utils

import (
	mathrand "math/rand"
	"sync"
	"time"
)

var (
	letters    = []rune("abcdefgjijklmnopqrstuvwxyz0123456789")
	lettersLen = len(letters)
	rand       = mathrand.New(mathrand.NewSource(time.Now().UTC().UnixNano()))
	mutex      sync.Mutex
)

func RandomStr(len int) string {
	b := make([]rune, len)
	for i := range b {
		b[i] = letters[intn(lettersLen)]
	}

	return string(b)
}

func intn(max int) int {
	mutex.Lock()
	defer mutex.Unlock()
	return rand.Intn(max)
}
