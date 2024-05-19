package kafkaconsumer

import (
	"context"
	"os"
	"time"
)

// HealthCheckMonitor starts a health check monitor that periodically runs health checks.
func (k *KafkaConsumerServer) HealthCheckMonitor(ctx context.Context) {
	timeoutContext, _ := context.WithTimeout(ctx, time.Second*time.Duration(k.c.healthCheckInSec))
	defer k.log.Warning(ctx, "Health check monitor stopped", nil)
	for {
		select {
		case <-ctx.Done():
			return
		case <-timeoutContext.Done():
			err := k.RunHealthCheck(ctx)
			if err != nil {
				deleteErr := os.Remove(k.c.healthFilePath)
				if deleteErr != nil {
					k.log.Error(ctx, "error deleting health file", deleteErr)
				}
				k.log.Emergency(ctx, "health check failed", err, nil)
			}
			timeoutContext, _ = context.WithTimeout(ctx, time.Second*time.Duration(k.c.healthCheckInSec))
		}
	}
}

// HealthCheck runs a health check on the Kafka consumer server.
func (k *KafkaConsumerServer) HealthCheck(ctx context.Context) error {
	k.client.Stats()
	return nil
}

// StatusCheck runs a status check on the Kafka consumer server.
func (k *KafkaConsumerServer) StatusCheck(ctx context.Context) (any, error) {
	return k.client.Stats(), nil
}
