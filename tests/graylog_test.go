package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sabariramc/goserverbase/utils/testutils"
	"gopkg.in/Graylog2/go-gelf.v2/gelf"
)

func TestGraylog(t *testing.T) {
	config := testutils.NewConfig()
	graylogAddr := fmt.Sprintf("%v:%v", config.Logger.GrayLog.Address, config.Logger.GrayLog.Port)

	if graylogAddr != "" {
		// If using UDP
		gelfWriter, err := gelf.NewUDPWriter(graylogAddr)
		// If using TCP
		// gelfWriter, err := gelf.NewTCPWriter(graylogAddr)
		if err != nil {
			t.Fatalf(err.Error())
		}
		err = gelfWriter.WriteMessage(&gelf.Message{
			Version:  "1.1",
			Host:     "localhost",
			Short:    "FROM container" + uuid.NewString(),
			Full:     "Whats my name" + uuid.NewString(),
			TimeUnix: float64(time.Now().UnixMilli()) / 1000,
			Level:    6,
			Extra: map[string]interface{}{
				"x-correlation-id": "fafasfsadf",
			},
		})
		err = gelfWriter.WriteMessage(&gelf.Message{
			Version:  "1.1",
			Host:     "localhost",
			Short:    "FROM container" + uuid.NewString(),
			Full:     "Whats my name" + uuid.NewString(),
			TimeUnix: float64(time.Now().UnixMilli()) / 1000,
			Level:    6,
			Extra: map[string]interface{}{
				"x-correlation-id": "fafasfsadf",
			},
		})
		if err != nil {
			t.Fatalf(err.Error())
		}
	}

	// From here on out, any calls to log.Print* functions
	// will appear on stdout, and be sent over UDP or TCP to the
	// specified Graylog2 server.

	// ...
}
