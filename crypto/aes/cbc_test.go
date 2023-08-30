package aes_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/sabariramc/goserverbase/v3/crypto"
	"github.com/sabariramc/goserverbase/v3/crypto/aes"
	"gotest.tools/assert"
)

func TestCBC(t *testing.T) {
	data := "fadsfadfsa"
	ctx := context.TODO()
	var chiper crypto.CipherText
	chiper, err := aes.NewAESCBCPKCS7(ctx, ServerTestLogger, strings.Replace(uuid.New().String(), "-", "", -1))
	assert.NilError(t, err)
	res, err := chiper.EncryptString(ctx, data)
	assert.NilError(t, err)
	deres, err := chiper.DecryptString(ctx, res)
	assert.NilError(t, err)
	assert.Equal(t, data, deres)
}

func TestCBCWithExternal(t *testing.T) {
	data := "DUMMY NAME FOR TEST"
	key := "f52a79201f314543aa731e82e87177e4"
	ctx := context.TODO()
	var chiper crypto.CipherText
	chiper, err := aes.NewAESCBCPKCS7(ctx, ServerTestLogger, key)
	assert.NilError(t, err)
	deres, err := chiper.DecryptString(ctx, "IxxYDxKa5u8Ddy3sE27YCQNZwCBEKc8n7KlSOAU1eGttfYKmp7zeMlTuNaJgCUSO")
	assert.NilError(t, err)
	assert.Equal(t, data, deres)
}

func TestCBCV2(t *testing.T) {
	data := "fadsfadfsa"
	ctx := context.TODO()
	var chiper crypto.CipherText
	key := strings.Replace(uuid.New().String(), "-", "", -1)
	iv := strings.Replace(uuid.New().String(), "-", "", -1)
	chiper, err := aes.NewAESCBCV2PKCS7(ctx, ServerTestLogger, key, []byte(iv)[:16])
	assert.NilError(t, err)
	res, err := chiper.EncryptString(ctx, data)
	assert.NilError(t, err)
	deres, err := chiper.DecryptString(ctx, res)
	assert.NilError(t, err)
	assert.Equal(t, data, deres)
}
