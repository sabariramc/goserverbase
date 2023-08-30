package aws_test

import (
	"testing"

	"github.com/sabariramc/goserverbase/v3/aws"
)

func TestAWSKMS(t *testing.T) {
	ctx := GetCorrelationContext()
	kms := aws.GetDefaultKMSClient(AWSTestLogger, AWSTestConfig.AWS.KMS_ARN)
	text := "asfasdfsaf"
	encryptedBlob, err := kms.Encrypt(ctx, []byte(text))
	if err != nil {
		t.Fatal(err)
	}
	plainText, err := kms.Decrypt(ctx, encryptedBlob)
	if err != nil {
		t.Fatal(err)
	}
	if string(plainText) != text {
		t.Fatalf("Texts are not matching %v, %v", text, plainText)
	}
}
