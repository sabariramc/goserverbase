package log_test

import (
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/sabariramc/goserverbase/v5/log"
	"github.com/sabariramc/goserverbase/v5/log/logwriter"
	"github.com/shopspring/decimal"
	"gotest.tools/assert"
)

type LogWriter struct {
	valueList []string
	i         int
	ch        chan []string
}

type TestVal struct {
	TestID     string
	IntVal     int64
	DecimalVal decimal.Decimal
	StrVal     string
	BoolVal    bool
	TimeVal    time.Time
}

func GetSampleData() *TestVal {
	val1, _ := decimal.NewFromString("123.1232")
	data := &TestVal{}
	data.TestID = "fasdfa"
	data.StrVal = "value1"
	data.IntVal = 123
	data.BoolVal = true
	data.DecimalVal = val1
	tval, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05+05:30")
	data.TimeVal = tval
	return data
}

func NewLogWriter(ch chan []string) *LogWriter {
	return &LogWriter{
		i:         -1,
		valueList: []string{"0", "1.234", "\"123.1232\"", "true", "abcd", "[\"asdf\",10]", "{\"a\":\"fadsf\",\"b\":10}", "{\"TestID\":\"fasdfa\",\"IntVal\":123,\"DecimalVal\":\"123.1232\",\"StrVal\":\"value1\",\"BoolVal\":true,\"TimeVal\":\"2006-01-02T15:04:05+05:30\"}"},
		ch:        ch,
	}
}

func (c *LogWriter) WriteMessage(ctx context.Context, l *log.LogMessage) error {
	cr := log.GetCorrelationParam(ctx)
	fmt.Printf("[%v] [%v] [%v] [%v] [%v] [%v] [%v] [%v]\n", l.Timestamp, l.LogLevelName, cr.CorrelationID, l.ServiceName, l.ModuleName, l.Message, logwriter.GetLogObjectType(l.LogObject), l.LogObject)
	c.i++
	c.ch <- []string{logwriter.ParseLogObject(l.LogObject, false), c.valueList[c.i]}
	return nil
}

func TestLogWriter(t *testing.T) {
	dec, _ := decimal.NewFromString("123.1232")
	valueList := []any{0, 1.234, dec, true, "abcd", []any{"asdf", 10}, map[string]any{"a": "fadsf", "b": 10}, GetSampleData()}
	ch := make(chan []string, len(valueList))
	lmux := log.NewDefaultLogMux(NewLogWriter(ch))
	log := log.NewLogger(context.Background(), &log.Config{LogLevelName: "DEBUG"}, "Test Logger", lmux, nil)
	for _, v := range valueList {
		log.Debug(context.Background(), "test", v)
	}
	i := 1
	for v := range ch {
		assert.Equal(t, v[0], v[1])
		if i == len(valueList) {
			close(ch)
		}
		i++
	}
}

var goprocs = runtime.GOMAXPROCS(0) // 8

type BenchLogWriter struct {
	i *int
}

func (b *BenchLogWriter) WriteMessage(context.Context, *log.LogMessage) error {
	*b.i = *b.i + 1
	return nil
}

func BenchmarkLogger(b *testing.B) {
	k := 0
	lmux := log.NewDefaultLogMux(&BenchLogWriter{i: &k})
	log := log.NewLogger(context.Background(), &log.Config{LogLevelName: "DEBUG"}, "Test Logger", lmux, nil)
	ctx := context.Background()
	for i := 1; i < 8; i += 2 {
		b.Run(fmt.Sprintf("goroutines-%d", i*goprocs), func(b *testing.B) {
			b.SetParallelism(i)
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					log.Debug(ctx, "fdasfasd", nil)
				}
			})
		})
	}
}

func BenchmarkParseLogObject(b *testing.B) {
	for i := 1; i < 8; i += 2 {
		b.Run(fmt.Sprintf("goroutines-%d", i*goprocs), func(b *testing.B) {
			b.SetParallelism(i)
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					logwriter.ParseObject(GetSampleData(), false)
				}
			})
		})
	}
}
