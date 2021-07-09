package vaxcert

import (
	"math/rand"
	"time"
)

var allowedChars = []rune("QWERTYUIOPASDFGHJKLZXCVBNM1234567890")

func init() {
	rand.Seed(time.Now().UnixNano())
}

func generateCertCode(n int) string {
	res := make([]rune, n)
	for i := range res {
		res[i] = allowedChars[rand.Intn(len(allowedChars))]
	}

	return string(res)
}
