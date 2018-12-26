package goscheme

import (
	"fmt"
	"go/types"
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
	"<=": Function(func(args ...Expression) Expression {
		op1 := expressionToNumber(args[0])
		op2 := expressionToNumber(args[1])
		if op1 <= op2 {
			return true
		}
		return false
	}),
	"display": Function(func(args ...Expression) Expression {
		fmt.Printf("%v", args[0])
		return undefObj
	}),
	"cons":   Function(consImpl),
	"car":    Function(carImpl),
	"cdr":    Function(cdrImpl),
	"list":   Function(listImpl),
	"append": Function(appendImpl),
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
	return builtinEnv
}
