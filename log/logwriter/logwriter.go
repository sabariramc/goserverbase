// Package logwriter provides interfaces for writing log messages.
package logwriter

import (
	"context"

	"github.com/sabariramc/goserverbase/v6/log/message"
)

// ChanneledLogWriter represents an interface for a log writer that consumes log messages from a channel.
//
// This interface defines methods for starting the log writer and writing log messages.
// It also provides a method to retrieve the buffer size of the channel.
type ChanneledLogWriter interface {
	// Start starts the log writer, consuming log messages from the given channel.
	Start(chan message.MuxLogMessage)

	// WriteMessage writes a log message to the log writer.
	WriteMessage(context.Context, *message.LogMessage) error

	// GetBufferSize returns the buffer size of the channel used by the log writer.
	GetBufferSize() int
}

// LogWriter represents a generic interface for writing log messages.
//
// This interface defines a single method for writing log messages.
type LogWriter interface {
	// WriteMessage writes a log message to the log writer.
	WriteMessage(context.Context, *message.LogMessage) error
}
