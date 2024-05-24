package errors_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"testing"

	"github.com/sabariramc/goserverbase/v6/errors"
)

func Example() {
	err := &errors.CustomError{"goserverbase.test.error", "test error message", "test error data", "test error description", false}
	fmt.Println(err)
	herr := errors.HTTPError{StatusCode: http.StatusConflict, CustomError: err}
	data, _ := json.MarshalIndent(herr, "", "    ")
	fmt.Println(string(data))
	//Output:
	//{
	//     "errorCode": "goserverbase.test.error",
	//     "errorMessage": "test error message",
	//     "errorData": "test error data",
	//     "errorDescription": "test error description"
	// }
	//{
	//     "errorCode": "goserverbase.test.error",
	//     "errorMessage": "test error message",
	//     "errorData": "test error data",
	//     "errorDescription": "test error description",
	//     "statusCode": 409
	// }
}

func TestCustomeErrorPrint(t *testing.T) {
	ctx := GetCorrelationContext()
	err := &errors.CustomError{"com.sabariram.test.error", "test error message", "test error data", "test error description", false}
	TestLogger.Error(ctx, "error", err)
}

const (
	start = 1 // actual = start  * goprocs
	end   = 8 // actual = end    * goprocs
	step  = 1
)

var goprocs = runtime.GOMAXPROCS(0) // 8

var benchmarkRes string

func BenchmarkCustomError(b *testing.B) {
	var str string
	err := fmt.Errorf("test error")
	err = &errors.CustomError{"com.sabariram.test.error", "test error message", "test error data", "test error description", false}
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
