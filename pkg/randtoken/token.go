package randtoken

import (
	"crypto/rand"
	"encoding/binary"
)

// https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // Number of letter indices fitting in 63 random bits
)

// New generates a random token of length n with 'a-zA-Z0-9' alphabet
func New(n int) string {
	b := make([]byte, n)
	// A 63-bit value provides enough random bits for letterIdxMax letters.
	for i, cache, remain := n-1, randomInt63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = randomInt63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(alphabet) {
			b[i] = alphabet[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

// randomInt63 returns a uniformly distributed, nonnegative 63-bit random number.
func randomInt63() int64 {
	var b [8]byte
	// Read fills the buffer or terminates the process; never use a predictable fallback.
	rand.Read(b[:])
	return int64(binary.LittleEndian.Uint64(b[:]) >> 1)
}
