package goscheme

import (
	"errors"
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
		{`"a string"`, []string{`"a string"`}},
		{`"a string\n\\"`, []string{"\"a string\n\\\""}},
		{`"paragraph 1
				paragraph 2"`, []string{`"paragraph 1
				paragraph 2"`}},
		{`(display "string ")`, []string{"(", "display", `"string "`, ")"}},
		{"3(display 1)", []string{"3", "(", "display", "1", ")"}},
		{"(define x 3)", []string{"(", "define", "x", "3", ")"}},
		{"(lambda (x y) (display x))", []string{"(", "lambda", "(", "x", "y", ")", "(", "display", "x", ")", ")"}},
		{`(lambda (x y) 
					(display x))`, []string{"(", "lambda", "(", "x", "y", ")", "(", "display", "x", ")", ")"}},
		{"(define (func x) (define (intern x) (x)))",
			[]string{"(", "define", "(", "func", "x", ")", "(", "define", "(", "intern", "x", ")", "(", "x", ")", ")", ")"}},
	}
	for _, c := range testCases {
		assert.Equal(t, c.expected, Tokenize(c.input))
	}
}

func TestParse(t *testing.T) {
	testCases := []struct {
		input    []string
		expected Expression
		err      error
	}{
		{[]string{"(", ")"}, []Expression{[]Expression{}}, nil},
		{[]string{"3"}, []Expression{"3"}, nil},
		{[]string{"3", "(", "define", "x", "1", ")"}, []Expression{"3", []Expression{"define", "x", "1"}}, nil},
		{[]string{"3", "(", "define", "x", "1"}, []Expression{"3", []Expression{"define", "x", 1}}, errors.New("syntax error")},
		{[]string{"(", "define", "x", "3", ")"}, []Expression{[]Expression{"define", "x", "3"}}, nil},
		{[]string{"(", "define", "(", "func", "x", ")", "(", "define", "(", "intern", "x", ")", "(", "x", ")", ")", ")"},
			[]Expression{[]Expression{"define",
				[]Expression{"func", "x"}, []Expression{"define", []Expression{"intern", "x"}, []Expression{"x"}}}}, nil},
	}
	for _, c := range testCases {
		ret, err := Parse(&c.input)
		if c.err != nil {
			assert.NotNil(t, err)
			continue
		}
		assert.Equal(t, c.expected, ret)
	}
}
