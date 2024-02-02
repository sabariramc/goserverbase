package utils_test

import (
	"sync"
	"testing"

	"github.com/sabariramc/goserverbase/v5/utils"
	"gotest.tools/assert"
)

func TestRandomStringGenerator(t *testing.T) {
	totalN := 1000000
	ch := make(chan string, totalN)
	var wg sync.WaitGroup
	conncurrencyFactor := 10000
	for i := 0; i < conncurrencyFactor; i++ {
		wg.Add(1)
		go func() {
			for j := 0; j < totalN/conncurrencyFactor; j++ {
				ch <- utils.GenerateID(20, "sch_")
			}
			wg.Done()
		}()
	}
	wg.Add(1)
	duplicateCount := 0
	go func() {
		idSet := make(map[string]bool, totalN)
		total := 0
		for id := range ch {
			if _, ok := idSet[id]; ok {
				duplicateCount++
			}
			idSet[id] = true
			total++
			if total == totalN {
				break
			}
		}
		wg.Done()
	}()
	wg.Wait()
	assert.Equal(t, duplicateCount, 0)
}

var benchmarkRes string

func BenchmarkRandomStringGenerator(b *testing.B) {
	var r string
	for i := 0; i < b.N; i++ {
		r = utils.GenerateID(20, "sch_")
	}
	benchmarkRes = r
}
