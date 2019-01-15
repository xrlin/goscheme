package goscheme

import (
	"errors"
	"fmt"
	"os"
	"path"
	"regexp"
	"strconv"
)

// Eval is the main function to evaluate the expression in an environment.
func Eval(exp Expression, env *Env) (ret Expression, err error) {
	for {
		if IsPrimitiveExpression(exp) {
			return evalPrimitive(exp)
		}
		if IsSymbol(exp) {
			s, _ := exp.(string)
			ret, err = env.Find(Symbol(s))
			return
		}
		if IsSyntaxExpression(exp) {
			syntaxName, args, err := retrieveSyntaxAndArgs(exp)
			if err != nil {
				return UndefObj, err
			}
			syntax, _ := SyntaxMap[syntaxName]
			exp, err = applySyntaxExpression(syntax, args, env)
			if err != nil {
				return UndefObj, err
			}
		} else {
			ops, ok := exp.([]Expression)
			if !ok {
				return UndefObj, fmt.Errorf("%s is not a valid expression", exp)
			}
			nextExp, newEnv, err := applyCallable(ops[0], ops[1:], env)
			if err != nil {
				return UndefObj, err
			}
			exp = nextExp
			env = newEnv
		}
	}
}

// for tail recursion optimization, return the next expression will be executed and the new environment to execute the
// next loop in eval
func applyCallable(process Expression, argExpressions []Expression, env *Env) (Expression, *Env, error) {
	fn, err := Eval(process, env)
	if err != nil {
		return fn, env, err
	}
	switch p := fn.(type) {
	case Function:
		var args []Expression
		for _, arg := range argExpressions {
			v, err := Eval(arg, env)
			if err != nil {
				return UndefObj, env, err
			}
			args = append(args, v)
		}
		ret, err := p.Call(args...)
		return ret, env, err
	case *LambdaProcess:
		newEnv := &Env{outer: p.env, frame: make(map[Symbol]Expression)}
		if len(argExpressions) != len(p.params) {
			return UndefObj, env, errors.New(fmt.Sprintf("%v\n", p.String()) + "require " + strconv.Itoa(len(p.params)) + " but " + strconv.Itoa(len(argExpressions)) + " provide")
		}
		for i, arg := range argExpressions {
			val, err := Eval(arg, env)
			if err != nil {
				return UndefObj, env, err
			}
			newEnv.Set(p.params[i], val)
		}
		return p.Body(), newEnv, nil
	default:
		return UndefObj, env, fmt.Errorf("%v is not callable", fn)
	}
}

func applySyntaxExpression(syntax *Syntax, args []Expression, env *Env) (Expression, error) {
	return syntax.Eval(args, env)
}

func retrieveSyntaxAndArgs(exp Expression) (syntaxName string, args []Expression, err error) {
	pieces, ok := exp.([]Expression)
	if !ok || len(pieces) < 1 {
		err = fmt.Errorf("%s is not a valid syntax expression", exp)
		return
	}
	syntaxName, _ = pieces[0].(string)
	args = pieces[1:]
	return
}
func evalPrimitive(exp Expression) (Expression, error) {
	if IsNullExp(exp) {
		return NilObj, nil
	}
	if IsUndefObj(exp) {
		return UndefObj, nil
	}
	if IsNumber(exp) {
		return expressionToNumber(exp)
	}
	if IsBoolean(exp) {
		return IsTrue(exp), nil
	}
	if IsString(exp) {
		return expToString(exp)
	}
	return exp, nil
}

func evalSet(args []Expression, env *Env) (Expression, error) {
	if len(args) != 2 {
		return UndefObj, errors.New("set!: syntax error (set! requires variable and value arguments)")
	}
	sym, err := transExpressionToSymbol(args[0])
	if err != nil {
		return UndefObj, err
	}
	val, err := Eval(args[1], env)
	currentEnv := env
	for currentEnv != nil {
		if _, ok := currentEnv.frame[sym]; ok {
			currentEnv.Set(sym, val)
			return UndefObj, nil
		}
		currentEnv = env.outer
	}
	return UndefObj, fmt.Errorf("variable %v cannot set! before define", sym)
}

