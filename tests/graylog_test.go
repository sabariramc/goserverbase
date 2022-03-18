package tests

import (
	"testing"
	"time"

	"gopkg.in/Graylog2/go-gelf.v2/gelf"
)

func TestGraylog(t *testing.T) {
	graylogAddr := "localhost:12202"

	if graylogAddr != "" {
		// If using UDP
		gelfWriter, err := gelf.NewUDPWriter(graylogAddr)
		// If using TCP
		// gelfWriter, err := gelf.NewTCPWriter(graylogAddr)
		if err != nil {
			t.Fatalf(err.Error())
		}
		_ = gelfWriter.WriteMessage(&gelf.Message{
			Version:  "1.1",
			Host:     "localhost",
			Short:    "Hello world",
			Full:     "Whats my name",
			TimeUnix: float64(time.Now().Unix()),
			Level:    6,
			Extra: map[string]interface{}{
				"x-correlation-id": "fafasfsadf",
			},
		})

	}

	// From here on out, any calls to log.Print* functions
	// will appear on stdout, and be sent over UDP or TCP to the
	// specified Graylog2 server.

	// ...
}
