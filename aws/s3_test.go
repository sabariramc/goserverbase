package aws_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/sabariramc/goserverbase/v6/aws"
	"gotest.tools/assert"
)

func TestS3(t *testing.T) {
	ctx := GetCorrelationContext()
	s3Client := aws.NewS3Client(s3.NewFromConfig(*aws.GetDefaultAWSConfig(), func(o *s3.Options) {
		o.UsePathStyle = true
	}), AWSTestLogger)
	path := fmt.Sprintf("dev/goserverbasetest/plain/%v.pdf", uuid.NewString())
	s3Bucket := AWSTestConfig.AWS.S3
	_, err := s3Client.PutFile(ctx, s3Bucket, path, "./testdata/sample.pdf")
	assert.NilError(t, err)
	err = s3Client.GetFile(ctx, s3Bucket, path, "./testdata/result/test.pdf")
	assert.NilError(t, err)
	_, err = s3Client.PresignGetObject(ctx, s3Bucket, path, 10*60)
	assert.NilError(t, err)
	path = fmt.Sprintf("dev/temp/goserverbasetest/%v.pdf", uuid.NewString())
	_, err = s3Client.PresignPutObject(ctx, s3Bucket, path, 10)
	assert.NilError(t, err)
}

func TestS3PII(t *testing.T) {
	ctx := GetCorrelationContext()
	keyArn := AWSTestConfig.AWS.KMS
	s3Client := aws.NewS3CryptoClient(aws.NewS3Client(s3.NewFromConfig(*aws.GetDefaultAWSConfig(), func(o *s3.Options) {
		o.UsePathStyle = true
	}), AWSTestLogger), aws.GetDefaultKMSClient(AWSTestLogger, keyArn), AWSTestLogger)
	path := fmt.Sprintf("dev/goserverbasetest/pii/%v.pdf", uuid.NewString())
	s3Bucker := AWSTestConfig.AWS.S3
	err := s3Client.PutFile(ctx, s3Bucker, path, "./testdata/sample.pdf")
	assert.NilError(t, err)
	err = s3Client.GetFile(ctx, s3Bucker, path, "./testdata/result/testpii.pdf")
	assert.NilError(t, err)
	_, err = s3Client.GetFileCache(ctx, s3Bucker, path, "testCache")
	assert.NilError(t, err)
}
