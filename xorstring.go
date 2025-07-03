package goxorstring

import (
	"math/rand"
	"time"
)

type XorString struct {
	Key       byte
	Encrypted []byte
}

func NewXorString(s string) *XorString {
	key := GenerateKey()
	encrypted := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		encrypted[i] = s[i] ^ key
	}
	return &XorString{
		Key:       key,
		Encrypted: encrypted,
	}
}

func (x *XorString) Decrypt() string {
	decrypted := make([]byte, len(x.Encrypted))
	for i := 0; i < len(x.Encrypted); i++ {
		decrypted[i] = x.Encrypted[i] ^ x.Key
	}
	return string(decrypted)
}

// GenerateKey a pseudo-random key seeded from compile time
func GenerateKey() byte {
	// get time as string
	t := time.Now().Format("150405") // HHMMSS
	// simulate seed logic from C++ __TIME__
	seed := int(t[5]-'0') + int(t[4]-'0')*10 +
		int(t[3]-'0')*60 + int(t[2]-'0')*600 +
		int(t[1]-'0')*3600 + int(t[0]-'0')*36000
	_rand := rand.New(rand.NewSource(int64(seed)))
	return byte(_rand.Intn(127-1) + 1)
}
