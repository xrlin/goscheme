package goscheme

import (
	"fmt"
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
