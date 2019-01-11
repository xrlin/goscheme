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
		{[]Expression{"cond", []Expression{"#f", "1", "2"}}, undefObj},
		{[]Expression{"cond", []Expression{"#f", "1", "2"}, []Expression{"#t", "2"}}, Number(2)},
		{[]Expression{"cond", []Expression{"#f", "1", "2"}, []Expression{"else", `"else clause"`}}, String(`else clause`)},
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
		{`"test"`, String("test")},
		{"\"test\r\n\"", String("test\r\n")},
		{"\"test\n\"", String("test\n")},
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
	assert.Equal(t, String("x"), EvalAll(strToToken(`(quote "x")`), builtinEnv))
	assert.Equal(t, &Pair{Number(1), &Pair{String("x"), NilObj}}, EvalAll(strToToken(`(quote (1 "x"))`), builtinEnv))
	assert.Equal(t, &Pair{Quote("cons"), &Pair{Number(1), &Pair{String("x"), NilObj}}}, EvalAll(strToToken(`(quote (cons 1 "x"))`), builtinEnv))
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

// test built in procedures
func TestEval2(t *testing.T) {
	env := setupBuiltinEnv()
	EvalAll(strToToken(`(display-pascal-indents 3 1)`), env)
}

// test comment
func TestEval3(t *testing.T) {
	testCases := []struct {
		input    string
		expected Expression
	}{
		{`
				;comment
				3`, Number(3)},
		{`
				;comment
				3
				;comment`, Number(3)},
		{`
				; comment
				(define x 3)
				; comment 2
				x`, Number(3)},
		{`
				;comment
				(define (func ; comment
						 x)
					x)
				(func 3)`, Number(3)},
		{`
				;comment
				(define (func ; comment
						 x)
					x)
				(func ";not a comment#f\n")`, String(";not a comment#f\n")},
	}
	env := setupBuiltinEnv()
	for _, c := range testCases {
		assert.Equal(t, c.expected, EvalAll(strToToken(c.input), env))
	}
}

// test eval/apply/load syntax
func TestEval4(t *testing.T) {
	testCases := []struct {
		input    string
		expected Expression
	}{
		{`(eval 3)`, Number(3)},
		{`(eval '3)`, Number(3)},
		{`(eval '(begin (display "") 3))`, Number(3)},
		{`
			(define fn '*)
(define x 3)
(define y (list '+ x 5))
(define z (list fn 10 y))
(eval y)`, Number(8)},
		{`
			(define fn '*)
(define x 3)
(define y (list '+ x 5))
(define z (list fn 10 y))
(eval z)`, Number(80)},
		{`(define x 3) (eval 'x)`, Number(3)},
		{`(define x 3) (eval ''x)`, Quote("x")},
		{`(apply display '(3))`, undefObj},
		{`(apply (lambda x x) '(3))`, Number(3)},
		{`(apply (lambda (x y) (+ x y)) '(3 4))`, Number(7)},
	}
	for _, c := range testCases {
		env := setupBuiltinEnv()
		assert.Equal(t, c.expected, EvalAll(strToToken(c.input), env))
	}
	// test nothing panic
	env := setupBuiltinEnv()
	assert.Equal(t, undefObj, Eval([]Expression{"load", "\"test.scm\""}, env))
	assert.Equal(t, undefObj, Eval([]Expression{"load", "\"test\""}, env))
	assert.Equal(t, undefObj, Eval([]Expression{"load", []Expression{"quote", []Expression{"test.scm"}}}, env))
	_, err := env.Find("test-method")
	assert.Nil(t, err)
}

// test lazy evaluation
func TestEval5(t *testing.T) {
	testCases := []struct {
		input    string
		expected Expression
	}{
		{`(thunk? (delay (+ 1 2)))`, true},
		{`(force (delay (+ 1 2)))`, Number(3)},
		{`(define (try a b) (if (= a 0) b a)) (try 1 (delay (+ 1 "x")))`, Number(1)},
		// eval error
		{`(define (try a b) (if (= a 0) (force b) (force a))) (try 0 (delay (+ 1 "x")))`, undefObj},
	}
	for _, c := range testCases {
		env := setupBuiltinEnv()
		assert.Equal(t, c.expected, EvalAll(strToToken(c.input), env))
	}
}

// test short circuit evaluation
func TestEval6(t *testing.T) {
	testCases := []struct {
		input    string
		expected Expression
	}{
		{`(or 1 unbound-syntax)`, true},
		// eval error
		{`(and 1 unbound-syntax)`, nil},
	}
	for _, c := range testCases {
		env := setupBuiltinEnv()
		assert.Equal(t, c.expected, EvalAll(strToToken(c.input), env))
	}
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
