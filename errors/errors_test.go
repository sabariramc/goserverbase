package errors_test

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/sabariramc/goserverbase/v4/errors"
)

func TestCustomeErrorPrint(t *testing.T) {
	ctx := GetCorrelationContext()
	err := fmt.Errorf("test error")
	err = errors.NewCustomError("com.sabariram.test.error", "test error message", "test error data", "test error description", false, err)
	TestLogger.Error(ctx, "error", err)
}

const (
	start = 1 // actual = start  * goprocs
	end   = 8 // actual = end    * goprocs
	step  = 1
)

var goprocs = runtime.GOMAXPROCS(0) // 8

var benchmarkRes string

func BenchmarkRoutes(b *testing.B) {
	var str string
	err := fmt.Errorf("test error")
	err = errors.NewCustomError("com.sabariram.test.error", "test error message", "test error data", "test error description", false, err)
	for i := start; i < end; i += step {
		b.Run(fmt.Sprintf("goroutines-%d", i*goprocs), func(b *testing.B) {
			b.SetParallelism(i)
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					str = err.Error()
				}
			})
		})
	}
	benchmarkRes = str
}
