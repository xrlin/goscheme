package goscheme

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPair_IsList(t *testing.T) {
	type testCase struct {
		Item     *Pair
		Expected bool
	}

	testCases := []testCase{
		{&Pair{NilObj, NilObj}, true},
		{&Pair{Car: NilObj, Cdr: NilObj}, true},
		{&Pair{1, NilObj}, true},
		{&Pair{1, 1}, false},
		{&Pair{1, &Pair{1, 2}}, false},
		{&Pair{NilObj, &Pair{1, 2}}, false},
		{&Pair{NilObj, &Pair{1, NilObj}}, true},
		{&Pair{NilObj, &Pair{1, &Pair{3, NilObj}}}, true},
	}
	for _, c := range testCases {
		assert.Equal(t, c.Expected, c.Item.IsList())
	}
}

func TestPair_String(t *testing.T) {
	testCases := []struct {
		Item     *Pair
		Expected string
	}{
		{&Pair{NilObj, NilObj}, "(())"},
		{&Pair{NilObj, 3}, "(() . 3)"},
		{&Pair{1, &Pair{1, 2}}, "(1 1 . 2)"},
		{&Pair{1, &Pair{2, &Pair{3, &Pair{4, NilObj}}}}, "(1 2 3 4)"},
		{&Pair{1, &Pair{NilObj, &Pair{&Pair{2, 3}, &Pair{4, 5}}}}, "(1 () (2 . 3) 4 . 5)"},
	}
	for _, c := range testCases {
		assert.Equal(t, c.Expected, c.Item.String())
	}
}