func evalLetRec(args []Expression, env *Env) (Expression, error) {
	if len(args) < 2 {
		return UndefObj, errors.New("letrec: syntax error (letrec should pass the variables and body)")
	}
	bindings, ok := args[0].([]Expression)
	if !ok {
		return UndefObj, errors.New("letrec: syntax error (not a valid binding)")
	}
	newEnv := &Env{outer: env, frame: make(map[Symbol]Expression)}
	// init symbols with undef
	for _, exp := range bindings {
		binding, ok := exp.([]Expression)
		if !ok || len(binding) != 2 {
			return UndefObj, errors.New("letrec: syntax error (not a valid binding)")
		}
		sym, err := transExpressionToSymbol(binding[0])
		if err != nil {
			return UndefObj, err
		}
		newEnv.Set(sym, UndefObj)
	}
	// set value for symbols
	for _, exp := range bindings {
		binding, _ := exp.([]Expression)
		sym, _ := transExpressionToSymbol(binding[0])
		val, err := Eval(binding[1], newEnv)
		if err != nil {
			return UndefObj, err
		}
		newEnv.Set(sym, val)
	}
	var ret Expression
	var err error
	for _, exp := range args[1:] {
		ret, err = Eval(exp, newEnv)
		if err != nil {
			return ret, err
		}
	}
	return ret, nil
}

func evalL2RLet(args []Expression, env *Env) (Expression, error) {
	if len(args) < 2 {
		return UndefObj, errors.New("let*: syntax error (let* should pass the variables and body)")
	}
	bindings, ok := args[0].([]Expression)
	if !ok {
		return UndefObj, errors.New("let*: syntax error (not a valid binding)")
	}
	var outerEnv, currentEnv *Env
	outerEnv = env
	for _, exp := range bindings {
		currentEnv = &Env{outer: outerEnv, frame: make(map[Symbol]Expression)}
		binding, ok := exp.([]Expression)
		if !ok || len(binding) != 2 {
			return UndefObj, errors.New("let*: syntax error (not a valid binding)")
		}
		sym, err := transExpressionToSymbol(binding[0])
		if err != nil {
			return UndefObj, err
		}
		val, err := Eval(binding[1], currentEnv)
		if err != nil {
			return UndefObj, nil
		}
		currentEnv.Set(sym, val)
		outerEnv = currentEnv
	}
	var ret Expression
	var err error
	for _, exp := range args[1:] {
		ret, err = Eval(exp, currentEnv)
		if err != nil {
			return ret, err
		}
	}
	return ret, nil
}

func evalLet(args []Expression, env *Env) (Expression, error) {
	if len(args) < 2 {
		return UndefObj, errors.New("let: syntax error (let should pass the variables and body)")
	}
	bindings, ok := args[0].([]Expression)
	if !ok {
		return UndefObj, errors.New("let: syntax error (not a valid binding)")
	}
	newEnv := &Env{outer: env, frame: make(map[Symbol]Expression)}
	for _, exp := range bindings {
		binding, ok := exp.([]Expression)
		if !ok || len(binding) != 2 {
			return UndefObj, errors.New("let: syntax error (not a valid binding)")
		}
		sym, err := transExpressionToSymbol(binding[0])
		if err != nil {
			return UndefObj, err
		}
		val, err := Eval(binding[1], env)
		if err != nil {
			return UndefObj, nil
		}
		newEnv.Set(sym, val)
	}
	var ret Expression
	var err error
	for _, exp := range args[1:] {
		ret, err = Eval(exp, newEnv)
		if err != nil {
			return ret, err
		}
	}
	return ret, nil
}

func evalAnd(args []Expression, env *Env) (Expression, error) {
	if len(args) < 1 {
		return UndefObj, errors.New("and require at least 1 argument")
	}
	for _, e := range args {
		val, err := Eval(e, env)
		if err != nil {
			return UndefObj, err
		}
		if !IsTrue(val) {
			return false, nil
		}
	}
	return true, nil
}

func evalOr(args []Expression, env *Env) (Expression, error) {
	if len(args) < 1 {
		return UndefObj, errors.New("or require at least 1 argument")
	}
	for _, e := range args {
		result, err := Eval(e, env)
		if err != nil {
			return UndefObj, err
		}
		if IsTrue(result) {
			return true, nil
		}
	}
	return false, nil
}

func evalDelay(args []Expression, env *Env) (Expression, error) {
	if len(args) == 0 {
		return UndefObj, errors.New("delay require 1 argument")
	}
	return NewThunk(args[0], env), nil
}

