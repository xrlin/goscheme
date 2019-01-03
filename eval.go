package goscheme

import (
	"fmt"
	"regexp"
	"strconv"
)

func Eval(exp Expression, env *Env) (ret Expression) {
	for {
		if isNullExp(exp) {
			return NilObj
		}
		if IsNumber(exp) {
			ret = expressionToNumber(exp)
			return
		} else if IsBoolean(exp) {
			return IsTrue(exp)
		} else if IsString(exp) {
			return expToString(exp)
		} else if IsSymbol(exp) {
			var err error
			s, _ := exp.(string)
			ret, err = env.Find(Symbol(s))
			if err != nil {
				panic(err)
			}
			return
		} else if IsSpecialSyntaxExpression(exp, "define") {
			operators, _ := exp.([]Expression)
			if len(operators) != 3 {
				panic("define require 3 arguments")
			}
			ret = evalDefine(operators[1], operators[2], env)
			return
		} else if IsSpecialSyntaxExpression(exp, "if") {
			e := exp.([]Expression)
			exp = evalIf(e, env)
		} else if IsSpecialSyntaxExpression(exp, "cond") {
			return evalCond(exp, env)
		} else if IsSpecialSyntaxExpression(exp, "begin") {
			e := exp.([]Expression)
			exp = evalBegin(e, env)
		} else if IsSpecialSyntaxExpression(exp, "lambda") {
			return evalLambda(exp, env)
		} else {
			ops := exp.([]Expression)
			if isQuoteExpression(exp) {
				return evalQuote(ops[1], env)
			}
			fn := Eval(ops[0], env)
			switch p := fn.(type) {
			case Function:
				var args []Expression
				for _, arg := range ops[1:] {
					args = append(args, Eval(arg, env))
				}
				return p(args...)
			case *LambdaProcess:
				newEnv := &Env{outer: p.env, frame: make(map[Symbol]Expression)}
				if len(ops[1:]) != len(p.params) {
					panic("require " + strconv.Itoa(len(p.params)) + " but " + strconv.Itoa(len(ops[1:])) + " provide")
				}
				for i, arg := range ops[1:] {
					newEnv.Set(p.params[i], Eval(arg, env))
				}
				exp = p.body
				env = newEnv
			default:
				panic(fmt.Sprintf("%v is not callable", fn))
			}
		}
	}
}

func isNullExp(exp Expression) bool {
	if exp == nil {
		return true
	}
	switch e := exp.(type) {
	case NilType:
		return true
	case Undef:
		return true
	case []Expression:
		if len(e) == 0 {
			return true
		}
		return false
	default:
		return false
	}
}

func isLambdaType(expression Expression) bool {
	_, ok := expression.(*LambdaProcess)
	return ok
}

func expToString(exp Expression) string {
	s, _ := exp.(string)
	pattern := regexp.MustCompile(`"((.|[\r\n])*?)"`)
	m := pattern.FindAllStringSubmatch(s, -1)
	return m[0][1]
}

func Apply(exp Expression) Expression {
	return nil
}

func evalQuote(exp Expression, env *Env) Expression {
	switch v := exp.(type) {
	case Number:
		return v
	case string:
		if IsNumber(v) {
			return expressionToNumber(exp)
		}
		if IsString(exp) {
			return expToString(exp)
		}
		return Quote(v)
	case []Expression:
		var args []Expression
		for _, exp := range v {
			args = append(args, evalQuote(exp, env))
		}
		return listImpl(args...)
	default:
		panic("invalid quote argument")
	}
}

func evalLambda(exp Expression, env *Env) *LambdaProcess {
	se, _ := exp.([]Expression)
	paramOperand := se[1]
	body := se[2]
	var paramNames []Symbol
	switch p := paramOperand.(type) {
	case []Expression:
		for _, e := range p {
			paramNames = append(paramNames, transExpressionToSymbol(e))
		}
	case Expression:
		paramNames = []Symbol{transExpressionToSymbol(p)}
	}
	return makeLambdaProcess(paramNames, body, env)
}

