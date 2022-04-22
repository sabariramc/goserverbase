package tests

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/sabariramc/goserverbase/aws"
)

func TestS3(t *testing.T) {
	ctx := GetCorrelationContext()
	s3Client := aws.GetDefaultS3Client(AWSTestLogger)
	path := fmt.Sprintf("dev/temp/goserverbasetest/%v.pdf", uuid.NewString())
	s3Bucket := AWSTestConfig.S3.BucketName
	err := s3Client.PutFile(ctx, s3Bucket, path, "../testfile/sample_aadhaar.pdf")
	if err != nil {
		t.Fatal(err)
	}
	err = s3Client.GetFile(ctx, s3Bucket, path, "../testfile/test.pdf")
	if err != nil {
		t.Fatal(err)
	}
	_, err = s3Client.CreatePresignedURLGET(ctx, s3Bucket, path, 10*60)
	if err != nil {
		t.Fatal(err)
	}
	path = fmt.Sprintf("dev/temp/goserverbasetest/%v.pdf", uuid.NewString())
	_, err = s3Client.CreatePresignedURLPUT(ctx, s3Bucket, path, 10)
	if err != nil {
		t.Fatal(err)
	}
}

func TestS3PII(t *testing.T) {
	ctx := GetCorrelationContext()
	keyArn := AWSTestConfig.KMS.Arn
	s3Client, err := aws.GetDefaultS3PIIClient(AWSTestLogger, keyArn)
	if err != nil {
		t.Fatal(err)
	}
	path := fmt.Sprintf("dev/temp/goserverbasetest/%v.pdf", uuid.NewString())
	s3Bucker := AWSTestConfig.S3.BucketName
	err = s3Client.PutFile(ctx, s3Bucker, path, "../testfile/sample_aadhaar.pdf")
	if err != nil {
		t.Fatal(err)
	}
	err = s3Client.GetFile(ctx, s3Bucker, path, "../testfile/testpii.pdf")
	if err != nil {
		t.Fatal(err)
	}
	_, err = s3Client.GetFileCache(ctx, s3Bucker, path, "dev", "testCache")
	if err != nil {
		t.Fatal(err)
	}
}
