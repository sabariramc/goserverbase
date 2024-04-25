package utils

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.New(rand.NewSource(time.Now().UnixNano()))
var mu sync.Mutex

// GenerateRandomString generates a random string of n characters, the generated string will of form ^[A-Za-z0-9]{n}$
func GenerateRandomString(n int) string {
	b := make([]byte, n)
	mu.Lock()
	// A src.Int63() generates 63 random biGenerts, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	mu.Unlock()
	return string(b)
}

/*
GenerateID generates a id with the prefix and totalLength

x := GenerateID(20, "cust_")

In the above example
len(x) would be equal to 20, prefix "cust_" of length 5 and the remaining slots(15) will be filled by random characters regex: ^cust_[a-zA-Z0-9]{15}$
*/
func GenerateID(totalLength int, prefix string) string {
	n := totalLength - len(prefix)
	return fmt.Sprintf("%v%v", prefix, GenerateRandomString(n))
}
