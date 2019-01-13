package goscheme

import (
	"errors"
	"fmt"
	"os"
	"path"
	"regexp"
	"strconv"
)

func Eval(exp Expression, env *Env) (ret Expression, err error) {
	for {
		if isNullExp(exp) {
			return NilObj, nil
		}
		if isUndefObj(exp) {
			return undefObj, nil
		}
		if IsNumber(exp) {
			ret, err = expressionToNumber(exp)
			return
		} else if IsBoolean(exp) {
			return IsTrue(exp), nil
		} else if IsString(exp) {
			return expToString(exp)
		} else if IsSymbol(exp) {
			s, _ := exp.(string)
			ret, err = env.Find(Symbol(s))
			return
		} else if IsSpecialSyntaxExpression(exp, "define") {
			operators, ok := exp.([]Expression)
			if !ok {
				err = fmt.Errorf("%v not a valid syntax expression", exp)
				return
			}
			ret, err = evalDefine(operators[1], operators[2:], env)
			return
		} else if IsSpecialSyntaxExpression(exp, "eval") {
			exps, ok := exp.([]Expression)
			if !ok {
				err = fmt.Errorf("%v not a valid syntax expression", exp)
				return
			}
			ret, err = evalEval(exps[1], env)
			return
		} else if IsSpecialSyntaxExpression(exp, "apply") {
			exps, ok := exp.([]Expression)
			if !ok {
				err = fmt.Errorf("%v not a valid syntax expression", exp)
				return
			}
			return evalApply(exps[1:], env)
		} else if IsSpecialSyntaxExpression(exp, "if") {
			e := exp.([]Expression)
			exp, err = evalIf(e, env)
			if err != nil {
				return
			}
		} else if IsSpecialSyntaxExpression(exp, "cond") {
			return evalCond(exp, env)
		} else if IsSpecialSyntaxExpression(exp, "begin") {
			e, ok := exp.([]Expression)
			if !ok {
				err = fmt.Errorf("%v not a valid syntax expression", exp)
				return
			}
			exp, err = evalBegin(e, env)
			if err != nil {
				return
			}
		} else if IsSpecialSyntaxExpression(exp, "lambda") {
			return evalLambda(exp, env)
		} else if IsSpecialSyntaxExpression(exp, "load") {
			exps, ok := exp.([]Expression)
			if !ok {
				err = fmt.Errorf("%v not a valid syntax expression", exp)
				return
			}
			return evalLoad(exps[1], env)
		} else if IsSpecialSyntaxExpression(exp, "delay") {
			return evalDelay(exp, env)
		} else if IsSpecialSyntaxExpression(exp, "and") {
			return evalAnd(exp, env)
		} else if IsSpecialSyntaxExpression(exp, "or") {
			return evalOr(exp, env)
		} else if IsSpecialSyntaxExpression(exp, "let") {
			return evalLet(exp, env)
		} else if IsSpecialSyntaxExpression(exp, "let*") {
			return evalL2RLet(exp, env)
		} else if IsSpecialSyntaxExpression(exp, "letrec") {
			return evalLetRec(exp, env)
		} else if IsSpecialSyntaxExpression(exp, "set!") {
			return evalSet(exp, env)
		} else {
			ops, ok := exp.([]Expression)
			if !ok {
				// exp is just a bottom builtin type, return it directly
				return exp, nil
			}
			if isQuoteExpression(exp) {
				return evalQuote(ops[1], env)
			}
			fn, err := Eval(ops[0], env)
			if err != nil {
				return fn, err
			}
			switch p := fn.(type) {
			case Function:
				var args []Expression
				for _, arg := range ops[1:] {
					v, err := Eval(arg, env)
					if err != nil {
						return undefObj, err
					}
					args = append(args, v)
				}
				return p.Call(args...)
			case *LambdaProcess:
				newEnv := &Env{outer: p.env, frame: make(map[Symbol]Expression)}
				if len(ops[1:]) != len(p.params) {
					return undefObj, errors.New(fmt.Sprintf("%v\n", p.String()) + "require " + strconv.Itoa(len(p.params)) + " but " + strconv.Itoa(len(ops[1:])) + " provide")
				}
				for i, arg := range ops[1:] {
					val, err := Eval(arg, env)
					if err != nil {
						return undefObj, err
					}
					newEnv.Set(p.params[i], val)
				}
				exp = p.Body()
				env = newEnv
			default:
				return undefObj, errors.New(fmt.Sprintf("%v is not callable", fn))
			}
		}
	}
}
func evalSet(exp Expression, env *Env) (Expression, error) {
	expressions, ok := exp.([]Expression)
	if !ok || len(expressions) != 3 {
		return undefObj, errors.New("set!: syntax error (set! requires variable and value arguments)")
	}
	sym, err := transExpressionToSymbol(expressions[1])
	if err != nil {
		return undefObj, err
	}
	val, err := Eval(expressions[2], env)
	currentEnv := env
	for currentEnv != nil {
		if _, ok := currentEnv.frame[sym]; ok {
			currentEnv.Set(sym, val)
			return undefObj, nil
		}
		currentEnv = env.outer
	}
	return undefObj, errors.New(fmt.Sprintf("variable %v cannot set! before define", sym))
}

