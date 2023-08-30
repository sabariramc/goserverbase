package aws

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"github.com/sabariramc/goserverbase/v3/crypto/aes"
	"github.com/sabariramc/goserverbase/v3/log"
	"github.com/sabariramc/goserverbase/v3/utils"
)

const (
	ConstMetadataKMSARN              = "x-kms-arn"
	ConstMetadataEncryptionAlgorithm = "x-encryption-algorithm"
	ConstMetadataContentKey          = "x-content-key"
	ConstEncryptionAlgorithm         = "AES-GCM-256"
)

type S3PII struct {
	_ struct{}
	*S3
	kms *KMS
	log *log.Logger
}

type urlCache struct {
	key         string
	expireTime  time.Time
	contentType string
}

var piiFileCache = make(map[string]*urlCache)

func GetDefaultS3PIIClient(logger *log.Logger, keyArn string) *S3PII {
	return NewS3PIIClient(GetDefaultS3Client(logger), GetDefaultKMSClient(logger, keyArn), logger)
}

func NewS3PIIClient(s3Client *S3, kms *KMS, logger *log.Logger) *S3PII {
	return &S3PII{kms: kms, log: logger, S3: s3Client}
}

func (s *S3PII) encrypt(ctx context.Context, body io.Reader) (io.Reader, map[string]string, error) {
	key := utils.GenerateRandomString(32)
	encryptedKey, err := s.kms.Encrypt(ctx, []byte(key))
	if err != nil {
		return nil, nil, fmt.Errorf("S3PII.encrypt:error on encrypting content key:%w", err)
	}
	data, err := io.ReadAll(body)
	if err != nil {
		return nil, nil, fmt.Errorf("S3PII.encrypt:error on reading content:%w", err)
	}
	cipher, err := aes.NewAESGCM(ctx, s.log, key)
	if err != nil {
		return nil, nil, fmt.Errorf("S3PII.encrypt:error on creating cipher:%w", err)
	}
	data, err = cipher.Encrypt(ctx, data)
	if err != nil {
		return nil, nil, fmt.Errorf("S3PII.encrypt:error on encrypting content:%w", err)
	}
	return bytes.NewReader(data), map[string]string{
		ConstMetadataKMSARN:              *s.kms.keyArn,
		ConstMetadataEncryptionAlgorithm: ConstEncryptionAlgorithm,
		ConstMetadataContentKey:          hex.EncodeToString(encryptedKey),
	}, nil
}

func (s *S3PII) PutObject(ctx context.Context, s3Bucket, s3Key string, body io.Reader, mimeType string) error {
	body, metadata, err := s.encrypt(ctx, body)
	if err != nil {
		return fmt.Errorf("S3PII.PutObject: %w", err)
	}
	_, err = s.S3.PutObject(ctx, s3Bucket, s3Key, body, mimeType, metadata)
	if err != nil {
		return fmt.Errorf("S3PII.PutObject: error on uploading file: %w", err)
	}
	return nil
}

func (s *S3PII) PutFile(ctx context.Context, s3Bucket, s3Key, localFilPath string) error {
	fp, err := os.Open(localFilPath)
	if err != nil {
		s.log.Error(ctx, "Error opening file", localFilPath)
		return fmt.Errorf("S3PII.PutFile: %w", err)
	}
	mime, err := mimetype.DetectFile(localFilPath)
	if err != nil {
		s.log.Notice(ctx, "Failed detecting mime type", err)
	}
	s.log.Debug(ctx, "File mimetype", mime)
	defer fp.Close()
	return s.PutObject(ctx, s3Bucket, s3Key, fp, mime.String())
}

