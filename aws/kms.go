package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/sabariramc/goserverbase/v3/log"
)

type KMS struct {
	_ struct{}
	*kms.Client
	keyArn *string
	log    *log.Logger
}

var defaultKMSClient *kms.Client

func NewAWSKMSClientWithConfig(awsConfig aws.Config) *kms.Client {
	client := kms.NewFromConfig(awsConfig)
	return client
}

func GetDefaultKMSClient(logger *log.Logger, keyArn string) *KMS {
	if defaultKMSClient == nil {
		defaultKMSClient = NewAWSKMSClientWithConfig(*defaultAWSConfig)
	}
	return NewKMSClient(logger, defaultKMSClient, keyArn)
}

func NewKMSClient(logger *log.Logger, client *kms.Client, keyArn string) *KMS {
	return &KMS{Client: client, keyArn: &keyArn, log: logger.NewResourceLogger("KMS")}
}

func (k *KMS) Encrypt(ctx context.Context, plainText []byte) (cipherBlob []byte, err error) {
	req := &kms.EncryptInput{
		KeyId:     k.keyArn,
		Plaintext: plainText,
	}
	k.log.Debug(ctx, "KMS encryption request", req)
	res, err := k.Client.Encrypt(ctx, req)
	if err != nil {
		err = fmt.Errorf("KMS.Encrypt: %w", err)
		return
	}
	k.log.Debug(ctx, "KMS encryption response", res)
	cipherBlob = res.CiphertextBlob
	return
}

func (k *KMS) Decrypt(ctx context.Context, cipherBlob []byte) (plainText []byte, err error) {
	req := &kms.DecryptInput{
		KeyId:          k.keyArn,
		CiphertextBlob: cipherBlob,
	}
	k.log.Debug(ctx, "KMS decryption request", req)
	res, err := k.Client.Decrypt(ctx, req)
	if err != nil {
		err = fmt.Errorf("KMS.Decrypt: %w", err)
		return
	}
	k.log.Debug(ctx, "KMS decryption response", res)
	plainText = res.Plaintext
	return
}
