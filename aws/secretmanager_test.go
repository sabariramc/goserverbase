package aws_test

import (
	"testing"

	"github.com/sabariramc/goserverbase/aws"
)

func TestSecretManager(t *testing.T) {
	ctx := GetCorrelationContext()
	client := aws.GetDefaultSecretManagerClient(AWSTestLogger)
	_, err := client.GetSecret(ctx, AWSTestConfig.SecretManager.Arn)
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.GetSecret(ctx, AWSTestConfig.SecretManager.Arn)
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.GetSecretNonCache(ctx, AWSTestConfig.SecretManager.Arn)
	if err != nil {
		t.Fatal(err)
	}
}