func (s *S3PII) decrypt(ctx context.Context, res *s3.GetObjectOutput) ([]byte, error) {
	for _, key := range []string{ConstMetadataKMSARN, ConstMetadataContentKey, ConstMetadataEncryptionAlgorithm} {
		if _, ok := res.Metadata[key]; !ok {
			return nil, fmt.Errorf(fmt.Sprintf("S3PII.decrypt: missing metadata %s", key))
		}
	}
	if res.Metadata[ConstMetadataEncryptionAlgorithm] != ConstEncryptionAlgorithm {
		return nil, fmt.Errorf("S3PII.decrypt: algorithm not supported :%s", res.Metadata[ConstMetadataEncryptionAlgorithm])
	}
	encryptedKey, err := hex.DecodeString(res.Metadata[ConstMetadataContentKey])
	if err != nil {
		return nil, fmt.Errorf("S3PII.decrypt:error on decoding content key:%w", err)
	}
	decryptKMS := NewKMSClient(s.log, s.kms.Client, res.Metadata[ConstMetadataKMSARN])
	key, err := decryptKMS.Decrypt(ctx, encryptedKey)
	if err != nil {
		return nil, fmt.Errorf("S3PII.decrypt:error on decrypting content key:%w", err)
	}
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("S3PII.decrypt:error on reading content:%w", err)
	}
	cipher, err := aes.NewAESGCM(ctx, s.log, string(key))
	if err != nil {
		return nil, fmt.Errorf("S3PII.decrypt:error on creating cipher:%w", err)
	}
	data, err = cipher.Decrypt(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("S3PII.decrypt:error on decrypting content:%w", err)
	}
	return data, nil
}

func (s *S3PII) GetObject(ctx context.Context, s3Bucket, s3Key string) ([]byte, error) {
	res, err := s.S3.GetObject(ctx, s3Bucket, s3Key)
	if err != nil {
		return nil, fmt.Errorf("S3PII.GetObject: %w", err)
	}
	return s.decrypt(ctx, res)
}

func (s *S3PII) GetFile(ctx context.Context, s3Bucket, s3Key, localFilePath string) error {
	blob, err := s.GetObject(ctx, s3Bucket, s3Key)
	if err != nil {
		return err
	}
	fp, err := os.Create(localFilePath)
	if err != nil {
		s.log.Error(ctx, "S3crypto get file - file creation error", err)
		return fmt.Errorf("S3PII.GetFile: %w", err)
	}
	defer fp.Close()
	n, err := fp.Write(blob)
	if err != nil {
		s.log.Error(ctx, "S3crypto get file - file writing error", err)
		return fmt.Errorf("S3PII.GetFile: %w", err)
	}
	if n != len(blob) {
		err := fmt.Errorf("total bytes %v, written bytes %v", len(blob), n)
		s.log.Error(ctx, "S3crypto get file - file writing error", err)
		return fmt.Errorf("S3PII.GetFile: %w", err)
	}
	return nil
}

type PIITempFile struct {
	Request     *v4.PresignedHTTPRequest `json:"req"`
	ExpiresAt   time.Time                `json:"expiresAt"`
	ContentType *string                  `json:"contentType"`
}

func (s *S3PII) GetFileCache(ctx context.Context, s3Bucket, s3Key, tempPathPart string) (*PIITempFile, error) {
	fullPath := s3Bucket + "/" + s3Key
	fileCache, ok := piiFileCache[fullPath]
	if ok && time.Now().Before(fileCache.expireTime) {
		s.log.Info(ctx, "File fetched from cache", nil)
	} else {
		blob, err := s.GetObject(ctx, s3Bucket, s3Key)
		if err != nil {
			return nil, fmt.Errorf("S3PII.GetFileCache: %w", err)
		}
		filePath := strings.Split(s3Key, "/")
		tempS3Key := fmt.Sprintf("/temp/%v/%v-%v", tempPathPart, uuid.NewString(), filePath[len(filePath)-1])
		mime := mimetype.Detect(blob)
		_, err = s.S3.PutObject(ctx, s3Bucket, tempS3Key, bytes.NewReader(blob), mime.String(), nil)
		if err != nil {
			return nil, fmt.Errorf("S3PII.GetFileCache: %w", err)
		}
		fileCache = &urlCache{expireTime: time.Now().Add(time.Hour * 20), key: tempS3Key, contentType: mime.String()}
		piiFileCache[fullPath] = fileCache
	}
	url, err := s.PresignGetObject(ctx, s3Bucket, fileCache.key, 30*60)
	if err != nil {
		return nil, fmt.Errorf("S3PII.GetFileCache: %w", err)
	}
	return &PIITempFile{Request: url, ContentType: &fileCache.contentType, ExpiresAt: time.Now().Add(time.Minute * 30)}, nil
}
