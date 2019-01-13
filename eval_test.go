package goscheme

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEval(t *testing.T) {
	builtinEnv := setupBuiltinEnv()
	ret, _ := Eval("3", builtinEnv)
	assert.Equal(t, ret, Number(3))
	ret, _ = Eval([]Expression{"define", "x", "3"}, builtinEnv)
	ret, _ = Eval("x", builtinEnv)
	assert.Equal(t, ret, Number(3))
	Eval([]Expression{"define", []Expression{"fn", "y"}, []Expression{"+", "x", "y"}}, builtinEnv)
	ret, _ = Eval([]Expression{"fn", "x"}, builtinEnv)
	assert.Equal(t, Number(6), ret)

	// test begin
	ret, _ = Eval([]Expression{"begin", "1"}, builtinEnv)
	assert.Equal(t, Number(1), ret)
	ret, _ = Eval([]Expression{"begin", "#t"}, builtinEnv)
	assert.Equal(t, true, ret)
	ret, _ = Eval([]Expression{"begin", "1", []Expression{"+", "1", "2", "3"}}, builtinEnv)
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
		ret, _ = Eval(c.input, builtinEnv)
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
		ret, _ = Eval(c.input, builtinEnv)
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
		ret, _ = Eval(c.input, builtinEnv)
		assert.Equal(t, c.expected, ret)
	}

	// test lambda
	ret, _ = EvalAll(strToToken("((lambda (x y) (+ x y)) 1 2)"), builtinEnv)
	assert.Equal(t, Number(3), ret)

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
		ret, _ = Eval(c.input, builtinEnv)
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
		ret, _ = Eval(c.input, builtinEnv)
		assert.Equal(t, c.expected, ret)
	}
	ret, _ = Eval([]Expression{"cons", "1", "2"}, builtinEnv)
	assert.Equal(t, &Pair{Number(1), Number(2)}, ret)

	//// test list
	ret, _ = Eval([]Expression{"list", "1", "2"}, builtinEnv)
	assert.Equal(t, &Pair{Number(1), &Pair{Number(2), NilObj}}, ret)
	ret, _ = Eval([]Expression{"list", "1"}, builtinEnv)
	assert.Equal(t, &Pair{Number(1), NilObj}, ret)
	ret, _ = Eval([]Expression{"list"}, builtinEnv)
	assert.Equal(t, NilObj, ret)
	ret, _ = Eval([]Expression{"list", "1", []Expression{"cons", "1", []Expression{}}}, builtinEnv)
	assert.Equal(t, &Pair{Number(1), &Pair{&Pair{Number(1), NilObj}, NilObj}}, ret)

	//// test append
	ret, _ = EvalAll(strToToken("(append (cons 1 ()) 2)"), builtinEnv)
	assert.Equal(t, &Pair{Number(1), &Pair{Number(2), NilObj}}, ret)
	ret, _ = EvalAll(strToToken("(append () 2)"), builtinEnv)
	assert.Equal(t, &Pair{Number(2), NilObj}, ret)
	ret, _ = EvalAll(strToToken("(append (cons 1 ()) (cons 2 ()))"), builtinEnv)
	assert.Equal(t, &Pair{Number(1), &Pair{Number(2), NilObj}}, ret)
	ret, _ = EvalAll(strToToken("(append (cons 1 ()) ())"), builtinEnv)
	assert.Equal(t, &Pair{Number(1), NilObj}, ret)
	ret, _ = EvalAll(strToToken("(append (cons 1 ()) 2 3)"), builtinEnv)
	assert.Equal(t, &Pair{Number(1), &Pair{Number(2), &Pair{Number(3), NilObj}}}, ret)
	ret, _ = EvalAll(strToToken("(append (cons 1 ()) (cons 2 ()) (cons 3 ()))"), builtinEnv)
	assert.Equal(t, &Pair{Number(1), &Pair{Number(2), &Pair{Number(3), NilObj}}}, ret)

	// test quote
	ret, _ = EvalAll(strToToken("(quote (1 2))"), builtinEnv)
	assert.Equal(t, &Pair{Number(1), &Pair{Number(2), NilObj}}, ret)
	ret, _ = EvalAll(strToToken("(quote 1)"), builtinEnv)
	assert.Equal(t, Number(1), ret)
	ret, _ = EvalAll(strToToken(`(quote "x")`), builtinEnv)
	assert.Equal(t, String("x"), ret)
	ret, _ = EvalAll(strToToken(`(quote (1 "x"))`), builtinEnv)
	assert.Equal(t, &Pair{Number(1), &Pair{String("x"), NilObj}}, ret)
	ret, _ = EvalAll(strToToken(`(quote (cons 1 "x"))`), builtinEnv)
	assert.Equal(t, &Pair{Quote("cons"), &Pair{Number(1), &Pair{String("x"), NilObj}}}, ret)
	ret, _ = EvalAll(strToToken(`(quote (1 (2 3) 4))`), builtinEnv)
	assert.Equal(t, &Pair{
		Number(1),
		&Pair{
			&Pair{Number(2), &Pair{Number(3), NilObj}}, &Pair{Number(4), NilObj}}},
		ret)
	ret, _ = EvalAll(strToToken("'(1 2)"), builtinEnv)
	assert.Equal(t, &Pair{Number(1), &Pair{Number(2), NilObj}}, ret)
	ret, _ = EvalAll(strToToken("'x"), builtinEnv)
	assert.Equal(t, Quote("x"), ret)
	ret, _ = EvalAll(strToToken("'(cons define 3)"), builtinEnv)
	assert.Equal(t, &Pair{Quote("cons"), &Pair{Quote("define"), &Pair{Number(3), NilObj}}}, ret)
	ret, _ = EvalAll(strToToken("''(cons define 3)"), builtinEnv)
	assert.Equal(t, &Pair{Quote("quote"), &Pair{&Pair{Quote("cons"), &Pair{Quote("define"), &Pair{Number(3), NilObj}}}, NilObj}}, ret)
}

