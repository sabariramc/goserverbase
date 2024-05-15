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
	"github.com/sabariramc/goserverbase/v6/crypto/aes"
	"github.com/sabariramc/goserverbase/v6/log"
	"github.com/sabariramc/randomstring"
)

const (
	ConstMetadataKMSARN              = "x-kms-arn"
	ConstMetadataEncryptionAlgorithm = "x-encryption-algorithm"
	ConstMetadataContentKey          = "x-content-key"
	ConstEncryptionAlgorithm         = "AES-GCM-256"
)

// S3Crypto extends S3 with client side object encryption, all the objects that are created will be in encrypted format
type S3Crypto struct {
	_ struct{}
	*S3
	kms *KMS
	log log.Log
}

type urlCache struct {
	key         string
	expireTime  time.Time
	contentType string
}

var piiFileCache = make(map[string]*urlCache)

func GetDefaultS3CryptoClient(logger log.Log, keyArn string) *S3Crypto {
	return NewS3CryptoClient(GetDefaultS3Client(logger), GetDefaultKMSClient(logger, keyArn), logger)
}

func NewS3CryptoClient(s3Client *S3, kms *KMS, logger log.Log) *S3Crypto {
	return &S3Crypto{kms: kms, log: logger.NewResourceLogger("S3Crypto"), S3: s3Client}
}

func (s *S3Crypto) encrypt(ctx context.Context, body io.Reader) (io.Reader, map[string]string, error) {
	key := randomstring.Generate(32)
	encryptedKey, err := s.kms.Encrypt(ctx, []byte(key))
	if err != nil {
		s.log.Error(ctx, "error encrypting content key", err)
		return nil, nil, fmt.Errorf("S3Crypto.encrypt: error encrypting content key: %w", err)
	}
	data, err := io.ReadAll(body)
	if err != nil {
		s.log.Error(ctx, "error reading content", err)
		return nil, nil, fmt.Errorf("S3Crypto.encrypt: error reading content: %w", err)
	}
	cipher, err := aes.NewGCM(ctx, s.log, key)
	if err != nil {
		s.log.Error(ctx, "error creating cipher", err)
		return nil, nil, fmt.Errorf("S3Crypto.encrypt: error creating cipher: %w", err)
	}
	data, err = cipher.Encrypt(ctx, data)
	if err != nil {
		s.log.Error(ctx, "error encrypting content", err)
		return nil, nil, fmt.Errorf("S3Crypto.encrypt: error encrypting content: %w", err)
	}
	return bytes.NewReader(data), map[string]string{
		ConstMetadataKMSARN:              *s.kms.keyArn,
		ConstMetadataEncryptionAlgorithm: ConstEncryptionAlgorithm,
		ConstMetadataContentKey:          hex.EncodeToString(encryptedKey),
	}, nil
}

func (s *S3Crypto) PutObject(ctx context.Context, s3Bucket, s3Key string, body io.Reader, mimeType string) error {
	body, metadata, err := s.encrypt(ctx, body)
	if err != nil {
		s.log.Error(ctx, "error encrypting content", err)
		return fmt.Errorf("S3Crypto.PutObject: error encrypting content: %w", err)
	}
	_, err = s.S3.PutObject(ctx, s3Bucket, s3Key, body, mimeType, metadata)
	if err != nil {
		s.log.Error(ctx, "error uploading file", err)
		return fmt.Errorf("S3Crypto.PutObject: error uploading file: %w", err)
	}
	return nil
}

func (s *S3Crypto) PutFile(ctx context.Context, s3Bucket, s3Key, localFilPath string) error {
	fp, err := os.Open(localFilPath)
	if err != nil {
		s.log.Error(ctx, "error opening file", localFilPath)
		return fmt.Errorf("S3Crypto.PutFile: error opening file: %w", err)
	}
	mime, err := mimetype.DetectFile(localFilPath)
	if err != nil {
		s.log.Notice(ctx, "Failed detecting mime type", err)
	}
	s.log.Notice(ctx, "File mimetype", mime)
	defer fp.Close()
	return s.PutObject(ctx, s3Bucket, s3Key, fp, mime.String())
}

