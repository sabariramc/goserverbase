package utils_test

import (
	"bytes"
	"testing"

	"github.com/sabariramc/goserverbase/v6/utils"
	"gotest.tools/assert"
)

type P struct {
	X, Y, Z int
	Name    string
}

type Q struct {
	X, Y *int32
	Name string
}

func TestEncode(t *testing.T) {
	var network bytes.Buffer
	p := P{3, 4, 5, "Pythagoras"}
	err := utils.Encode(p, &network)
	assert.NilError(t, err)
	var q Q
	err = utils.Decode(&network, &q)
	assert.NilError(t, err)
	assert.Equal(t, p.X, int(*q.X))
	assert.Equal(t, p.Y, int(*q.Y))
	assert.Equal(t, p.Name, q.Name)
}
