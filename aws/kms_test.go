package aws_test

import (
	"testing"

	"github.com/sabariramc/goserverbase/aws"
)

func TestAWSKMS(t *testing.T) {
	ctx := GetCorrelationContext()
	kms := aws.GetDefaultKMSClient(AWSTestLogger, AWSTestConfig.KMS.Arn)
	text := "asfasdfsaf"
	_, encryptedText, err := kms.Encrypt(ctx, &text)
	if err != nil {
		t.Fatal(err)
	}
	plainText, err := kms.Decrypt(ctx, &encryptedText)
	if err != nil {
		t.Fatal(err)
	}
	if plainText != text {
		t.Fatalf("Texts are not matching %v, %v", text, plainText)
	}
}
