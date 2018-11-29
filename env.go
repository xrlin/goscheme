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
