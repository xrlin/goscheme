package goscheme

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEval(t *testing.T) {
	builtinEnv := setupBuiltinEnv()
	//ret := Eval("3", builtinEnv)
	//assert.Equal(t, ret, Number(3))
	//ret = Eval([]Expression{"define", "x", "3"}, builtinEnv)
	//assert.Equal(t, Eval("x", builtinEnv), Number(3))
	//Eval([]Expression{"define", []Expression{"fn", "y"}, []Expression{"+", "x", "y"}}, builtinEnv)
	//assert.Equal(t, Number(6), Eval([]Expression{"fn", "x"}, builtinEnv))

	// test begin
	ret := Eval([]Expression{"begin", "1"}, builtinEnv)
	assert.Equal(t, Number(1), ret)
	ret = Eval([]Expression{"begin", "#t"}, builtinEnv)
	assert.Equal(t, true, ret)
	ret = Eval([]Expression{"begin", "1", []Expression{"+", "1", "2", "3"}}, builtinEnv)
	assert.Equal(t, Number(6), ret)

	// test if
	testCases := []struct {
		input    Expression
		expected Expression
	}{
		{[]Expression{"if", "#t", "1", "0"}, Number(1)},
		{[]Expression{"if", "#f", "1", "0"}, Number(0)},
	}
	for _, c := range testCases {
		ret = Eval(c.input, builtinEnv)
		assert.Equal(t, c.expected, ret)
	}

	// test cond
	testCases = []struct {
		input    Expression
		expected Expression
	}{
		{[]Expression{"cond", []Expression{"#t", "1", "2"}}, Number(2)},
		{[]Expression{"cond", []Expression{"#f", "1", "2"}}, NilObj},
		{[]Expression{"cond", []Expression{"#f", "1", "2"}, []Expression{"#t", "2"}}, Number(2)},
		{[]Expression{"cond", []Expression{"#f", "1", "2"}, []Expression{"else", `"else clause"`}}, `else clause`},
	}
	for _, c := range testCases {
		ret = Eval(c.input, builtinEnv)
		assert.Equal(t, c.expected, ret)
	}

	// test recursion
	tz := NewTokenizerFromString(
		`(define (fact1 n)
					(if (<= n 0)
						1 
						(* n (fact1 (- n 1)))))
				(define (fact2 n) (fact-tail n 1))
				(define (fact-tail n a)
						(if (<= n 0)
							(* 1 a)
							(fact-tail (- n 1) (* n a))))`)
	tokens := tz.Tokens()
	expressions, _ := Parse(&tokens)
	EvalAll(expressions, builtinEnv)
	testCases = []struct {
		input    Expression
		expected Expression
	}{
		{[]Expression{"fact1", "2"}, Number(2)},
		{[]Expression{"fact1", "6"}, Number(720)},
		{[]Expression{"fact1", "0"}, Number(1)},
		// tail recursion
		{[]Expression{"fact2", "2"}, Number(2)},
		{[]Expression{"fact2", "6"}, Number(720)},
		{[]Expression{"fact2", "0"}, Number(1)},
	}
	for _, c := range testCases {
		ret = Eval(c.input, builtinEnv)
		assert.Equal(t, c.expected, ret)
	}
}

func TestIsSyntaxExpression(t *testing.T) {
	assert.Equal(t, true, IsSyntaxExpression([]Expression{"begin"}))
}
