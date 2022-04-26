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
