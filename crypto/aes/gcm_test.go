package aes_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/sabariramc/goserverbase/v6/crypto"
	"github.com/sabariramc/goserverbase/v6/crypto/aes"
	"github.com/sabariramc/randomstring"
	"gotest.tools/assert"
)

func TestAESGCM(t *testing.T) {
	ctx := GetCorrelationContext()
	key := randomstring.Generate(32)
	var chiper crypto.CipherText
	chiper, err := aes.NewGCM(ctx, ServerTestLogger, key)
	assert.NilError(t, err)
	data := "fasdfasdasffasdfas fasdfasdfa sfasdf asd fasd fasdf"
	res, _ := chiper.Encrypt(context.TODO(), []byte(data))
	decryptedData, _ := chiper.Decrypt(context.TODO(), res)
	assert.Equal(t, data, string(decryptedData))
}

func BenchmarkAESGCM(b *testing.B) {
	ctx := GetCorrelationContext()
	key := randomstring.Generate(32)
	var chiper crypto.CipherText
	chiper, err := aes.NewGCM(ctx, ServerTestLogger, key)
	assert.NilError(b, err)
	for i := 0; i < b.N; i++ {
		data := uuid.NewString() + uuid.NewString()
		res, _ := chiper.Encrypt(context.TODO(), []byte(data))
		decryptedData, _ := chiper.Decrypt(context.TODO(), res)
		assert.Equal(b, data, string(decryptedData))
	}

}
