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

var builtinFuncs = map[Symbol]Function{
	"exit": Function(func(args ...Expression) Expression {
		exit <- os.Interrupt
		return NilObj
	}),
	"+": Function(func(args ...Expression) Expression {
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
	}),
	"-": Function(func(args ...Expression) Expression {
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
	}),
	"*": Function(func(args ...Expression) Expression {
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
	}),
	"/": Function(func(args ...Expression) Expression {
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
	}),
	"=": Function(func(args ...Expression) Expression {
		if args[0] == args[1] {
			return true
		}
		return false
	}),
	"<": Function(func(args ...Expression) Expression {
		op1 := expressionToNumber(args[0])
		op2 := expressionToNumber(args[1])
		if op1 < op2 {
			return true
		}
		return false
	}),
	">": Function(func(args ...Expression) Expression {
		op1 := expressionToNumber(args[0])
		op2 := expressionToNumber(args[1])
		if op1 > op2 {
			return true
		}
		return false
	}),
	"<=": Function(func(args ...Expression) Expression {
		op1 := expressionToNumber(args[0])
		op2 := expressionToNumber(args[1])
		if op1 <= op2 {
			return true
		}
		return false
	}),
	">=": Function(func(args ...Expression) Expression {
		op1 := expressionToNumber(args[0])
		op2 := expressionToNumber(args[1])
		if op1 >= op2 {
			return true
		}
		return false
	}),
	"display": Function(func(args ...Expression) Expression {
		exp := args[0]
		switch v := exp.(type) {
		case String:
			fmt.Print(string(v))
		default:
			fmt.Printf("%v", v)
		}
		return undefObj
	}),
	"displayln": Function(func(args ...Expression) Expression {
		exp := args[0]
		switch v := exp.(type) {
		case String:
			fmt.Println(string(v))
		default:
			fmt.Printf("%v\n", v)
		}
		return undefObj
	}),
	"true?": Function(func(args ...Expression) Expression {
		if len(args) != 1 {
			panic("true? require 1 argument")
		}
		return IsTrue(args[0])
	}),
	"null?": Function(func(args ...Expression) Expression {
		if len(args) != 1 {
			panic("null? require 1 argument")
		}
		return isNullExp(args[0])
	}),
	"string?": Function(func(args ...Expression) Expression {
		exp := args[0]
		return IsString(exp)
	}),
	"not": Function(func(args ...Expression) Expression {
		if len(args) != 1 {
			panic("not require 1 argument")
		}
		return !IsTrue(args[0])
	}),
	"and": Function(func(args ...Expression) Expression {
		if len(args) == 0 {
			panic("and require more than 1 arguments")
		}
		for _, exp := range args {
			if !IsTrue(exp) {
				return false
			}
		}
		return true
	}),
	"or": Function(func(args ...Expression) Expression {
		if len(args) == 0 {
			panic("or require more than 1 arguments")
		}
		for _, exp := range args {
			if IsTrue(exp) {
				return true
			}
		}
		return false
	}),
	"cons":   Function(consImpl),
	"car":    Function(carImpl),
	"cdr":    Function(cdrImpl),
	"list":   Function(listImpl),
	"append": Function(appendImpl),
	//"eval": Function(func(args ...Expression) Expression {
	//	return EvalAll(args)
	//})
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
	if len(args) != 1 {
		panic("require 1 argument")
	}
	v := args[0]
	switch p := v.(type) {
	case *Pair:
		return p.Car
	default:
		panic("argument is not a pair")
	}
}

func cdrImpl(args ...Expression) Expression {
	if len(args) != 1 {
		panic("require 1 argument")
	}
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
	for k, fn := range builtinFuncs {
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
`

func loadBuiltinProcedures(env *Env) {
	t := NewTokenizerFromString(builtinProcedures)
	tokens := t.Tokens()
	expressions, _ := Parse(&tokens)
	EvalAll(expressions, env)
}