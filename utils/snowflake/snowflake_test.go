package snowflake_test

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/sabariramc/goserverbase/v5/utils/snowflake"
	"gotest.tools/assert"
)

func TestSnowflake(t *testing.T) {
	s, err := snowflake.New(1, 0)
	assert.NilError(t, err)
	fmt.Printf("%b\n", time.Now().UnixMilli())
	for i := 0; i < 10000; i++ {
		id, err := s.ID()
		assert.NilError(t, err)
		fmt.Printf("%b\n", id)
	}
}

func TestSnowflakeTimer(t *testing.T) {
	s, err := snowflake.New(1, 0)
	assert.NilError(t, err)
	st := time.Now()
	for i := 0; i < 10000; i++ {
		_, err := s.ID()
		assert.NilError(t, err)
	}
	assert.Assert(t, 6 >= time.Since(st).Milliseconds())
}

func TestSnowflakeDuplicacy(t *testing.T) {
	s, err := snowflake.New(1, 0)
	assert.NilError(t, err)
	totalN := 1000000
	ch := make(chan int64, totalN)
	var wg sync.WaitGroup
	conncurrencyFactor := 10000
	for i := 0; i < conncurrencyFactor; i++ {
		wg.Add(1)
		go func() {
			for j := 0; j < totalN/conncurrencyFactor; j++ {
				id, err := s.ID()
				assert.NilError(t, err)
				ch <- id
			}
			wg.Done()
		}()
	}
	wg.Add(1)
	duplicateCount := 0
	go func() {
		idSet := make(map[int64]bool, totalN)
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

func BenchmarkSnowflake(b *testing.B) {
	var v int64
	s, err := snowflake.New(1, 0)
	assert.NilError(b, err)
	var goprocs = runtime.GOMAXPROCS(0)
	for i := 1; i < 1000; i += 50 {
		b.Run(fmt.Sprintf("goroutines-%d", i*goprocs), func(b *testing.B) {
			b.SetParallelism(i)
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					_, err := s.ID()
					assert.NilError(b, err)
					v += 1
				}
			})
		})
	}
}
