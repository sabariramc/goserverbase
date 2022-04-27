package tests

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/sabariramc/goserverbase/crypto/aes"
	"gotest.tools/assert"
)

func TestCBC(t *testing.T) {
	data := "fadsfadfsa"
	chiper, err := aes.NewAESCBCPKCS7(context.TODO(), ServerTestLogger, strings.Replace(uuid.New().String(), "-", "", -1))
	assert.NilError(t, err)
	res, err := chiper.EncryptString(data)
	assert.NilError(t, err)
	deres, err := chiper.DecryptString(res)
	assert.NilError(t, err)
	assert.Equal(t, data, deres)
}

func TestCBCWithPython(t *testing.T) {
	data := "DUMMY NAME FOR TEST"
	key := "f52a79201f314543aa731e82e87177e4"
	chiper, err := aes.NewAESCBCPKCS7(context.TODO(), ServerTestLogger, key)
	assert.NilError(t, err)
	deres, err := chiper.DecryptString("IxxYDxKa5u8Ddy3sE27YCQNZwCBEKc8n7KlSOAU1eGttfYKmp7zeMlTuNaJgCUSO")
	assert.NilError(t, err)
	assert.Equal(t, data, deres)
}