func evalLetRec(exp Expression, env *Env) (Expression, error) {
	expressions, ok := exp.([]Expression)
	if !ok || len(expressions) < 3 {
		return undefObj, errors.New("letrec: syntax error (letrec should pass the variables and body)")
	}
	bindings, ok := expressions[1].([]Expression)
	if !ok {
		return undefObj, errors.New("letrec: syntax error (not a valid binding)")
	}
	newEnv := &Env{outer: env, frame: make(map[Symbol]Expression)}
	// init symbols with undef
	for _, exp := range bindings {
		binding, ok := exp.([]Expression)
		if !ok || len(binding) != 2 {
			return undefObj, errors.New("letrec: syntax error (not a valid binding)")
		}
		sym, err := transExpressionToSymbol(binding[0])
		if err != nil {
			return undefObj, err
		}
		newEnv.Set(sym, undefObj)
	}
	// set value for symbols
	for _, exp := range bindings {
		binding, _ := exp.([]Expression)
		sym, _ := transExpressionToSymbol(binding[0])
		val, err := Eval(binding[1], newEnv)
		if err != nil {
			return undefObj, err
		}
		newEnv.Set(sym, val)
	}
	var ret Expression
	var err error
	for _, exp := range expressions[2:] {
		ret, err = Eval(exp, newEnv)
		if err != nil {
			return ret, err
		}
	}
	return ret, nil
}

func evalL2RLet(exp Expression, env *Env) (Expression, error) {
	expressions, ok := exp.([]Expression)
	if !ok || len(expressions) < 3 {
		return undefObj, errors.New("let*: syntax error (let* should pass the variables and body)")
	}
	bindings, ok := expressions[1].([]Expression)
	if !ok {
		return undefObj, errors.New("let*: syntax error (not a valid binding)")
	}
	var outerEnv, currentEnv *Env
	outerEnv = env
	for _, exp := range bindings {
		currentEnv = &Env{outer: outerEnv, frame: make(map[Symbol]Expression)}
		binding, ok := exp.([]Expression)
		if !ok || len(binding) != 2 {
			return undefObj, errors.New("let*: syntax error (not a valid binding)")
		}
		sym, err := transExpressionToSymbol(binding[0])
		if err != nil {
			return undefObj, err
		}
		val, err := Eval(binding[1], currentEnv)
		if err != nil {
			return undefObj, nil
		}
		currentEnv.Set(sym, val)
		outerEnv = currentEnv
	}
	var ret Expression
	var err error
	for _, exp := range expressions[2:] {
		ret, err = Eval(exp, currentEnv)
		if err != nil {
			return ret, err
		}
	}
	return ret, nil
}

