// Package snowflake is a implement of https://en.wikipedia.org/wiki/Snowflake_ID
package snowflake

import (
	"fmt"
	"sync"
	"time"
)

type Snowflake struct {
	machineID      int64
	sequenceNumber int64
	lock           sync.Mutex
	ch             chan int64
}

const maxSequence int64 = 0xFFF
const maxMachineID int64 = 0x3FF
const maxTimestamp int64 = 0x1FFFFFFFFFF

func New(machineID, sequenceNumber int64) (*Snowflake, error) {
	if machineID > maxMachineID {
		return nil, fmt.Errorf("invalid machine id for Snowflake")
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

func (s *Snowflake) ID() (int64, error) {
	s.lock.Lock()
	if s.sequenceNumber == maxSequence {
		s.sequenceNumber = 0
	}
	sequenceNumber := s.sequenceNumber
	s.sequenceNumber++
	ts := time.Now().UnixMilli()
	s.lock.Unlock()
	if ts > maxTimestamp {
		return -1, fmt.Errorf("timestamp exceed max limit")
	}
	res := ts<<22 | s.machineID | sequenceNumber
	return res, nil
}
