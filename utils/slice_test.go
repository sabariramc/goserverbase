package utils_test

import (
	"testing"

	"github.com/sabariramc/goserverbase/v4/utils"
	"gotest.tools/assert"
)

func TestPrepend(t *testing.T) {
	a := []int{1, 2}
	for i := 0; i < 10; i++ {
		a = utils.Prepend[int](a, i)
	}
	assert.DeepEqual(t, a, []int{9, 8, 7, 6, 5, 4, 3, 2, 1, 0, 1, 2})
}

func BenchmarkPrepend(b *testing.B) {
	a := []int{1, 2}
	for i := 0; i < b.N; i++ {
		a = utils.Prepend[int](a, i)
	}

	// for i := start; i < end; i += step {
	// 	b.Run(fmt.Sprintf("goroutines-%d", i*goprocs), func(b *testing.B) {
	// 		b.SetParallelism(i)
	// 		b.RunParallel(func(pb *testing.PB) {
	// 			a := []int{}
	// 			for pb.Next() {
	// 				a = utils.Prepend[int](a, i)
	// 			}
	// 		})
	// 	})
	// }
}
