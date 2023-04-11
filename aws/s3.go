package aws

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gabriel-vasile/mimetype"
	"github.com/sabariramc/goserverbase/v2/log"
)

type S3 struct {
	_ struct{}
	*s3.S3
	log *log.Logger
}

var defaultS3Client *s3.S3

func NewS3ClientWithSession(awsSession *session.Session) *s3.S3 {
	return s3.New(awsSession)
}

func GetDefaultS3Client(logger *log.Logger) *S3 {
	if defaultS3Client == nil {
		defaultS3Client = NewS3ClientWithSession(defaultAWSSession)
	}
	return NewS3Client(defaultS3Client, logger)
}

func NewS3Client(client *s3.S3, logger *log.Logger) *S3 {
	return &S3{S3: client, log: logger}
}

func (s *S3) PutObjectWithContext(ctx context.Context, s3Bucket, s3Key string, body io.ReadSeeker, mimeType string) error {
	req := &s3.PutObjectInput{Bucket: &s3Bucket, Key: &s3Key, Body: body, ContentType: &mimeType}
	s.log.Debug(ctx, "S3 put object request", req)
	res, err := s.S3.PutObjectWithContext(ctx, req)
	if err != nil {
		s.log.Error(ctx, "S3 put object error", err)
		return fmt.Errorf("S3.PutObject: %w", err)
	}
	s.log.Debug(ctx, "S3 put object response", res)
	return nil
}

func (s *S3) PutFile(ctx context.Context, s3Bucket, s3Key, localFilPath string) error {
	fp, err := os.Open(localFilPath)
	if err != nil {
		s.log.Error(ctx, "Error opening file", localFilPath)
		return fmt.Errorf("S3.PutFile: %w", err)
	}
	mime, err := mimetype.DetectFile(localFilPath)
	if err != nil {
		s.log.Notice(ctx, "Failed detecting mime type", err)
	}
	s.log.Debug(ctx, "File mimetype", mime)
	defer fp.Close()
	return s.PutObjectWithContext(ctx, s3Bucket, s3Key, fp, mime.String())
}

func (s *S3) GetObjectWithContext(ctx context.Context, s3Bucket, s3Key string) ([]byte, error) {
	req := &s3.GetObjectInput{Bucket: &s3Bucket, Key: &s3Key}
	s.log.Debug(ctx, "S3 get object request", req)
	res, err := s.S3.GetObjectWithContext(ctx, req)
	if err != nil {
		s.log.Error(ctx, "S3 get object error", err)
		return nil, fmt.Errorf("S3.GetObject: %w", err)
	}
	s.log.Debug(ctx, "S3 get object response", res)
	blob, err := io.ReadAll(res.Body)
	if err != nil {
		s.log.Error(ctx, "S3 get object read error", err)
		return nil, fmt.Errorf("S3.GetObject: %w", err)
	}
	return blob, nil
}

func (s *S3) GetFile(ctx context.Context, s3Bucket, s3Key, localFilePath string) error {
	blob, err := s.GetObjectWithContext(ctx, s3Bucket, s3Key)
	if err != nil {
		return fmt.Errorf("S3.GetFile: %w", err)
	}
	fp, err := os.Create(localFilePath)
	if err != nil {
		s.log.Error(ctx, "S3 get file - file creation error", err)
		return fmt.Errorf("S3.GetFile: %w", err)
	}
	defer fp.Close()
	n, err := fp.Write(blob)
	if err != nil {
		s.log.Error(ctx, "S3 get file - file writing error", err)
		return fmt.Errorf("S3.GetFile: %w", err)
	}
	if n != len(blob) {
		err := fmt.Errorf("total bytes %v, written bytes %v", len(blob), n)
		s.log.Error(ctx, "S3 get file - file writing error", err)
		return fmt.Errorf("S3.GetFile: %w", err)
	}
	return nil
}

func (s *S3) CreatePresignedUrlGET(ctx context.Context, s3Bucket, s3Key string, expireTimeInSeconds int) (*string, error) {
	req, _ := s.GetObjectRequest(&s3.GetObjectInput{
		Bucket: &s3Bucket,
		Key:    &s3Key,
	})
	urlStr, err := req.Presign(time.Duration(expireTimeInSeconds) * time.Second)
	if err != nil {
		s.log.Error(ctx, "S3 failed to sign GET request", err)
		return nil, fmt.Errorf("S3.CreatePresignedUrlGET: %w", err)
	}
	s.log.Debug(ctx, "S3 presigned GET url", urlStr)
	return &urlStr, nil
}

func (s *S3) CreatePresignedUrlPUT(ctx context.Context, s3Bucket, s3Key string, expireTimeInSeconds int) (*string, error) {
	req, _ := s.PutObjectRequest(&s3.PutObjectInput{
		Bucket: &s3Bucket,
		Key:    &s3Key,
	})
	urlStr, err := req.Presign(time.Duration(expireTimeInSeconds) * time.Second)
	if err != nil {
		s.log.Error(ctx, "S3 failed to sign PUT request", err)
		return nil, fmt.Errorf("S3.CreatePresignedUrlPUT: %w", err)
	}
	s.log.Debug(ctx, "S3 presigned PUT url", urlStr)
	return &urlStr, nil
}