func (s *S3Crypto) decrypt(ctx context.Context, res *s3.GetObjectOutput) ([]byte, error) {
	for _, key := range []string{ConstMetadataKMSARN, ConstMetadataContentKey, ConstMetadataEncryptionAlgorithm} {
		if _, ok := res.Metadata[key]; !ok {
			s.log.Error(ctx, "missing metadata", key)
			return nil, fmt.Errorf(fmt.Sprintf("S3Crypto.decrypt: missing metadata %s", key))
		}
	}
	if res.Metadata[ConstMetadataEncryptionAlgorithm] != ConstEncryptionAlgorithm {
		s.log.Error(ctx, "algorithm not supported", res.Metadata[ConstMetadataEncryptionAlgorithm])
		return nil, fmt.Errorf("S3Crypto.decrypt: algorithm not supported: %s", res.Metadata[ConstMetadataEncryptionAlgorithm])
	}
	encryptedKey, err := hex.DecodeString(res.Metadata[ConstMetadataContentKey])
	if err != nil {
		s.log.Error(ctx, "error decoding content key", err)
		return nil, fmt.Errorf("S3Crypto.decrypt: error decoding content key: %w", err)
	}
	decryptKMS := NewKMSClient(s.log, s.kms.Client, res.Metadata[ConstMetadataKMSARN])
	key, err := decryptKMS.Decrypt(ctx, encryptedKey)
	if err != nil {
		s.log.Error(ctx, "error decrypting content key", err)
		return nil, fmt.Errorf("S3Crypto.decrypt: error decrypting content key: %w", err)
	}
	data, err := io.ReadAll(res.Body)
	if err != nil {
		s.log.Error(ctx, "error reading content", err)
		return nil, fmt.Errorf("S3Crypto.decrypt: error reading content: %w", err)
	}
	cipher, err := aes.NewGCM(ctx, s.log, string(key))
	if err != nil {
		s.log.Error(ctx, "error creating cipher", err)
		return nil, fmt.Errorf("S3Crypto.decrypt: error creating cipher: %w", err)
	}
	data, err = cipher.Decrypt(ctx, data)
	if err != nil {
		s.log.Error(ctx, "error decrypting content", err)
		return nil, fmt.Errorf("S3Crypto.decrypt: error decrypting content: %w", err)
	}
	return data, nil
}

func (s *S3Crypto) GetObject(ctx context.Context, s3Bucket, s3Key string) ([]byte, error) {
	res, err := s.S3.GetObject(ctx, s3Bucket, s3Key)
	if err != nil {
		s.log.Error(ctx, "error decrypting content", err)
		return nil, fmt.Errorf("S3Crypto.GetObject: error get object: %w", err)
	}
	return s.decrypt(ctx, res)
}

func (s *S3Crypto) GetFile(ctx context.Context, s3Bucket, s3Key, localFilePath string) error {
	blob, err := s.GetObject(ctx, s3Bucket, s3Key)
	if err != nil {
		return err
	}
	fp, err := os.Create(localFilePath)
	if err != nil {
		s.log.Error(ctx, "error creating local file", err)
		return fmt.Errorf("S3Crypto.GetFile: error creating local file: %w", err)
	}
	defer fp.Close()
	n, err := fp.Write(blob)
	if err != nil {
		s.log.Error(ctx, "error writing to file", err)
		return fmt.Errorf("S3Crypto.GetFile: error writing to file: %w", err)
	}
	if n != len(blob) {
		err := fmt.Errorf("total bytes %v, written bytes %v", len(blob), n)
		s.log.Error(ctx, "S3crypto get file - file writing error", err)
		return fmt.Errorf("S3Crypto.GetFile: %w", err)
	}
	return nil
}

type PIITempFile struct {
	Request     *v4.PresignedHTTPRequest `json:"req"`
	ExpiresAt   time.Time                `json:"expiresAt"`
	ContentType *string                  `json:"contentType"`
}

func (s *S3Crypto) GetFileCache(ctx context.Context, s3Bucket, s3Key, tempPathPart string) (*PIITempFile, error) {
	fullPath := s3Bucket + "/" + s3Key
	fileCache, ok := piiFileCache[fullPath]
	if ok && time.Now().Before(fileCache.expireTime) {
		s.log.Notice(ctx, "File fetched from cache", nil)
	} else {
		blob, err := s.GetObject(ctx, s3Bucket, s3Key)
		if err != nil {
			s.log.Error(ctx, "error downloading file", err)
			return nil, fmt.Errorf("S3Crypto.GetFileCache: error downloading file: %w", err)
		}
		filePath := strings.Split(s3Key, "/")
		tempS3Key := fmt.Sprintf("temp/%v/%v-%v", tempPathPart, uuid.NewString(), filePath[len(filePath)-1])
		mime := mimetype.Detect(blob)
		_, err = s.S3.PutObject(ctx, s3Bucket, tempS3Key, bytes.NewReader(blob), mime.String(), nil)
		if err != nil {
			s.log.Error(ctx, "error uploading temp file", err)
			return nil, fmt.Errorf("S3Crypto.GetFileCache: error uploading temp file: %w", err)
		}
		fileCache = &urlCache{expireTime: time.Now().Add(time.Hour * 20), key: tempS3Key, contentType: mime.String()}
		piiFileCache[fullPath] = fileCache
	}
	url, err := s.PresignGetObject(ctx, s3Bucket, fileCache.key, 30*60)
	if err != nil {
		s.log.Error(ctx, "error pre-signing temp file", err)
		return nil, fmt.Errorf("S3Crypto.GetFileCache: error pre-signing temp file: %w", err)
	}
	return &PIITempFile{Request: url, ContentType: &fileCache.contentType, ExpiresAt: time.Now().Add(time.Minute * 30)}, nil
}
