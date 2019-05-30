package oauth2clients

import "crypto/rand"

func mustCryptoRandRead(b []byte) {
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
}
