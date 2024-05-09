package utils

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

var letterPoolMap = map[string]string{
	"111": "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
	"110": "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
	"101": "abcdefghijklmnopqrstuvwxyz0123456789",
	"100": "abcdefghijklmnopqrstuvwxyz",
	"011": "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
	"010": "ABCDEFGHIJKLMNOPQRSTUVWXYZ",
	"001": "0123456789",
}

const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.New(rand.NewSource(time.Now().UnixNano()))
var mu sync.Mutex

type LetterPollConfig struct {
	Int       bool
	UpperCase bool
	LowerCase bool
}

var defaultConfig = LetterPollConfig{
	LowerCase: true,
	UpperCase: true,
	Int:       true,
}

type LetterPoolConfig func(*LetterPollConfig)

func NoInt() LetterPoolConfig {
	return func(c *LetterPollConfig) {
		c.Int = false
	}
}

func NoUpperCase() LetterPoolConfig {
	return func(c *LetterPollConfig) {
		c.UpperCase = false
	}
}

func NoLowerCase() LetterPoolConfig {
	return func(c *LetterPollConfig) {
		c.LowerCase = false
	}
}

func GetLetterPoolConfig(options ...LetterPoolConfig) LetterPollConfig {
	config := defaultConfig
	for _, fu := range options {
		fu(&config)
	}
	return config
}

func GetLetterPool(c LetterPollConfig) string {
	letterPoolIdx := []byte{'1', '1', '1'}
	if !c.LowerCase {
		letterPoolIdx[0] = '0'
	}
	if !c.UpperCase {
		letterPoolIdx[1] = '0'
	}
	if !c.Int {
		letterPoolIdx[2] = '0'
	}
	res, ok := letterPoolMap[string(letterPoolIdx)]
	if !ok {
		return ""
	}
	return res
}

// RandomStringGenerator generates a random string of n characters, the generated string will of form ^[A-Za-z0-9]{n}$ by default, charset is customizable using options
type RandomStringGenerator struct {
	letterPool string
	src        *rand.Rand
	lock       sync.Mutex
}

func NewRandomNumberGenerator(options ...LetterPoolConfig) *RandomStringGenerator {
	config := GetLetterPoolConfig(options...)
	letterPool := GetLetterPool(config)
	return &RandomStringGenerator{
		letterPool: letterPool,
		src:        rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (r *RandomStringGenerator) String(n int) string {
	if len(r.letterPool) == 0 {
		return ""
	}
	b := make([]byte, n)
	r.lock.Lock()
	defer r.lock.Unlock()
	// A src.Int63() generates 63 random biGenerts, enough for letterIdxMax characters!
	for i, cache, remain := n-1, r.src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = r.src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(r.letterPool) {
			b[i] = r.letterPool[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return string(b)
}

var defaultRandomNumberGenerator = NewRandomNumberGenerator()

func GenerateRandomString(n int) string {
	return defaultRandomNumberGenerator.String(n)
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
