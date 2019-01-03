package goscheme

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEval(t *testing.T) {
	builtinEnv := setupBuiltinEnv()
	ret := Eval("3", builtinEnv)
	assert.Equal(t, ret, Number(3))
	ret = Eval([]Expression{"define", "x", "3"}, builtinEnv)
	assert.Equal(t, Eval("x", builtinEnv), Number(3))
	Eval([]Expression{"define", []Expression{"fn", "y"}, []Expression{"+", "x", "y"}}, builtinEnv)
	assert.Equal(t, Number(6), Eval([]Expression{"fn", "x"}, builtinEnv))

	// test begin
	ret = Eval([]Expression{"begin", "1"}, builtinEnv)
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

	// test string
	testCases = []struct {
		input    Expression
		expected Expression
	}{
		{`"test"`, "test"},
		{"\"test\r\n\"", "test\r\n"},
		{"\"test\n\"", "test\n"},
	}
	for _, c := range testCases {
		ret = Eval(c.input, builtinEnv)
		assert.Equal(t, c.expected, ret)
	}

	// test lambda
	assert.Equal(t, Number(3), EvalAll(strToToken("((lambda (x y) (+ x y)) 1 2)"), builtinEnv))

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

	// test cons
	testCases = []struct {
		input    Expression
		expected Expression
	}{
		{[]Expression{"cons", "1", "2"}, &Pair{Number(1), Number(2)}},
	}
	for _, c := range testCases {
		ret = Eval(c.input, builtinEnv)
		assert.Equal(t, c.expected, ret)
	}
	assert.Equal(t, &Pair{Number(1), Number(2)}, Eval([]Expression{"cons", "1", "2"}, builtinEnv))

	//// test list
	assert.Equal(t, &Pair{Number(1), &Pair{Number(2), NilObj}}, Eval([]Expression{"list", "1", "2"}, builtinEnv))
	assert.Equal(t, &Pair{Number(1), NilObj}, Eval([]Expression{"list", "1"}, builtinEnv))
	assert.Equal(t, NilObj, Eval([]Expression{"list"}, builtinEnv))
	assert.Equal(t, &Pair{Number(1), &Pair{&Pair{Number(1), NilObj}, NilObj}}, Eval([]Expression{"list", "1", []Expression{"cons", "1", []Expression{}}}, builtinEnv))

	//// test append
	assert.Equal(t, &Pair{Number(1), &Pair{Number(2), NilObj}}, EvalAll(strToToken("(append (cons 1 ()) 2)"), builtinEnv))
	assert.Equal(t, &Pair{Number(2), NilObj}, EvalAll(strToToken("(append () 2)"), builtinEnv))
	assert.Equal(t, &Pair{Number(1), &Pair{Number(2), NilObj}}, EvalAll(strToToken("(append (cons 1 ()) (cons 2 ()))"), builtinEnv))
	assert.Equal(t, &Pair{Number(1), NilObj}, EvalAll(strToToken("(append (cons 1 ()) ())"), builtinEnv))
	assert.Equal(t, &Pair{Number(1), &Pair{Number(2), &Pair{Number(3), NilObj}}}, EvalAll(strToToken("(append (cons 1 ()) 2 3)"), builtinEnv))
	assert.Equal(t, &Pair{Number(1), &Pair{Number(2), &Pair{Number(3), NilObj}}}, EvalAll(strToToken("(append (cons 1 ()) (cons 2 ()) (cons 3 ()))"), builtinEnv))

	// test quote
	assert.Equal(t, &Pair{Number(1), &Pair{Number(2), NilObj}}, EvalAll(strToToken("(quote (1 2))"), builtinEnv))
	assert.Equal(t, Number(1), EvalAll(strToToken("(quote 1)"), builtinEnv))
	assert.Equal(t, "x", EvalAll(strToToken(`(quote "x")`), builtinEnv))
	assert.Equal(t, &Pair{Number(1), &Pair{"x", NilObj}}, EvalAll(strToToken(`(quote (1 "x"))`), builtinEnv))
	assert.Equal(t, &Pair{Quote("cons"), &Pair{Number(1), &Pair{"x", NilObj}}}, EvalAll(strToToken(`(quote (cons 1 "x"))`), builtinEnv))
	assert.Equal(t, &Pair{
		Number(1),
		&Pair{
			&Pair{Number(2), &Pair{Number(3), NilObj}}, &Pair{Number(4), NilObj}}},
		EvalAll(strToToken(`(quote (1 (2 3) 4))`), builtinEnv))
	assert.Equal(t, &Pair{Number(1), &Pair{Number(2), NilObj}}, EvalAll(strToToken("'(1 2)"), builtinEnv))
	assert.Equal(t, Quote("x"), EvalAll(strToToken("'x"), builtinEnv))
	assert.Equal(t, &Pair{Quote("cons"), &Pair{Quote("define"), &Pair{Number(3), NilObj}}}, EvalAll(strToToken("'(cons define 3)"), builtinEnv))
	assert.Equal(t, &Pair{Quote("quote"), &Pair{&Pair{Quote("cons"), &Pair{Quote("define"), &Pair{Number(3), NilObj}}}, NilObj}}, EvalAll(strToToken("''(cons define 3)"), builtinEnv))
}

func TestIsSyntaxExpression(t *testing.T) {
	assert.Equal(t, true, IsSyntaxExpression([]Expression{"begin"}))
}

func strToToken(input string) []Expression {
	tz := NewTokenizerFromString(input)
	tokens := tz.Tokens()
	expressions, _ := Parse(&tokens)
	return expressions
}
