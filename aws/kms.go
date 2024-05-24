package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/sabariramc/goserverbase/v6/log"
)

// KMS provides an interface to interact with AWS Key Management Service (KMS).
type KMS struct {
	_ struct{}
	*kms.Client
	keyArn *string
	log    log.Log
}

var defaultKMSClient *kms.Client

// NewKMSClientWithConfig creates a new KMS client with the provided AWS configuration.
func NewKMSClientWithConfig(awsConfig aws.Config) *kms.Client {
	client := kms.NewFromConfig(awsConfig)
	return client
}

// GetDefaultKMSClient retrieves the default KMS client using the provided logger and key ARN.
func GetDefaultKMSClient(logger log.Log, keyArn string) *KMS {
	if defaultKMSClient == nil {
		defaultKMSClient = NewKMSClientWithConfig(*defaultAWSConfig)
	}
	return NewKMSClient(logger, defaultKMSClient, keyArn)
}

// NewKMSClient creates a new KMS client with the provided logger, client, and key ARN.
func NewKMSClient(logger log.Log, client *kms.Client, keyArn string) *KMS {
	return &KMS{Client: client, keyArn: &keyArn, log: logger.NewResourceLogger("KMS")}
}

// Encrypt encrypts the plaintext using the KMS client and returns the ciphertext.
func (k *KMS) Encrypt(ctx context.Context, plainText []byte) (cipherBlob []byte, err error) {
	req := &kms.EncryptInput{
		KeyId:     k.keyArn,
		Plaintext: plainText,
	}
	res, err := k.Client.Encrypt(ctx, req)
	if err != nil {
		k.log.Error(ctx, "error encrypting content", err)
		err = fmt.Errorf("KMS.Encrypt: error encrypting content: %w", err)
		return
	}
	cipherBlob = res.CiphertextBlob
	return
}

// Decrypt decrypts the ciphertext using the KMS client and returns the plaintext.
func (k *KMS) Decrypt(ctx context.Context, cipherBlob []byte) (plainText []byte, err error) {
	req := &kms.DecryptInput{
		KeyId:          k.keyArn,
		CiphertextBlob: cipherBlob,
	}
	res, err := k.Client.Decrypt(ctx, req)
	if err != nil {
		k.log.Error(ctx, "error decrypting content", err)
		err = fmt.Errorf("KMS.Decrypt: error decrypting content: %w", err)
		return
	}
	plainText = res.Plaintext
	return
}
