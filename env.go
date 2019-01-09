package goscheme

import (
	"fmt"
	"go/types"
	"os"
	"strconv"
)

type Env struct {
	outer *Env
	frame map[Symbol]Expression
}

func (e *Env) Find(symbol Symbol) (Expression, error) {
	ret, ok := e.frame[symbol]
	if ok {
		return ret, nil
	} else {
		if e.outer == nil {
			return nil, fmt.Errorf("symbol %v unbound", symbol)
		} else {
			return e.outer.Find(symbol)
		}
	}
}

func (e *Env) Set(symbol Symbol, value Expression) {
	e.frame[symbol] = value
}

func exitFunc(args ...Expression) Expression {
	exit <- os.Interrupt
	return NilObj
}

func addFunc(args ...Expression) Expression {
	var ret Number
	for _, e := range args {
		if !IsNumber(e) {
			panic(fmt.Sprintf("%v is not a number", e))
		}
	}
	for _, arg := range args {
		ret = ret + expressionToNumber(arg)
	}
	return ret
}

func minusFunc(args ...Expression) Expression {
	var ret Number
	for _, e := range args {
		if !IsNumber(e) {
			panic(fmt.Sprintf("%v is not a number", e))
		}
	}
	ret = expressionToNumber(args[0])
	for i, arg := range args {
		if i == 0 {
			continue
		}
		ret = ret - expressionToNumber(arg)
	}
	return ret
}

func plusFunc(args ...Expression) Expression {
	var ret Number = 1
	for _, e := range args {
		if !IsNumber(e) {
			panic(fmt.Sprintf("%v is not a number", e))
		}
	}
	for _, arg := range args {
		ret = ret * expressionToNumber(arg)
	}
	return ret
}

func divFunc(args ...Expression) Expression {
	var ret Number = 1
	for _, e := range args {
		if !IsNumber(e) {
			panic(fmt.Sprintf("%v is not a number", e))
		}
	}
	ret = expressionToNumber(args[0])
	for i, arg := range args {
		if i == 0 {
			continue
		}
		ret = ret / expressionToNumber(arg)
	}
	return ret
}

func eqlFunc(args ...Expression) Expression {
	if args[0] == args[1] {
		return true
	}
	return false
}

func lessFunc(args ...Expression) Expression {
	op1 := expressionToNumber(args[0])
	op2 := expressionToNumber(args[1])
	if op1 < op2 {
		return true
	}
	return false
}

func greaterFunc(args ...Expression) Expression {
	op1 := expressionToNumber(args[0])
	op2 := expressionToNumber(args[1])
	if op1 > op2 {
		return true
	}
	return false
}

func lessEqualFunc(args ...Expression) Expression {
	op1 := expressionToNumber(args[0])
	op2 := expressionToNumber(args[1])
	if op1 <= op2 {
		return true
	}
	return false
}

func greatEqualFunc(args ...Expression) Expression {
	op1 := expressionToNumber(args[0])
	op2 := expressionToNumber(args[1])
	if op1 >= op2 {
		return true
	}
	return false
}

func displayFunc(args ...Expression) Expression {
	exp := args[0]
	switch v := exp.(type) {
	case String:
		fmt.Print(string(v))
	default:
		fmt.Printf("%v", toString(v))
	}
	return undefObj
}

func displaylnFunc(args ...Expression) Expression {
	ret := displayFunc(args...)
	fmt.Println()
	return ret
}

func isNullFunc(args ...Expression) Expression {
	return isNullExp(args[0])
}

func isStringFunc(args ...Expression) Expression {
	exp := args[0]
	return IsString(exp)
}

func notFunc(args ...Expression) Expression {
	return !IsTrue(args[0])
}

func andFunc(args ...Expression) Expression {
	for _, exp := range args {
		if !IsTrue(exp) {
			return false
		}
	}
	return true
}

func orFunc(args ...Expression) Expression {
	for _, exp := range args {
		if IsTrue(exp) {
			return true
		}
	}
	return false
}

// concatFunc concat the strings
func concatFunc(args ...Expression) Expression {
	var ret String
	for _, arg := range args {
		s, ok := arg.(String)
		if !ok {
			panic(fmt.Sprintf("argument %v is not a String", arg))
		}
		ret += s
	}
	return ret
}