func expToString(exp Expression) (String, error) {
	switch s := exp.(type) {
	case string:
		pattern := regexp.MustCompile(`"((.|[\r\n])*?)"`)
		m := pattern.FindAllStringSubmatch(s, -1)
		if len(m) < 1 || len(m[0]) < 2 {
			return "", errors.New("not a string, format invalid")
		}
		return String(m[0][1]), nil
	case String:
		return s, nil
	default:
		return "", errors.New("not as string")
	}
}

// evalEval eval the scheme object and calculate its value
func evalEval(args []Expression, env *Env) (Expression, error) {
	if len(args) != 1 {
		return UndefObj, errors.New("syntax error (requires 1 argument)")
	}
	expression := args[0]
	arg, err := Eval(expression, env)
	if !validEvalExp(arg) {
		return UndefObj, errors.New("error: malformed list")
	}
	expStr := valueToString(arg)
	t := NewTokenizerFromString(expStr)
	tokens := t.Tokens()
	ret, err := Parse(&tokens)
	if err != nil {
		return UndefObj, err
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

func evalApply(args []Expression, env *Env) (Expression, error) {
	if len(args) != 2 {
		return UndefObj, errors.New("syntax error (requires 2 argument)")
	}
	procedure, err := Eval(args[0], env)
	if err != nil {
		return UndefObj, nil
	}
	arg, err := Eval(args[1], env)
	if err != nil {
		return UndefObj, nil
	}
	if !isList(arg) {
		return UndefObj, errors.New("argument must be a list")
	}
	argList := arg.(*Pair)
	var argSlice = make([]Expression, 0, 1)
	argSlice = append(argSlice, extractList(argList)...)
	var expression []Expression
	expression = append(expression, procedure)
	expression = append(expression, argSlice...)
	return Eval(expression, env)
}

// load other scheme script files
func evalLoad(expression []Expression, env *Env) (Expression, error) {
	if len(expression) != 1 {
		return UndefObj, errors.New("syntax error (requires 1 argument)")
	}
	argValue, err := Eval(expression[0], env)
	if err != nil {
		return UndefObj, err
	}
	switch v := argValue.(type) {
	case String:
		if err := loadFile(string(v), env); err != nil {
			return UndefObj, err
		}
	case Quote:
		if err := loadFile(string(v), env); err != nil {
			return UndefObj, err
		}
	case *Pair:
		if isList(v) {
			expressions := extractList(v)
			for _, p := range expressions {
				ret, err := evalLoad([]Expression{p}, env)
				if err != nil {
					return ret, err
				}
			}
		}
	default:
		return UndefObj, errors.New("argument can only contains string, quote or list")
	}
	return UndefObj, nil
}

func loadFile(filePath string, env *Env) error {
	ext := path.Ext(filePath)
	if ext != ".scm" {
		filePath += ".scm"
	}
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("load %s failed: %s", filePath, err)
	}
	i := NewFileInterpreterWithEnv(f, env)
	return i.Run()
}

func evalQuote(args []Expression, env *Env) (Expression, error) {
	if len(args) != 1 {
		return UndefObj, errors.New("syntax error (requires 1 argument)")
	}
	exp := args[0]
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
			q, err := evalQuote([]Expression{exp}, env)
			if err != nil {
				return UndefObj, err
			}
			args = append(args, q)
		}
		return listImpl(args...)
	default:
		return UndefObj, errors.New("invalid quote argument")
	}
}

