package kafkaconsumer

import "context"

func (k *KafkaConsumerServer) Shutdown(ctx context.Context) error {
	defer k.shutdownWG.Done()
	k.shutdownPoll()
	k.requestWG.Wait()
	k.shutdown()
	k.client.Close(ctx)
	return nil
}
