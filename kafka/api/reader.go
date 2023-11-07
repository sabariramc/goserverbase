package api

import (
	"context"

	"github.com/sabariramc/goserverbase/v3/log"
	"github.com/segmentio/kafka-go"
)

type Reader struct {
	*kafka.Reader
	log log.Logger
}

func NewReader(ctx context.Context, r *kafka.Reader, log log.Logger) *Reader {
	return &Reader{
		Reader: r,
		log:    log,
	}
}
