package aws_test

import (
	"fmt"
	"testing"

	"github.com/sabariramc/goserverbase/v6/aws"
	"gotest.tools/assert"
)

func TestSecretManager(t *testing.T) {
	ctx := GetCorrelationContext()
	client := aws.GetDefaultSecretManagerClient(AWSTestLogger)
	secret, err := client.GetSecretString(ctx, AWSTestConfig.AWS.SECRET)
	fmt.Println(secret)
	assert.NilError(t, err)
	_, err = client.GetSecretString(ctx, AWSTestConfig.AWS.SECRET)
	assert.NilError(t, err)
	_, err = client.GetSecretString(ctx, AWSTestConfig.AWS.SECRET)
	assert.NilError(t, err)
}
