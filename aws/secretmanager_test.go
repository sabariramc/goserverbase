package aws_test

import (
	"fmt"
	"testing"

	"github.com/sabariramc/goserverbase/v4/aws"
	"gotest.tools/assert"
)

func TestSecretManager(t *testing.T) {
	ctx := GetCorrelationContext()
	client := aws.GetDefaultSecretManagerClient(AWSTestLogger)
	secret, err := client.GetSecret(ctx, AWSTestConfig.AWS.SECRET_ARN)
	fmt.Println(secret)
	if err != nil {
		assert.NilError(t, err)
	}
	_, err = client.GetSecret(ctx, AWSTestConfig.AWS.SECRET_ARN)
	if err != nil {
		assert.NilError(t, err)
	}
	_, err = client.GetSecret(ctx, AWSTestConfig.AWS.SECRET_ARN)
	if err != nil {
		assert.NilError(t, err)
	}
}
