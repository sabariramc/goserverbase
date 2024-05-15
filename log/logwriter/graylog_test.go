package logwriter_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sabariramc/goserverbase/v6/log"
	"github.com/sabariramc/goserverbase/v6/log/logwriter"
	"github.com/sabariramc/goserverbase/v6/testutils"
	"gopkg.in/Graylog2/go-gelf.v2/gelf"
	"gotest.tools/assert"
)

func TestGraylog(t *testing.T) {
	testutils.LoadEnv("../../.env")
	config := testutils.NewConfig()
	graylogAddr := fmt.Sprintf("%v:%v", config.Graylog.Address, config.Graylog.Port)

	if graylogAddr != "" {
		// If using UDP
		gelfWriter, err := gelf.NewUDPWriter(graylogAddr)
		// If using TCP
		// gelfWriter, err := gelf.NewTCPWriter(graylogAddr)
		assert.NilError(t, err, "Error createing udp writer")
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
		assert.NilError(t, err, "Error writing message")
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
		assert.NilError(t, err, "Error writing message")
		if err != nil {
			t.Fatalf(err.Error())
		}
	}
}

func TestGraylogWrapper(t *testing.T) {
	testutils.LoadEnv("../../.env")
	config := testutils.NewConfig()
	grey, err := logwriter.NewGraylogUDP(logwriter.NewConsoleWriter(), config.Graylog)
	assert.NilError(t, err)
	grey.WriteMessage(context.Background(), &log.LogMessage{
		Message: "Test Wrapper" + uuid.NewString(),
	})
}