func evalLet(exp Expression, env *Env) (Expression, error) {
	expressions, ok := exp.([]Expression)
	if !ok || len(expressions) < 3 {
		return undefObj, errors.New("let: syntax error (let should pass the variables and body)")
	}
	bindings, ok := expressions[1].([]Expression)
	if !ok {
		return undefObj, errors.New("let: syntax error (not a valid binding)")
	}
	newEnv := &Env{outer: env, frame: make(map[Symbol]Expression)}
	for _, exp := range bindings {
		binding, ok := exp.([]Expression)
		if !ok || len(binding) != 2 {
			return undefObj, errors.New("let: syntax error (not a valid binding)")
		}
		sym, err := transExpressionToSymbol(binding[0])
		if err != nil {

		}
		val, err := Eval(binding[1], env)
		if err != nil {
			return undefObj, nil
		}
		newEnv.Set(sym, val)
	}
	var ret Expression
	var err error
	for _, exp := range expressions[2:] {
		ret, err = Eval(exp, newEnv)
		if err != nil {
			return ret, err
		}
	}
	return ret, nil
}

func evalAnd(exp Expression, env *Env) (Expression, error) {
	expressions, ok := exp.([]Expression)
	if !ok || len(expressions) < 2 {
		return undefObj, errors.New("and require at least 1 argument")
	}
	for _, e := range expressions[1:] {
		val, err := Eval(e, env)
		if err != nil {
			return undefObj, err
		}
		if !IsTrue(val) {
			return false, nil
		}
	}
	return true, nil
}

func evalOr(exp Expression, env *Env) (Expression, error) {
	expressions, ok := exp.([]Expression)
	if !ok || len(expressions) < 2 {
		return undefObj, errors.New("or require at least 1 argument")
	}
	for _, e := range expressions[1:] {
		result, err := Eval(e, env)
		if err != nil {
			return undefObj, err
		}
		if IsTrue(result) {
			return true, nil
		}
	}
	return false, nil
}

func evalDelay(exp Expression, env *Env) (Expression, error) {
	exps, ok := exp.([]Expression)
	if !ok || len(exps) < 2 {
		return undefObj, errors.New("delay require one argument")
	}
	return NewThunk(exps[1], env), nil
}

