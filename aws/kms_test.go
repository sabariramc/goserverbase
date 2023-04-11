package aws_test

import (
	"testing"

	"github.com/sabariramc/goserverbase/aws"
)

func TestAWSKMS(t *testing.T) {
	ctx := GetCorrelationContext()
	kms := aws.GetDefaultKMSClient(AWSTestLogger, AWSTestConfig.AWS.KMS_ARN)
	text := "asfasdfsaf"
	_, encryptedText, err := kms.EncryptWithContext(ctx, &text)
	if err != nil {
		t.Fatal(err)
	}
	plainText, err := kms.DecryptWithContext(ctx, &encryptedText)
	if err != nil {
		t.Fatal(err)
	}
	if plainText != text {
		t.Fatalf("Texts are not matching %v, %v", text, plainText)
	}
}
