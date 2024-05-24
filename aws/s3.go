package aws

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gabriel-vasile/mimetype"
	"github.com/sabariramc/goserverbase/v6/log"
)

// S3 provides an interface for interacting with AWS Simple Storage Service (S3).
type S3 struct {
	_ struct{}
	*s3.Client
	*s3.PresignClient
	log log.Log
}

var defaultS3Client *s3.Client

// NewS3ClientWithConfig creates a new S3 client with the provided AWS configuration.
func NewS3ClientWithConfig(awsConfig aws.Config) *s3.Client {
	return s3.NewFromConfig(awsConfig)
}

// GetDefaultS3Client retrieves the default S3 client using the provided logger.
func GetDefaultS3Client(logger log.Log) *S3 {
	if defaultS3Client == nil {
		defaultS3Client = NewS3ClientWithConfig(*defaultAWSConfig)
	}
	return NewS3Client(defaultS3Client, logger)
}

// NewS3Client creates a new S3 client with the provided client and logger.
func NewS3Client(client *s3.Client, logger log.Log) *S3 {
	return &S3{Client: client, log: logger.NewResourceLogger("S3"), PresignClient: s3.NewPresignClient(client)}
}

// PutObject uploads an object to S3 and returns the response.
func (s *S3) PutObject(ctx context.Context, s3Bucket, s3Key string, body io.Reader, mimeType string, metadata map[string]string) (*s3.PutObjectOutput, error) {
	req := &s3.PutObjectInput{Bucket: &s3Bucket, Key: &s3Key, Body: body, ContentType: &mimeType, Metadata: metadata}
	res, err := s.Client.PutObject(ctx, req)
	if err != nil {
		s.log.Error(ctx, "error uploading file", err)
		return nil, fmt.Errorf("S3.PutObject: error uploading file: %w", err)
	}
	return res, nil
}

// PutFile uploads a file to S3 using the provided local file path and returns the response.
func (s *S3) PutFile(ctx context.Context, s3Bucket, s3Key, localFilePath string) (*s3.PutObjectOutput, error) {
	fp, err := os.Open(localFilePath)
	if err != nil {
		s.log.Error(ctx, "Error opening file", localFilePath)
		return nil, fmt.Errorf("S3.PutFile: error opening file: %w", err)
	}
	defer fp.Close()
	mime, err := mimetype.DetectFile(localFilePath)
	if err != nil {
		s.log.Notice(ctx, "Failed detecting mime type", err)
	}
	s.log.Notice(ctx, "File mimetype", mime.String())
	return s.PutObject(ctx, s3Bucket, s3Key, fp, mime.String(), nil)
}

// GetObject retrieves an object from S3 and returns the response.
func (s *S3) GetObject(ctx context.Context, s3Bucket, s3Key string) (*s3.GetObjectOutput, error) {
	req := &s3.GetObjectInput{Bucket: &s3Bucket, Key: &s3Key}
	res, err := s.Client.GetObject(ctx, req)
	if err != nil {
		s.log.Error(ctx, "error downloading file", err)
		return nil, fmt.Errorf("S3.GetObject: error downloading file: %w", err)
	}
	return res, nil
}

// GetFile downloads an object from S3 to the specified local file path.
func (s *S3) GetFile(ctx context.Context, s3Bucket, s3Key, localFilePath string) error {
	res, err := s.GetObject(ctx, s3Bucket, s3Key)
	if err != nil {
		return fmt.Errorf("S3.GetFile: %w", err)
	}
	blob, err := io.ReadAll(res.Body)
	if err != nil {
		s.log.Error(ctx, "error reading remote content", err)
		return fmt.Errorf("S3.GetFile: error reading remote content: %w", err)
	}
	fp, err := os.Create(localFilePath)
	if err != nil {
		s.log.Error(ctx, "error creating local file", err)
		return fmt.Errorf("S3.GetFile: error creating local file: %w", err)
	}
	defer fp.Close()
	n, err := fp.Write(blob)
	if err != nil {
		s.log.Error(ctx, "error writing to local file", err)
		return fmt.Errorf("S3.GetFile: error writing to local file: %w", err)
	}
	if n != len(blob) {
		s.log.Error(ctx, fmt.Sprintf("total bytes %v, written bytes %v", len(blob), n), nil)
		return fmt.Errorf("S3.GetFile: incomplete local write")
	}
	return nil
}

// PresignGetObject creates a presigned URL for downloading an object from S3.
func (s *S3) PresignGetObject(ctx context.Context, s3Bucket, s3Key string, expireTimeInSeconds int64) (*v4.PresignedHTTPRequest, error) {
	request, err := s.PresignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(s3Key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(expireTimeInSeconds * int64(time.Second))
	})
	if err != nil {
		s.log.Error(ctx, "error creating presigned get request", err)
		return nil, fmt.Errorf("S3.CreatePresignedUrlGET: error creating presigned get request: %w", err)
	}
	return request, nil
}

// PresignPutObject creates a presigned URL for uploading an object to S3.
func (s *S3) PresignPutObject(ctx context.Context, s3Bucket, s3Key string, expireTimeInSeconds int64) (*v4.PresignedHTTPRequest, error) {
	request, err := s.PresignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(s3Key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(expireTimeInSeconds * int64(time.Second))
	})
	if err != nil {
		s.log.Error(ctx, "error creating presigned put request", err)
		return nil, fmt.Errorf("S3.CreatePresignedUrlPUT: error creating presigned put request: %w", err)
	}
	return request, nil
}
