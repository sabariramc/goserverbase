package aws

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3crypto"
	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"sabariram.com/goserverbase/log"
)

type S3PII struct {
	_                struct{}
	encryptionClient *s3crypto.EncryptionClientV2
	decryptionClient *s3crypto.DecryptionClient
	log              *log.Logger
	s3Client         *S3
}

type urlCache struct {
	key         string
	expireTime  time.Time
	contentType string
}

var piiFileCache = make(map[string]*urlCache)

var defaultS3EncryptionClient *s3crypto.EncryptionClientV2
var defaultS3DecryptionClient *s3crypto.DecryptionClient

func GetAWSS3EncryptionClient(awsSession *session.Session, keyArn string) (*s3crypto.EncryptionClientV2, error) {
	var matdesc s3crypto.MaterialDescription
	keywrap := s3crypto.NewKMSContextKeyGenerator(kms.New(awsSession), keyArn, matdesc)
	builder := s3crypto.AESGCMContentCipherBuilderV2(keywrap)
	return s3crypto.NewEncryptionClientV2(awsSession, builder)
}

func GetAWSS3DecryptionClient(awsSession *session.Session) *s3crypto.DecryptionClient {
	return s3crypto.NewDecryptionClient(awsSession)
}

func GetDefaultS3PIIClient(logger *log.Logger, keyArn string) (*S3PII, error) {
	if defaultS3EncryptionClient == nil {
		client, err := GetAWSS3EncryptionClient(defaultAWSSession, keyArn)
		if err != nil {
			return nil, err
		}
		defaultS3EncryptionClient = client
	}
	if defaultS3DecryptionClient == nil {
		defaultS3DecryptionClient = GetAWSS3DecryptionClient(defaultAWSSession)
	}
	return GetS3PIIClient(defaultS3EncryptionClient, defaultS3DecryptionClient, GetDefaultS3Client(logger), logger), nil
}

func GetS3PIIClient(encryptionClient *s3crypto.EncryptionClientV2, decryptionClient *s3crypto.DecryptionClient, s3Client *S3, logger *log.Logger) *S3PII {
	return &S3PII{encryptionClient: encryptionClient, decryptionClient: decryptionClient, log: logger, s3Client: s3Client}
}

func (s *S3PII) PutObject(ctx context.Context, s3Bucket, s3Key string, body io.ReadSeeker, mimeType string) error {
	req := &s3.PutObjectInput{Bucket: &s3Bucket, Key: &s3Key, Body: body, ContentType: &mimeType}
	s.log.Debug(ctx, "S3crypto put object request", req)
	res, err := s.encryptionClient.PutObjectWithContext(ctx, req)
	if err != nil {
		s.log.Error(ctx, "S3crypto put object error", err)
		return err
	}
	s.log.Debug(ctx, "S3crypto put object response", res)
	return nil
}

func (s *S3PII) PutFile(ctx context.Context, s3Bucket, s3Key, localFilPath string) error {
	fp, err := os.Open(localFilPath)
	if err != nil {
		s.log.Error(ctx, "Error opening file", localFilPath)
		return err
	}
	mime, err := mimetype.DetectFile(localFilPath)
	if err != nil {
		s.log.Error(ctx, "Error detecting mime type", err)
	}
	s.log.Debug(ctx, "File mimetype", mime)
	defer fp.Close()
	return s.PutObject(ctx, s3Bucket, s3Key, fp, mime.String())
}

func (s *S3PII) GetObject(ctx context.Context, s3Bucket, s3Key string) ([]byte, error) {
	req := &s3.GetObjectInput{Bucket: &s3Bucket, Key: &s3Key}
	s.log.Debug(ctx, "S3crypto get object request", req)
	res, err := s.decryptionClient.GetObjectWithContext(ctx, req)
	if err != nil {
		s.log.Error(ctx, "S3crypto get object error", err)
		return nil, err
	}
	s.log.Debug(ctx, "S3crypto get object response", res)
	blob, err := io.ReadAll(res.Body)
	if err != nil {
		s.log.Error(ctx, "S3crypto get object read error", err)
		return nil, err
	}
	return blob, nil
}

func (s *S3PII) GetFile(ctx context.Context, s3Bucket, s3Key, localFilePath string) error {
	blob, err := s.GetObject(ctx, s3Bucket, s3Key)
	if err != nil {
		return err
	}
	fp, err := os.Create(localFilePath)
	if err != nil {
		s.log.Error(ctx, "S3crypto get file - file creation error", err)
		return err
	}
	defer fp.Close()
	n, err := fp.Write(blob)
	if err != nil {
		s.log.Error(ctx, "S3crypto get file - file writing error", err)
		return err
	}
	if n != len(blob) {
		err := fmt.Errorf("total bytes %v, written bytes %v", len(blob), n)
		s.log.Error(ctx, "S3crypto get file - file writing error", err)
		return err
	}
	return nil
}

type PIITempFile struct {
	URL         *string   `json:"url"`
	ExpiresAt   time.Time `json:"expiresAt"`
	ContentType *string   `json:"contentType"`
}

func (s *S3PII) GetFileCache(ctx context.Context, s3Bucket, s3Key, stage, tempPathPart string) (*PIITempFile, error) {
	fullPath := s3Bucket + "/" + s3Key
	fileCache, ok := piiFileCache[fullPath]
	if ok && time.Now().Before(fileCache.expireTime) {
		s.log.Info(ctx, "File fetched from cache", nil)
	} else {
		blob, err := s.GetObject(ctx, s3Bucket, s3Key)
		if err != nil {
			return nil, err
		}
		filePath := strings.Split(s3Key, "/")
		tempS3Key := fmt.Sprintf("/%v/temp/%v/%v-%v", stage, tempPathPart, uuid.NewString(), filePath[len(filePath)-1])
		mime := mimetype.Detect(blob)
		err = s.s3Client.PutObject(ctx, s3Bucket, tempS3Key, bytes.NewReader(blob), mime.String())
		if err != nil {
			return nil, err
		}
		fileCache = &urlCache{expireTime: time.Now().Add(time.Hour * 20), key: tempS3Key, contentType: mime.String()}
		piiFileCache[fullPath] = fileCache
	}
	url, err := s.s3Client.CreatePresignedURLGET(ctx, s3Bucket, fileCache.key, 30*60)
	if err != nil {
		return nil, err
	}
	return &PIITempFile{URL: url, ContentType: &fileCache.contentType, ExpiresAt: time.Now().Add(time.Minute * 30)}, nil
}
