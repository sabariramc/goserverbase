package aws

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/sabariramc/goserverbase/log"
)

type KMS struct {
	_ struct{}
	*kms.KMS
	keyArn *string
	log    *log.Logger
}

var defaultKMSClient *kms.KMS

func NewKMSClientWithSession(awsSession *session.Session) *kms.KMS {
	client := kms.New(awsSession)
	return client
}

func GetDefaultKMSClient(logger *log.Logger, keyArn string) *KMS {
	if defaultKMSClient == nil {
		defaultKMSClient = NewKMSClientWithSession(defaultAWSSession)
	}
	return NewKMSClient(logger, defaultKMSClient, keyArn)
}

func NewKMSClient(logger *log.Logger, client *kms.KMS, keyArn string) *KMS {
	return &KMS{KMS: client, keyArn: &keyArn, log: logger}
}

func (k *KMS) EncryptWithContext(ctx context.Context, plainText *string) (cipherTextBlob []byte, b64EncodedText string, err error) {
	req := &kms.EncryptInput{
		KeyId:     k.keyArn,
		Plaintext: []byte(*plainText),
	}
	k.log.Debug(ctx, "KMS encryption request", req)
	res, err := k.KMS.EncryptWithContext(ctx, req)
	if err != nil {
		k.log.Error(ctx, "KMS encryption error", err)
		err = fmt.Errorf("KMS.Encrypt: %w", err)
		return
	}
	k.log.Debug(ctx, "KMS encryption response", res)
	cipherTextBlob = res.CiphertextBlob
	b64EncodedText = base64.StdEncoding.EncodeToString(cipherTextBlob)
	return
}

func (k *KMS) DecryptWithContext(ctx context.Context, b64EncodedText *string) (plainText string, err error) {
	data, err := base64.StdEncoding.DecodeString(*b64EncodedText)
	if err != nil {
		return
	}
	req := &kms.DecryptInput{
		KeyId:          k.keyArn,
		CiphertextBlob: []byte(data),
	}
	k.log.Debug(ctx, "KMS decryption request", req)
	res, err := k.KMS.DecryptWithContext(ctx, req)
	if err != nil {
		k.log.Error(ctx, "KMS decryption error", err)
		err = fmt.Errorf("KMS.Decrypt: %w", err)
		return
	}
	k.log.Debug(ctx, "KMS decryption response", res)
	plainText = string(res.Plaintext)
	return
}