func evalLambda(args []Expression, env *Env) (Expression, error) {
	if len(args) < 2 {
		return nil, errors.New("not a valid lambda expression")
	}
	paramOperand := args[0]
	body := args[1:]
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

func evalDefine(args []Expression, env *Env) (Expression, error) {
	if len(args) < 2 {
		return UndefObj, errors.New("syntax error, require more than two arguments")
	}
	// fetch the symbol/argument names and value/body
	s, val := args[0], args[1:]
	switch se := s.(type) {
	case []Expression:
		var symbols []Symbol
		for _, e := range se {
			sym, err := transExpressionToSymbol(e)
			if err != nil {
				return UndefObj, err
			}
			symbols = append(symbols, sym)
		}
		p := makeLambdaProcess(symbols[1:], val, env)
		env.Set(Symbol(symbols[0]), p)
	case Expression:
		if len(val) != 1 {
			return UndefObj, errors.New("define: bad syntax (multiple expressions after identifier)")
		}
		sym, err := transExpressionToSymbol(se)
		if err != nil {
			return UndefObj, err
		}
		val, err := Eval(val[0], env)
		if err != nil {
			return UndefObj, err
		}
		env.Set(sym, val)
	}
	return UndefObj, nil
}

func transExpressionToSymbol(s Expression) (Symbol, error) {
	if IsSymbol(s) {
		s, _ := s.(string)
		return Symbol(s), nil
	}
	return "", fmt.Errorf("%v is not a symbol", s)
}

func makeLambdaProcess(paramNames []Symbol, body []Expression, env *Env) *LambdaProcess {
	return &LambdaProcess{paramNames, body, env}
}

// EvalAll iterate the sequence of expressions and evaluate each one.
// Returns the last evaluated value as the result
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
		return 0, fmt.Errorf("%v is not a number", v)
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
	if len(exp) < 2 {
		return UndefObj, errors.New("not a valid if expression")
	}
	return exp[0], nil
}

func trueExpOfIfExpression(exp []Expression) (Expression, error) {
	if len(exp) < 2 {
		return UndefObj, errors.New("not a valid if expression")
	}
	return exp[1], nil
}

func elseExpOfIfExpression(exp []Expression) (Expression, error) {
	if len(exp) < 2 {
		return UndefObj, errors.New("not a valid if expression")
	}
	if len(exp) < 3 {
		return UndefObj, nil
	}
	return exp[2], nil
}

func evalIf(args []Expression, env *Env) (Expression, error) {
	if len(args) < 2 {
		return UndefObj, errors.New("syntax error (requires 2 argument)")
	}
	conditionExp, err := conditionOfIfExpression(args)
	if err != nil {
		return UndefObj, err
	}
	condition, err := Eval(conditionExp, env)
	if err != nil {
		return UndefObj, err
	}
	if IsTrue(condition) {
		return trueExpOfIfExpression(args)
	}
	return elseExpOfIfExpression(args)
}

func evalBegin(args []Expression, env *Env) (Expression, error) {
	if len(args) < 1 {
		return UndefObj, errors.New("syntax error (requires more than 1 arguments)")
	}
	for _, e := range args[:len(args)-1] {
		Eval(e, env)
	}
	return args[len(args)-1], nil
}

func evalCond(exp []Expression, env *Env) (Expression, error) {
	equalIfExp, err := expandCond(exp)
	if err != nil {
		return UndefObj, err
	}
	return Eval(equalIfExp, env)
}

func makeIf(condition, trueExp, elseExp Expression) []Expression {
	return []Expression{"if", condition, trueExp, elseExp}
}

func condClauses(exp []Expression) []Expression {
	return exp[:]
}

func expandCond(exp Expression) (Expression, error) {
	e, ok := exp.([]Expression)
	if !ok {
		return UndefObj, fmt.Errorf("%v not a valid expression", exp)
	}
	return condClausesToIf(condClauses(e))
}

func conditionOfClause(exp []Expression) (Expression, error) {
	if len(exp) == 0 {
		return UndefObj, fmt.Errorf("cannot find clause of %v", exp)
	}
	return exp[0], nil
}

func processesOfClause(exp []Expression) (Expression, error) {
	if len(exp) < 2 {
		return UndefObj, errors.New("clause of expression not found")
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
	if IsNullExp(exp) {
		// just a nil obj
		return UndefObj, nil
	}
	first, ok := exp[0].([]Expression)
	if !ok {
		return UndefObj, fmt.Errorf("%v not a valid expression", exp[0])
	}
	rest := exp[1:]
	if isElseClause(first) {
		if len(rest) != 0 {
			return UndefObj, errors.New("else clause must in the last position: cond->if")
		}
		clause, err := processesOfClause(first)
		if err != nil {
			return UndefObj, err
		}
		return sequenceToExp(clause), nil
	}
	var condition, seq, clause Expression
	condition, err := conditionOfClause(first)
	if err != nil {
		return UndefObj, err
	}
	clause, err = processesOfClause(first)
	if err != nil {
		return UndefObj, err
	}
	seq = sequenceToExp(clause)
	elseIfClause, err := condClausesToIf(rest)
	if err != nil {
		return UndefObj, err
	}
	return makeIf(condition, seq, elseIfClause), nil

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
	return UndefObj
}
