package aws_test

import (
	"fmt"
	"testing"

	"github.com/sabariramc/goserverbase/v4/aws"
)

func TestSecretManager(t *testing.T) {
	ctx := GetCorrelationContext()
	client := aws.GetDefaultSecretManagerClient(AWSTestLogger)
	secret, err := client.GetSecret(ctx, AWSTestConfig.AWS.SECRET_ARN)
	fmt.Println(secret)
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.GetSecret(ctx, AWSTestConfig.AWS.SECRET_ARN)
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.GetSecret(ctx, AWSTestConfig.AWS.SECRET_ARN)
	if err != nil {
		t.Fatal(err)
	}
}
