package kafka

import (
	"context"
	"fmt"

	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/plain"
)

const (
	SASL_PLAIN   = "PLAIN"
	SASL_AWS_IAM = "AWS_MSK_IAM"
)

func NewSASL(ctx context.Context, config SASLCredential) (sasl.Mechanism, error) {
	if config.SASLMechanism == SASL_PLAIN {
		mech, ok := config.SASLCredential.(*plain.Mechanism)
		if !ok {
			return nil, fmt.Errorf("kafka.NewSASL: Invalid credential object for type `plain`")
		}
		return mech, nil
	} else if config.SASLMechanism == SASL_AWS_IAM {

	}
	return nil, fmt.Errorf("kafka.NewSASL: Invalid sasl type `%v`", config.SASLMechanism)
}

func NewAWSIAMSASL() {

}
