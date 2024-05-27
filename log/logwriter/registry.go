package logwriter

import (
	"fmt"
)

// registry stores registered log writers.
var registry map[string]LogWriter

// GetLogWriter retrieves a log writer by its name from the registry.
//
// If a log writer with the specified name exists in the registry, it is returned.
// If no log writer with the specified name is found, nil is returned.
func GetLogWriter(name string) LogWriter {
	l, ok := registry[name]
	if !ok {
		return nil
	}
	return l
}

// AddLogWriter adds a log writer to the registry with the specified name.
//
// If a log writer with the same name already exists in the registry, an error is returned.
// Otherwise, the log writer is added to the registry.
func AddLogWriter(name string, l LogWriter) error {
	_, ok := registry[name]
	if ok {
		return fmt.Errorf("'%v' is a duplicate index name", name)
	}
	registry[name] = l
	return nil
}

// init initializes the registry by adding default log writers.
//
// Default log writers include "CONSOLE" and "JSONL" log writers.
func init() {
	AddLogWriter("CONSOLE", NewConsoleWriter())
	AddLogWriter("JSONL", NewJSONLConsoleWriter(DefaultLogMapper))
}
