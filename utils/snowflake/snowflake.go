// Package snowflake is a implement of https://en.wikipedia.org/wiki/Snowflake_ID
package snowflake

import (
	"fmt"
	"sync"
	"time"
)

type ID int64

type Snowflake struct {
	machineID      int64
	sequenceNumber int64
	lock           sync.Mutex
}

const maxSequence int64 = 0xFFF
const maxMachineID int64 = 0x3FF
const maxTimestamp int64 = 0x1FFFFFFFFFF

func New(machineID, sequenceNumber int64) (*Snowflake, error) {
	if 0 > machineID || machineID > maxMachineID {
		return nil, fmt.Errorf("invalid machine id for Snowflake")
	}
	if 0 > sequenceNumber {
		return nil, fmt.Errorf("invalid initial sequence for Snowflake")
	}
	if sequenceNumber > maxSequence {
		sequenceNumber = 0
	}
	return &Snowflake{
		machineID:      machineID << 12,
		sequenceNumber: sequenceNumber,
	}, nil
}

func (s *Snowflake) Stats() (int64, int64) {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.machineID, s.sequenceNumber
}

func (s *Snowflake) GenerateID() (ID, error) {
	var sequenceNumber int64
	s.lock.Lock()
	ts := time.Now().UnixMilli()
	if s.sequenceNumber == maxSequence {
		s.sequenceNumber = 0
	}
	sequenceNumber, s.sequenceNumber = s.sequenceNumber, s.sequenceNumber+1
	s.lock.Unlock()
	if ts > maxTimestamp {
		return -1, fmt.Errorf("timestamp exceed max limit")
	}
	return ID(ts<<22 | s.machineID | sequenceNumber), nil
}
