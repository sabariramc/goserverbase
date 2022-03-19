package aws

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"sabariram.com/goserverbase/log"
)

type KMS struct {
	_      struct{}
	client *kms.KMS
	keyArn *string
	log    *log.Logger
}

var defaultKMSClient *kms.KMS

func GetAWSKMSClient(awsSession *session.Session) *kms.KMS {
	client := kms.New(awsSession)
	return client
}

func GetDefaultKMSClient(logger *log.Logger, keyArn string) *KMS {
	if defaultKMSClient == nil {
		defaultKMSClient = GetAWSKMSClient(defaultAWSSession)
	}
	return GetKMSClient(logger, defaultKMSClient, keyArn)
}

func GetKMSClient(logger *log.Logger, client *kms.KMS, keyArn string) *KMS {
	return &KMS{client: client, keyArn: &keyArn, log: logger}
}

func (k *KMS) Encrypt(ctx context.Context, plainText *string) (cipherTextBlob []byte, b64EncodedText string, err error) {
	req := &kms.EncryptInput{
		KeyId:     k.keyArn,
		Plaintext: []byte(*plainText),
	}
	k.log.Debug(ctx, "KMS encryption request", fmt.Sprint(req))
	res, err := k.client.EncryptWithContext(ctx, req)
	if err != nil {
		k.log.Error(ctx, "KMS encryption error", err)
		return
	}
	k.log.Debug(ctx, "KMS encryption response", res)
	cipherTextBlob = res.CiphertextBlob
	b64EncodedText = base64.StdEncoding.EncodeToString(cipherTextBlob)
	return
}

func (k *KMS) Decrypt(ctx context.Context, b64EncodedText *string) (plainText string, err error) {
	data, err := base64.StdEncoding.DecodeString(*b64EncodedText)
	if err != nil {
		return
	}
	req := &kms.DecryptInput{
		KeyId:          k.keyArn,
		CiphertextBlob: []byte(data),
	}
	k.log.Debug(ctx, "KMS decryption request", req)
	res, err := k.client.DecryptWithContext(ctx, req)
	if err != nil {
		k.log.Error(ctx, "KMS decryption error", err)
		return
	}
	k.log.Debug(ctx, "KMS decryption response", fmt.Sprint(res))
	plainText = string(res.Plaintext)
	return
}
