package aws_test

import (
	"testing"

	"github.com/sabariramc/goserverbase/v5/aws"
	"gotest.tools/assert"
)

func TestAWSKMS(t *testing.T) {
	ctx := GetCorrelationContext()
	kms := aws.GetDefaultKMSClient(AWSTestLogger, AWSTestConfig.AWS.KMS_ARN)
	text := "asfasdfsaf"
	encryptedBlob, err := kms.Encrypt(ctx, []byte(text))
	if err != nil {
		assert.NilError(t, err)
	}
	plainText, err := kms.Decrypt(ctx, encryptedBlob)
	if err != nil {
		assert.NilError(t, err)
	}
	if string(plainText) != text {
		t.Fatalf("Texts are not matching %v, %v", text, plainText)
	}
}
