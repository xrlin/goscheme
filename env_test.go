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

func Test_listImpl(t *testing.T) {
	testCases := []struct {
		input    []Expression
		expected *Pair
	}{
		{[]Expression{1, 2, 3}, &Pair{1, &Pair{2, &Pair{3, NilObj}}}},
		{[]Expression{1, &Pair{Car: 2}, 3}, &Pair{1, &Pair{&Pair{Car: 2}, &Pair{3, NilObj}}}},
	}
	for _, c := range testCases {
		assert.Equal(t, c.expected, listImpl(c.input...))
	}
}

func Test_appendImpl(t *testing.T) {
	testCases := []struct {
		input    []Expression
		expected *Pair
	}{
		{[]Expression{&Pair{1, NilObj}, 2}, &Pair{1, &Pair{2, NilObj}}},
		{[]Expression{&Pair{1, NilObj}, &Pair{2, NilObj}}, &Pair{1, &Pair{2, NilObj}}},
		{[]Expression{&Pair{1, NilObj}, &Pair{2, NilObj}, 3}, &Pair{1, &Pair{2, &Pair{3, NilObj}}}},
	}
	for _, c := range testCases {
		assert.Equal(t, c.expected, appendImpl(c.input...))
	}
}
