package aes_test

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/sabariramc/goserverbase/v5/crypto/aes"
	"gotest.tools/assert"
)

func TestAESECBPKC5(t *testing.T) {
	key, err := base64.StdEncoding.DecodeString("pVH68zuXerD+SvkGhJFQGw==")
	assert.NilError(t, err)
	cip, err := aes.NewECBPKC5(GetCorrelationContext(), key, ServerTestLogger)
	assert.NilError(t, err)
	data := "fasdfasdasffasdfas fasdfasdfa sfasdf asd fasd fasdf"
	res, _ := cip.Encrypt(context.TODO(), []byte(data))
	decryptedData, _ := cip.Decrypt(context.TODO(), res)
	assert.Equal(t, data, string(decryptedData))
}
