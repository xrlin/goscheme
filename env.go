package goscheme

import (
	"errors"
	"fmt"
	"os"
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

// Symbols returns the bound symbols including the outer frame
func (e *Env) Symbols() []Symbol {
	var ret []Symbol
	for k := range e.frame {
		ret = append(ret, k)
	}
	if e.outer != nil {
		ret = append(ret, e.outer.Symbols()...)
	}
	return ret
}

func exitFunc(args ...Expression) (Expression, error) {
	exit <- os.Interrupt
	return NilObj, nil
}

func addFunc(args ...Expression) (Expression, error) {
	var ret Number
	for _, arg := range args {
		num, err := expressionToNumber(arg)
		if err != nil {
			return 0, err
		}
		ret = ret + num
	}
	return ret, nil
}

func minusFunc(args ...Expression) (Expression, error) {
	var ret Number
	ret, err := expressionToNumber(args[0])
	if err != nil {
		return undefObj, err
	}
	for i, arg := range args {
		if i == 0 {
			continue
		}
		num, err := expressionToNumber(arg)
		if err != nil {
			return undefObj, err
		}
		ret = ret - num
	}
	return ret, nil
}

func plusFunc(args ...Expression) (Expression, error) {
	var ret Number = 1
	for _, arg := range args {
		num, err := expressionToNumber(arg)
		if err != nil {
			return undefObj, err
		}
		ret = ret * num
	}
	return ret, nil
}

func divFunc(args ...Expression) (Expression, error) {
	var ret Number = 1
	ret, err := expressionToNumber(args[0])
	if err != nil {
		return undefObj, err
	}
	for i, arg := range args {
		if i == 0 {
			continue
		}
		num, err := expressionToNumber(arg)
		if err != nil {
			return undefObj, err
		}
		ret = ret / num
	}
	return ret, nil
}

func eqlFunc(args ...Expression) (Expression, error) {
	if args[0] == args[1] {
		return true, nil
	}
	return false, nil
}

func lessFunc(args ...Expression) (Expression, error) {
	op1, err := expressionToNumber(args[0])
	if err != nil {
		return undefObj, err
	}
	op2, err := expressionToNumber(args[1])
	if err != nil {
		return undefObj, err
	}
	if op1 < op2 {
		return true, nil
	}
	return false, nil
}

func greaterFunc(args ...Expression) (Expression, error) {
	op1, err := expressionToNumber(args[0])
	if err != nil {
		return undefObj, err
	}
	op2, err := expressionToNumber(args[1])
	if err != nil {
		return undefObj, err
	}
	if op1 > op2 {
		return true, nil
	}
	return false, nil
}

func lessEqualFunc(args ...Expression) (Expression, error) {
	op1, err := expressionToNumber(args[0])
	if err != nil {
		return undefObj, err
	}
	op2, err := expressionToNumber(args[1])
	if err != nil {
		return undefObj, err
	}
	if op1 <= op2 {
		return true, nil
	}
	return false, nil
}

func greatEqualFunc(args ...Expression) (Expression, error) {
	op1, err := expressionToNumber(args[0])
	if err != nil {
		return undefObj, err
	}
	op2, err := expressionToNumber(args[1])
	if err != nil {
		return undefObj, err
	}
	if op1 >= op2 {
		return true, nil
	}
	return false, nil
}

func displayFunc(args ...Expression) (Expression, error) {
	exp := args[0]
	switch v := exp.(type) {
	case String:
		fmt.Print(string(v))
	default:
		fmt.Printf("%v", valueToString(v))
	}
	return undefObj, nil
}

func displaylnFunc(args ...Expression) (Expression, error) {
	ret, err := displayFunc(args...)
	fmt.Println()
	return ret, err
}

func isNullFunc(args ...Expression) (Expression, error) {
	return isNullExp(args[0]), nil
}

func isStringFunc(args ...Expression) (Expression, error) {
	exp := args[0]
	return IsString(exp), nil
}

func notFunc(args ...Expression) (Expression, error) {
	return !IsTrue(args[0]), nil
}

// concatFunc concat the strings
func concatFunc(args ...Expression) (Expression, error) {
	var ret String
	for _, arg := range args {
		v := arg
		s, ok := v.(String)
		if !ok {
			return undefObj, errors.New(fmt.Sprintf("argument %v is not a String", v))
		}
		ret += s
	}
	return ret, nil
}

func checkThunkFunc(args ...Expression) (Expression, error) {
	return IsThunk(args[0]), nil
}

func forceFunc(args ...Expression) (Expression, error) {
	return ActualValue(args[0])
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
	//"and":       NewFunction("and", andFunc, 1, -1),
	//"or":        NewFunction("or", orFunc, 1, -1),
	"cons":     NewFunction("cons", consImpl, 2, 2),
	"car":      NewFunction("car", carImpl, 1, 1),
	"cdr":      NewFunction("cdr", cdrImpl, 1, 1),
	"list":     NewFunction("list", listImpl, -1, -1),
	"append":   NewFunction("append", appendImpl, 2, -1),
	"set-car!": NewFunction("set-car!", setCarImpl, 2, 2),
	"set-cdr!": NewFunction("set-cdr!", setCdrImpl, 2, 2),
	"concat":   NewFunction("concat", concatFunc, 2, -1),
	"thunk?":   NewFunction("thunk?", checkThunkFunc, 1, 1),
	"force":    NewFunction("thunk?", forceFunc, 1, 1),
}

func setCarImpl(args ...Expression) (Expression, error) {
	exp := args[0]
	newValue := args[1]
	switch p := exp.(type) {
	case *Pair:
		p.Car = newValue
	default:
		return undefObj, errors.New(fmt.Sprintf("%v is not a pair", exp))
	}
	return undefObj, nil
}

func setCdrImpl(args ...Expression) (Expression, error) {
	exp := args[0]
	newValue := args[1]
	switch p := exp.(type) {
	case *Pair:
		p.Cdr = newValue
	default:
		return undefObj, errors.New(fmt.Sprintf("%v is not a pair", exp))
	}
	return undefObj, nil
}

func consImpl(args ...Expression) (Expression, error) {
	return &Pair{args[0], args[1]}, nil
}

func listImpl(args ...Expression) (Expression, error) {
	if len(args) == 0 {
		return NilObj, nil
	}
	cdr, err := listImpl(args[1:]...)
	if err != nil {
		return NilObj, err
	}
	return &Pair{args[0], cdr}, nil
}

func carImpl(args ...Expression) (Expression, error) {
	v := args[0]
	switch p := v.(type) {
	case *Pair:
		return p.Car, nil
	default:
		return undefObj, errors.New("argument is not a pair")
	}
}

func cdrImpl(args ...Expression) (Expression, error) {
	v := args[0]
	switch p := v.(type) {
	case *Pair:
		return p.Cdr, nil
	default:
		return undefObj, errors.New("argument is not a pair")
	}
}

func appendImpl(args ...Expression) (ret Expression, err error) {
	ret = args[0]
	for i := 1; i < len(args); i++ {
		ret, err = merge(ret, args[i])
		if err != nil {
			return undefObj, err
		}
	}
	return ret, nil
}

// append arg2 to arg1 and return the new *pair
func merge(arg1, arg2 Expression) (Expression, error) {
	if !isList(arg1) {
		return undefObj, errors.New("not a list")
	}
	if isNullExp(arg1) {
		if isList(arg2) {
			return arg2, nil
		}
		return listImpl(arg2)
	}
	rest, err := cdrImpl(arg1)
	if err != nil {
		return undefObj, err
	}
	newArg, err := merge(rest, arg2)
	if err != nil {
		return undefObj, err
	}
	car, err := carImpl(arg1)
	if err != nil {
		return NilObj, err
	}
	return consImpl(car, newArg)
}

func isList(exp Expression) bool {
	switch l := exp.(type) {
	case *Pair:
		return l.IsList()
	case NilType:
		return true
	default:
		return false
	}
}

func setupBuiltinEnv() *Env {
	initSyntax()
	var builtinEnv = &Env{
		outer: nil,
		frame: make(map[Symbol]Expression),
	}
	for key, syntax := range SyntaxMap {
		builtinEnv.Set(Symbol(key), syntax)
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
