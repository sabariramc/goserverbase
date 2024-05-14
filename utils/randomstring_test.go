package utils_test

import (
	"fmt"
	"regexp"
	"sync"
	"testing"

	"github.com/sabariramc/goserverbase/v5/utils"
	"gotest.tools/assert"
)

func ExampleGenerateID() {
	x := utils.GenerateID(20, "cust_")
	fmt.Println(len(x))
	match, _ := regexp.MatchString("^cust_[a-zA-Z0-9]{15}$", x)
	fmt.Println(match)
	//Output:
	//20
	//true
}

func ExampleRandomStringGenerator() {
	gen := utils.NewRandomStringGenerator()
	x := gen.Generate(10)
	match, _ := regexp.MatchString("^[a-zA-Z0-9]{10}$", x)
	fmt.Println(match)
	//Output:
	//true
}

func ExampleRandomStringGenerator_onlynumerals() {
	gen := utils.NewRandomStringGenerator(utils.NoLowerCase(), utils.NoUpperCase())
	x := gen.Generate(10)
	match, _ := regexp.MatchString("^[0-9]{10}$", x)
	fmt.Println(match)
	//Output:
	//true
}

func ExampleRandomStringGenerator_onlyuppercase() {
	gen := utils.NewRandomStringGenerator(utils.NoLowerCase(), utils.NoInt())
	x := gen.Generate(10)
	match, _ := regexp.MatchString("^[A-Z]{10}$", x)
	fmt.Println(match)
	//Output:
	//true
}

func ExampleRandomStringGenerator_onlylowercase() {
	gen := utils.NewRandomStringGenerator(utils.NoUpperCase(), utils.NoInt())
	x := gen.Generate(10)
	match, _ := regexp.MatchString("^[a-z]{10}$", x)
	fmt.Println(match)
	//Output:
	//true
}

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

func BenchmarkRandomStringGenerator(b *testing.B) {
	var v int64
	for i := 0; i < b.N; i++ {
		utils.GenerateID(20, "sch_")
		v++
	}
}

func BenchmarkRandomStringGeneratorParallel(b *testing.B) {
	var v int64
	for i := start; i < end; i += step {
		b.Run(fmt.Sprintf("goroutines-%d", i*goprocs), func(b *testing.B) {
			b.SetParallelism(i)
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					utils.GenerateID(20, "sch_")
					v += 1
				}
			})
		})
	}
}