func isQuoteExpression(exp Expression) bool {
	if exp == "quote" {
		return true
	}
	ops, ok := exp.([]Expression)
	if !ok {
		return false
	}
	return ops[0] == "quote"
}

func evalDefine(s Expression, val Expression, env *Env) Expression {
	switch se := s.(type) {
	case []Expression:
		var symbols []Symbol
		for _, e := range se {
			symbols = append(symbols, transExpressionToSymbol(e))
		}
		p := makeLambdaProcess(symbols[1:], val, env)
		env.Set(Symbol(symbols[0]), p)
	case Expression:
		env.Set(transExpressionToSymbol(se), Eval(val, env))
	}
	return undefObj
}

func transExpressionToSymbol(s Expression) Symbol {
	if IsSymbol(s) {
		s, _ := s.(string)
		return Symbol(s)
	}
	panic(fmt.Sprintf("%v is not a symbol", s))
}

func getParamSymbols(input []string) (ret []Symbol) {
	for _, s := range input {
		ret = append(ret, Symbol(s))
	}
	return
}

func makeLambdaProcess(paramNames []Symbol, body Expression, env *Env) *LambdaProcess {
	return &LambdaProcess{paramNames, body, env}
}

func EvalAll(exps []Expression, env *Env) (ret Expression) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
		}
	}()
	for _, exp := range exps {
		ret = Eval(exp, env)
	}
	return
}

func expressionToNumber(exp Expression) Number {
	if !IsNumber(exp) {
		panic(fmt.Sprintf("%v is not a number", exp))
	}
	switch t := exp.(type) {
	case Number:
		return t
	case string:
		f, _ := strconv.ParseFloat(t, 64)
		return Number(f)
	}
	return 0
}

func conditionOfIfExpression(exp []Expression) Expression {
	return exp[1]
}

func trueExpOfIfExpression(exp []Expression) Expression {
	return exp[2]
}

func elseExpOfIfExpression(exp []Expression) Expression {
	if len(exp) < 3 {
		return undefObj
	}
	return exp[3]
}

func evalIf(exp []Expression, env *Env) Expression {
	if IsTrue(Eval(conditionOfIfExpression(exp), env)) {
		return trueExpOfIfExpression(exp)
	} else {
		return elseExpOfIfExpression(exp)
	}
}

func evalBegin(exp []Expression, env *Env) Expression {
	//var ret Expression
	for _, e := range exp[1 : len(exp)-1] {
		Eval(e, env)
	}
	return exp[len(exp)-1]
	//return ret
}

func evalCond(exp Expression, env *Env) Expression {
	equalIfExp := expandCond(exp)
	return Eval(equalIfExp, env)
}

func makeIf(condition, trueExp, elseExp Expression) []Expression {
	return []Expression{"if", condition, trueExp, elseExp}
}

func condClauses(exp []Expression) []Expression {
	return exp[1:]
}

func expandCond(exp Expression) Expression {
	e := exp.([]Expression)
	return condClausesToIf(condClauses(e))
}

func conditionOfClause(exp []Expression) Expression {
	return exp[0]
}

func processesOfClause(exp []Expression) Expression {
	return exp[1:]
}

func isElseClause(clause Expression) bool {
	switch v := clause.(type) {
	case []Expression:
		return v[0] == "else"
	default:
		return false
	}
}

func condClausesToIf(exp []Expression) Expression {
	if isNullExp(exp) {
		// just a nil obj
		return undefObj
	}
	first := exp[0].([]Expression)
	rest := exp[1:]
	if isElseClause(first) {
		if len(rest) != 0 {
			panic("else clause must in the last position: cond->if")
		}
		return sequenceToExp(processesOfClause(first))
	} else {
		return makeIf(conditionOfClause(first), sequenceToExp(processesOfClause(first)), condClausesToIf(rest))
	}

}

func sequenceToExp(exp Expression) Expression {
	switch exs := exp.(type) {
	case []Expression:
		ret := []Expression{"begin"}
		ret = append(ret, exs...)
		return ret
	case Expression:
		return exs
	}
	return undefObj
}
