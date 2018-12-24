package goscheme

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNeededIndents(t *testing.T) {
	testCases := []struct {
		input    string
		expected int
	}{
		{input: "fn", expected: 0},
		{input: "(fn", expected: 1},
		{input: "(fn x)", expected: 0},
		{input: `(fn
					x)`, expected: 0},
	}
	for _, c := range testCases {
		ret := neededIndents(bytes.NewReader([]byte(c.input)))
		assert.Equal(t, c.expected, ret)
	}
}
