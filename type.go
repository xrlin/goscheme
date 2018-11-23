package goscheme

import (
	"fmt"
	"strings"
)

type Expression interface{}

type Number float64

type String string

type Symbol string

type Boolean bool

type NilType struct{}

func (n NilType) String() string {
	return "()"
}

var NilObj = NilType{}

func IsNilObj(obj Expression) bool {
	switch obj.(type) {
	case NilType:
		return true
	default:
		return false
	}
}

func IsPair(obj Expression) bool {
	switch obj.(type) {
	case *Pair:
		return true
	default:
		return false
	}
}

// Should only use with pointer
type Pair struct {
	Car, Cdr Expression
}

func (p *Pair) IsNull() bool {
	return p.Car == nil && p.Cdr == nil
}

func (p *Pair) IsList() bool {
	currentPair := p
	for {
		if currentPair.IsNull() {
			return true
		}
		switch cdr := currentPair.Cdr.(type) {
		case *Pair:
			currentPair = cdr
		case NilType:
			return true
		default:
			return false
		}
	}
}

func (p *Pair) String() string {

	currentPair := p

	var strSlices []string

	for !currentPair.IsNull() {
		if IsPair(currentPair.Car) {
			strSlices = append(strSlices, currentPair.Car.(*Pair).String())
		} else {
			strSlices = append(strSlices, fmt.Sprintf("%v", currentPair.Car))
		}

		if IsPair(currentPair.Cdr) {
			currentPair = currentPair.Cdr.(*Pair)
		} else {
			if IsNilObj(currentPair.Cdr) {
				break
			}
			strSlices = append(strSlices, ".")
			strSlices = append(strSlices, fmt.Sprintf("%v", currentPair.Cdr))
			break
		}
	}

	return "(" + strings.Join(strSlices, " ") + ")"
}
