package kafkaconsumer

import "context"

func (k *KafkaConsumerServer) Shutdown(ctx context.Context) error {
	defer k.wg.Done()
	k.wg.Add(1)
	k.shutdownPoll()
	k.client.Close(ctx)
	k.shutdown()
	return nil
}
