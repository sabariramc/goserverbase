package utils_test

import (
	"testing"

	"github.com/sabariramc/goserverbase/v3/utils"
	"gotest.tools/assert"
)

func TestGetHash(t *testing.T) {
	val := "3edcRFV5tgb"
	assert.Equal(t, utils.GetHash(val), utils.GetHash(val))
}
