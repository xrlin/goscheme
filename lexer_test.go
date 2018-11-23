package goscheme

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTokenize(t *testing.T) {
	testCases := []struct {
		input    string
		expected []string
	}{
		{"()", []string{"(", ")"}},
		{"( )", []string{"(", ")"}},
		{"(define x 3)", []string{"(", "define", "x", "3", ")"}},
		{"(lambda (x y) (display x))", []string{"(", "lambda", "(", "x", "y", ")", "(", "display", "x", ")", ")"}},
		{"(define (func x) (define (intern x) (x)))",
			[]string{"(", "define", "(", "func", "x", ")", "(", "define", "(", "intern", "x", ")", "(", "x", ")", ")", ")"}},
	}
	for _, c := range testCases {
		assert.Equal(t, c.expected, Tokenize(c.input))
	}
}