func isNullExp(exp Expression) bool {
	if exp == nil {
		return true
	}
	switch e := exp.(type) {
	case NilType:
		return true
	case *Pair:
		return e.IsNull()
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

func expToString(exp Expression) (String, error) {
	s, _ := exp.(string)
	pattern := regexp.MustCompile(`"((.|[\r\n])*?)"`)
	m := pattern.FindAllStringSubmatch(s, -1)
	if len(m) < 1 || len(m[0]) < 2 {
		return "", errors.New("not a string, format invalid.")
	}
	return String(m[0][1]), nil
}

func Apply(exp Expression) Expression {
	return nil
}

func evalEval(exp Expression, env *Env) (Expression, error) {
	arg, err := Eval(exp, env)
	if !validEvalExp(arg) {
		return undefObj, errors.New("error: malformed list")
	}
	expStr := valueToString(arg)
	t := NewTokenizerFromString(expStr)
	tokens := t.Tokens()
	ret, err := Parse(&tokens)
	if err != nil {
		return undefObj, err
	}
	return EvalAll(ret, env)
}

func validEvalExp(exp Expression) bool {
	switch p := exp.(type) {
	case *Pair:
		if !p.IsList() {
			return false
		}
		return validEvalExp(p.Car) && validEvalExp(p.Cdr)
	default:
		return true
	}
}

func evalApply(exp Expression, env *Env) (Expression, error) {
	args, ok := exp.([]Expression)
	if !ok || len(args) != 2 {
		return undefObj, errors.New("apply require 2 arguments")
	}
	procedure, err := Eval(args[0], env)
	if err != nil {
		return undefObj, nil
	}
	arg, err := Eval(args[1], env)
	if err != nil {
		return undefObj, nil
	}
	if !isList(arg) {
		return undefObj, errors.New("argument must be a list")
	}
	argList := arg.(*Pair)
	var argSlice = make([]Expression, 0, 1)
	argSlice = append(argSlice, extractList(argList)...)
	var expression []Expression
	expression = append(expression, procedure)
	expression = append(expression, argSlice...)
	return Eval(expression, env)
}

// load other scheme script file
func evalLoad(exp Expression, env *Env) (Expression, error) {
	argValue, err := Eval(exp, env)
	if err != nil {
		return undefObj, err
	}
	switch v := argValue.(type) {
	case String:
		if err := loadFile(string(v), env); err != nil {
			return undefObj, err
		}
	case Quote:
		if err := loadFile(string(v), env); err != nil {
			return undefObj, err
		}
	case *Pair:
		if isList(v) {
			expressions := extractList(v)
			for _, p := range expressions {
				ret, err := evalLoad(p, env)
				if err != nil {
					return ret, err
				}
			}
		}
	default:
		return undefObj, errors.New("argument can only contains string, quote or list")
	}
	return undefObj, nil
}

func loadFile(filePath string, env *Env) error {
	ext := path.Ext(filePath)
	if ext != ".scm" {
		filePath += ".scm"
	}
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("load %s failed: %s\n", filePath, err)
	}
	i := NewFileInterpreterWithEnv(f, env)
	return i.Run()
}

func evalQuote(exp Expression, env *Env) (Expression, error) {
	switch v := exp.(type) {
	case Number:
		return v, nil
	case string:
		if IsNumber(v) {
			return expressionToNumber(exp)
		}
		if IsString(exp) {
			return expToString(exp)
		}
		return Quote(v), nil
	case []Expression:
		var args []Expression
		for _, exp := range v {
			q, err := evalQuote(exp, env)
			if err != nil {
				return undefObj, err
			}
			args = append(args, q)
		}
		return listImpl(args...)
	default:
		return undefObj, errors.New("invalid quote argument")
	}
}

func evalLambda(exp Expression, env *Env) (*LambdaProcess, error) {
	se, ok := exp.([]Expression)
	if !ok || len(se) < 3 {
		return nil, errors.New("not a valid lambda expression")
	}
	paramOperand := se[1]
	body := se[2:]
	var paramNames []Symbol
	switch p := paramOperand.(type) {
	case []Expression:
		for _, e := range p {
			sym, err := transExpressionToSymbol(e)
			if err != nil {
				return nil, err
			}
			paramNames = append(paramNames, sym)
		}
	case Expression:
		sym, err := transExpressionToSymbol(p)
		if err != nil {
			return nil, err
		}
		paramNames = []Symbol{sym}
	}
	return makeLambdaProcess(paramNames, body, env), nil
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

func evalDefine(s Expression, val []Expression, env *Env) (Expression, error) {
	switch se := s.(type) {
	case []Expression:
		var symbols []Symbol
		for _, e := range se {
			sym, err := transExpressionToSymbol(e)
			if err != nil {
				return undefObj, err
			}
			symbols = append(symbols, sym)
		}
		p := makeLambdaProcess(symbols[1:], val, env)
		env.Set(Symbol(symbols[0]), p)
	case Expression:
		if len(val) != 1 {
			return undefObj, errors.New("define: bad syntax (multiple expressions after identifier")
		}
		sym, err := transExpressionToSymbol(se)
		if err != nil {
			return undefObj, err
		}
		val, err := Eval(val[0], env)
		if err != nil {
			return undefObj, err
		}
		env.Set(sym, val)
	}
	return undefObj, nil
}

func transExpressionToSymbol(s Expression) (Symbol, error) {
	if IsSymbol(s) {
		s, _ := s.(string)
		return Symbol(s), nil
	}
	return "", errors.New(fmt.Sprintf("%v is not a symbol", s))
}

func getParamSymbols(input []string) (ret []Symbol) {
	for _, s := range input {
		ret = append(ret, Symbol(s))
	}
	return
}

func makeLambdaProcess(paramNames []Symbol, body []Expression, env *Env) *LambdaProcess {
	return &LambdaProcess{paramNames, body, env}
}

func EvalAll(exps []Expression, env *Env) (ret Expression, err error) {
	for _, exp := range exps {
		ret, err = Eval(exp, env)
		if err != nil {
			return
		}
	}
	return
}

func expressionToNumber(exp Expression) (Number, error) {
	v := exp
	if !IsNumber(v) {
		return 0, errors.New(fmt.Sprintf("%v is not a number", v))
	}
	switch t := v.(type) {
	case Number:
		return t, nil
	case string:
		f, _ := strconv.ParseFloat(t, 64)
		return Number(f), nil
	}
	return 0, nil
}

func conditionOfIfExpression(exp []Expression) (Expression, error) {
	if len(exp) < 3 {
		return undefObj, errors.New("not a valid if expression")
	}
	return exp[1], nil
}

func trueExpOfIfExpression(exp []Expression) (Expression, error) {
	if len(exp) < 3 {
		return undefObj, errors.New("not a valid if expression")
	}
	return exp[2], nil
}

func elseExpOfIfExpression(exp []Expression) (Expression, error) {
	if len(exp) < 3 {
		return undefObj, errors.New("not a valid if expression")
	}
	if len(exp) < 4 {
		return undefObj, nil
	}
	return exp[3], nil
}

func evalIf(exp []Expression, env *Env) (Expression, error) {
	conditionExp, err := conditionOfIfExpression(exp)
	if err != nil {
		return undefObj, err
	}
	condition, err := Eval(conditionExp, env)
	if err != nil {
		return undefObj, err
	}
	if IsTrue(condition) {
		return trueExpOfIfExpression(exp)
	} else {
		return elseExpOfIfExpression(exp)
	}
}

func evalBegin(exp []Expression, env *Env) (Expression, error) {
	if len(exp) < 2 {
		return undefObj, errors.New("not a valid begin expression")
	}
	for _, e := range exp[1 : len(exp)-1] {
		Eval(e, env)
	}
	return exp[len(exp)-1], nil
}

func evalCond(exp Expression, env *Env) (Expression, error) {
	equalIfExp, err := expandCond(exp)
	if err != nil {
		return undefObj, err
	}
	return Eval(equalIfExp, env)
}

func makeIf(condition, trueExp, elseExp Expression) []Expression {
	return []Expression{"if", condition, trueExp, elseExp}
}

func condClauses(exp []Expression) []Expression {
	return exp[1:]
}

func expandCond(exp Expression) (Expression, error) {
	e, ok := exp.([]Expression)
	if !ok {
		return undefObj, fmt.Errorf("%v not a valid expression", exp)
	}
	return condClausesToIf(condClauses(e))
}

func conditionOfClause(exp []Expression) (Expression, error) {
	if len(exp) == 0 {
		return undefObj, fmt.Errorf("cannot find clause of %v", exp)
	}
	return exp[0], nil
}

func processesOfClause(exp []Expression) (Expression, error) {
	if len(exp) < 2 {
		return undefObj, errors.New("clause of expression not found")
	}
	return exp[1:], nil
}

func isElseClause(clause Expression) bool {
	switch v := clause.(type) {
	case []Expression:
		return v[0] == "else"
	default:
		return false
	}
}

func condClausesToIf(exp []Expression) (Expression, error) {
	if isNullExp(exp) {
		// just a nil obj
		return undefObj, nil
	}
	first, ok := exp[0].([]Expression)
	if !ok {
		return undefObj, fmt.Errorf("%v not a valid expression", exp[0])
	}
	rest := exp[1:]
	if isElseClause(first) {
		if len(rest) != 0 {
			return undefObj, errors.New("else clause must in the last position: cond->if")
		}
		clause, err := processesOfClause(first)
		if err != nil {
			return undefObj, err
		}
		return sequenceToExp(clause), nil
	} else {
		var condition, seq, clause Expression
		condition, err := conditionOfClause(first)
		if err != nil {
			return undefObj, err
		}
		clause, err = processesOfClause(first)
		if err != nil {
			return undefObj, err
		}
		seq = sequenceToExp(clause)
		elseIfClause, err := condClausesToIf(rest)
		if err != nil {
			return undefObj, err
		}
		return makeIf(condition, seq, elseIfClause), nil
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