// test built in procedures
func TestEval2(t *testing.T) {
	//env := setupBuiltinEnv()
	//_, err := EvalAll(strToToken(`(display-pascal-indents 3 1)`), env)
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
		ret, _ := EvalAll(strToToken(c.input), env)
		assert.Equal(t, c.expected, ret)
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
		ret, _ := EvalAll(strToToken(c.input), env)
		assert.Equal(t, c.expected, ret)
	}
	// test nothing panic
	env := setupBuiltinEnv()
	ret, _ := Eval([]Expression{"load", "\"test.scm\""}, env)
	assert.Equal(t, undefObj, ret)
	ret, _ = Eval([]Expression{"load", "\"test\""}, env)
	assert.Equal(t, undefObj, ret)
	ret, _ = Eval([]Expression{"load", []Expression{"quote", []Expression{"test.scm"}}}, env)
	assert.Equal(t, undefObj, ret)
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
		ret, _ := EvalAll(strToToken(c.input), env)
		assert.Equal(t, c.expected, ret)
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
		{`(and 1 unbound-syntax)`, undefObj},
	}
	for _, c := range testCases {
		env := setupBuiltinEnv()
		ret, _ := EvalAll(strToToken(c.input), env)
		assert.Equal(t, c.expected, ret)
	}
}

// test let, let*, letrec, set!
func TestEval7(t *testing.T) {
	testCases := []struct {
		input    string
		expected Expression
	}{
		{`(let ((x 2) (y 3))
  					(* x y))`, Number(6)},
		{`(let ((x 2) (y 3))
  					(let ((foo (lambda (z) (+ x y z)))
        				(x 7))
    					(foo 4))) `, Number(9)},
		{`(let ((x 2) (y 3))
  					(let* ((x 7)
         				(z (+ x y)))
						(* z x)))`, Number(70)},
		{`(letrec (
					(zero? (lambda (x) (= x 0)))
					(even?
         			(lambda (n)
           			(if (zero? n)
               			#t
							(odd? (- n 1)))))
					(odd?
						(lambda (n)
           				(if (zero? n)
               				#f
               				(even? (- n 1))))))
				(even? 88))`, true},
		{`(letrec ((b a) (a 1)) b)`, undefObj},
		{`(define (f a)
					(let ((b 3)) (set! a 3))
					a)
				(f 4)`, Number(3)},
	}
	for _, c := range testCases {
		env := setupBuiltinEnv()
		ret, _ := EvalAll(strToToken(c.input), env)
		assert.Equal(t, c.expected, ret)
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
