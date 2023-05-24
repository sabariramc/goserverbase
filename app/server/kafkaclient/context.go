package kafkaclient

import (
	"github.com/sabariramc/goserverbase/v2/log"
	"github.com/sabariramc/goserverbase/v2/utils"
)

func (k *KafkaClient) GetCorrelationParams(headers map[string]string) *log.CorrelationParam {
	correlation := log.GetDefaultCorrelationParam(k.c.ServiceName)
	err := utils.LenientJsonTransformer(headers, correlation)
	if err != nil {
		return log.GetDefaultCorrelationParam(k.c.ServiceName)
	}
	return correlation
}

func (k *KafkaClient) GetCustomerId(headers map[string]string) *log.CustomerIdentifier {
	customerId := &log.CustomerIdentifier{}
	err := utils.LenientJsonTransformer(headers, customerId)
	if err != nil {
		return &log.CustomerIdentifier{}
	}
	return customerId
}
