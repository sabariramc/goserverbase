package utils_test

import (
	"fmt"
	"testing"

	"github.com/sabariramc/goserverbase/v3/utils"
	"gotest.tools/assert"
)

func TestGenerateId(t *testing.T) {
	for i := 0; i < 5000; i++ {
		x := utils.GenerateId(50, "abc_")
		assert.Equal(t, len(x), 50)
		fmt.Println(x)
	}
}

func TestGetHash(t *testing.T) {
	fmt.Println(utils.GetHash("3edcRFV5tgb"))
}
