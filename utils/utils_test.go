package utils_test

import (
	"fmt"
	"runtime"
	"sync"
	"testing"

	"github.com/sabariramc/goserverbase/v6/utils"
	"gotest.tools/assert"
)

func TestGetHash(t *testing.T) {
	val := "3edcRFV5tgb"
	assert.Equal(t, utils.GetHash(val), utils.GetHash(val))
}

const (
	start = 1   // actual = start  * goprocs
	end   = 600 // actual = end    * goprocs
	step  = 50
)

var goprocs = runtime.GOMAXPROCS(0) // 8

// https://perf.golang.org/search?q=upload:20190819.3
func BenchmarkChanWrite(b *testing.B) {
	var v int64
	ch := make(chan int, 1)
	ch <- 1
	for i := start; i < end; i += step {
		b.Run(fmt.Sprintf("goroutines-%d", i*goprocs), func(b *testing.B) {
			b.SetParallelism(i)
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					<-ch
					v += 1
					ch <- 1
				}
			})
		})
	}
}

// https://perf.golang.org/search?q=upload:20190819.2
func BenchmarkMutexWrite(b *testing.B) {
	var v int64
	mu := sync.Mutex{}
	for i := start; i < end; i += step {
		b.Run(fmt.Sprintf("goroutines-%d", i*goprocs), func(b *testing.B) {
			b.SetParallelism(i)
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					mu.Lock()
					v += 1
					mu.Unlock()
				}
			})
		})
	}
}
