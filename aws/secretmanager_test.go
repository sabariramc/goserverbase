package aws_test

import (
	"fmt"
	"testing"

	"github.com/sabariramc/goserverbase/v5/aws"
	"gotest.tools/assert"
)

func TestSecretManager(t *testing.T) {
	ctx := GetCorrelationContext()
	client := aws.GetDefaultSecretManagerClient(AWSTestLogger)
	secret, err := client.GetSecret(ctx, AWSTestConfig.AWS.SECRET_ARN)
	fmt.Println(secret)
	assert.NilError(t, err)
	_, err = client.GetSecret(ctx, AWSTestConfig.AWS.SECRET_ARN)
	assert.NilError(t, err)
	_, err = client.GetSecret(ctx, AWSTestConfig.AWS.SECRET_ARN)
	assert.NilError(t, err)
}
