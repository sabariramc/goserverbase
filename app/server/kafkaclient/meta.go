package kafkaclient

import (
	"context"
	"os"
	"time"
)

// HealthCheckMonitor starts a health check monitor that periodically runs health checks.
func (k *KafkaClient) HealthCheckMonitor(ctx context.Context) {
	timeoutContext, _ := context.WithTimeout(ctx, time.Second*time.Duration(k.c.HealthCheckInterval))
	defer k.log.Warning(ctx, "Health check monitor stopped", nil)
	for {
		select {
		case <-ctx.Done():
			return
		case <-timeoutContext.Done():
			err := k.RunHealthCheck(ctx)
			if err != nil {
				deleteErr := os.Remove(k.c.HealthCheckResultPath)
				if deleteErr != nil {
					k.log.Error(ctx, "error deleting health file", deleteErr)
				}
				k.log.Emergency(ctx, "health check failed", err, nil)
			}
			timeoutContext, _ = context.WithTimeout(ctx, time.Second*time.Duration(k.c.HealthCheckInterval))
		}
	}
}

// HealthCheck runs a health check on the Kafka consumer server.
func (k *KafkaClient) HealthCheck(ctx context.Context) error {
	k.client.Stats()
	return nil
}

// StatusCheck runs a status check on the Kafka consumer server.
func (k *KafkaClient) StatusCheck(ctx context.Context) (any, error) {
	return k.client.Stats(), nil
}
