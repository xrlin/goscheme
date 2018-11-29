package goscheme

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEnv_Find(t *testing.T) {
	env := &Env{
		outer: nil,
		frame: map[Symbol]Expression{"x": 1, "y": 1},
	}
	ret, err := env.Find("x")
	assert.Equal(t, 1, ret)
	assert.Nil(t, err)

	// nested env

	env2 := &Env{
		outer: env,
		frame: map[Symbol]Expression{"x": 2},
	}
	ret, err = env2.Find("x")
	assert.Equal(t, 2, ret)
	assert.Nil(t, err)
	ret, err = env2.Find("y")
	assert.Equal(t, 1, ret)
	assert.Nil(t, err)
	ret, err = env2.Find("unknown")
	assert.NotNil(t, err)

}