var builtinFunctions = map[Symbol]Function{
	"exit":      NewFunction("exit", exitFunc, 0, 0),
	"+":         NewFunction("+", addFunc, 1, -1),
	"-":         NewFunction("-", minusFunc, 1, -1),
	"*":         NewFunction("*", plusFunc, 1, -1),
	"/":         NewFunction("/", divFunc, 1, -1),
	"=":         NewFunction("=", eqlFunc, 2, 2),
	"<":         NewFunction("<", lessFunc, 2, 2),
	">":         NewFunction(">", greaterFunc, 2, 2),
	"<=":        NewFunction("<=", lessEqualFunc, 2, 2),
	">=":        NewFunction(">=", greatEqualFunc, 2, 2),
	"display":   NewFunction("display", displayFunc, 1, 1),
	"displayln": NewFunction("displayln", displaylnFunc, 1, 1),
	"null?":     NewFunction("null?", isNullFunc, 1, 1),
	"string?":   NewFunction("string?", isStringFunc, 1, 1),
	"not":       NewFunction("not", notFunc, 1, 1),
	"and":       NewFunction("and", andFunc, 1, -1),
	"or":        NewFunction("or", orFunc, 1, -1),
	"cons":      NewFunction("cons", consImpl, 2, 2),
	"car":       NewFunction("car", carImpl, 1, 1),
	"cdr":       NewFunction("cdr", cdrImpl, 1, 1),
	"list":      NewFunction("list", listImpl, -1, -1),
	"append":    NewFunction("append", appendImpl, 2, -1),
	"set-car!":  NewFunction("set-car!", setCarImpl, 2, 2),
	"set-cdr!":  NewFunction("set-cdr!", setCdrImpl, 2, 2),
	"concat":    NewFunction("concat", concatFunc, 2, -1),
}

func setCarImpl(args ...Expression) Expression {
	exp := args[0]
	newValue := args[1]
	switch p := exp.(type) {
	case *Pair:
		p.Car = newValue
	default:
		panic(fmt.Sprintf("%v is not a pair", exp))
	}
	return undefObj
}

func setCdrImpl(args ...Expression) Expression {
	exp := args[0]
	newValue := args[1]
	switch p := exp.(type) {
	case *Pair:
		p.Cdr = newValue
	default:
		panic(fmt.Sprintf("%v is not a pair", exp))
	}
	return undefObj
}

func consImpl(args ...Expression) Expression {
	if len(args) != 2 {
		panic("cons takes 2 arguments but " + strconv.Itoa(len(args)) + " provided")
	}
	return &Pair{args[0], args[1]}
}

func listImpl(args ...Expression) Expression {
	if len(args) == 0 {
		return NilObj
	}
	return &Pair{args[0], listImpl(args[1:]...)}
}

func carImpl(args ...Expression) Expression {
	v := args[0]
	switch p := v.(type) {
	case *Pair:
		return p.Car
	default:
		panic("argument is not a pair")
	}
}

func cdrImpl(args ...Expression) Expression {
	v := args[0]
	switch p := v.(type) {
	case *Pair:
		return p.Cdr
	default:
		panic("argument is not a pair")
	}
}

func appendImpl(args ...Expression) Expression {
	if len(args) < 2 {
		panic("must provide 2 arguments")
	}
	ret := args[0]
	for i := 1; i < len(args); i++ {
		ret = merge(ret, args[i])
	}
	return ret
}

// append arg2 to arg1 and return the new *pair
func merge(arg1, arg2 Expression) Expression {
	if !isList(arg1) {
		panic("not a list")
	}
	if isNullExp(arg1) {
		if isList(arg2) {
			return arg2
		}
		return listImpl(arg2)
	}
	return consImpl(carImpl(arg1), merge(cdrImpl(arg1), arg2))
}

func isList(exp Expression) bool {
	switch l := exp.(type) {
	case *Pair:
		return l.IsList()
	case NilType:
		return true
	case types.Nil:
		return true
	default:
		return false
	}
}

func setupBuiltinEnv() *Env {
	var builtinEnv = &Env{
		outer: nil,
		frame: make(map[Symbol]Expression),
	}
	for k, fn := range builtinFunctions {
		builtinEnv.Set(k, fn)
	}
	loadBuiltinProcedures(builtinEnv)
	return builtinEnv
}

const builtinProcedures = `
(define (map procedure list-arguments)
	(cond
		((null? list-arguments) '())
		(else 
			(cons (procedure (car list-arguments)) 
					(map procedure 
							(cdr list-arguments))))))

(define (filter predicate sequence)
  (cond ((null? sequence) '())
        ((predicate (car sequence))
         (cons (car sequence)
               (filter predicate (cdr sequence))))
        (else (filter predicate (cdr sequence)))))

(define (reduce proc items)
  (if (null? items)
      0
      (proc (car items) (reduce proc (cdr items)))))

(define (remainder a b)
  (if (< a b)
      a
      (remainder (- a b) b)))

(define list-ref
    (lambda (lst place)
      (if (null? lst)
          '()
          (if (= place 0)
          (car lst)
          (list-ref (cdr lst) (- place 1))))))

(define (list-set! list k val)
    (if (= k 0)
        (set-car! list val)
        (list-set! (cdr list) (- k 1) val)))

(define (list-length lst)
	(if (null? lst) 0 (+ (list-length (cdr lst)) 1)))

`

func loadBuiltinProcedures(env *Env) {
	t := NewTokenizerFromString(builtinProcedures)
	tokens := t.Tokens()
	expressions, _ := Parse(&tokens)
	EvalAll(expressions, env)
}
