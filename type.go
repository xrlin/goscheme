package goscheme

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Expression represent the parsed tokens of scheme syntax tree or the low level builtin types.
type Expression interface{}

// Number in scheme.
type Number float64

// Represents string in scheme.
type String string

// String return the string to display wrapping the low level string with quotes.
func (s String) String() string {
	return "\"" + string(s) + "\""
}

// Contains all defined scheme syntax.
var SyntaxMap = make(map[string]*Syntax)

// The actual func to eval and translate the syntax.
type SyntaxFunc func(args []Expression, env *Env) (Expression, error)

// Syntax wrap a syntax and give method to eval it.
type Syntax struct {
	fn   SyntaxFunc
	name string
}

// String return the string to display representing the syntax.
func (s *Syntax) String() string {
	return fmt.Sprintf("#[Syntax %s]", s.name)
}

// Eval runs the syntax and return the result.
func (s *Syntax) Eval(args []Expression, env *Env) (Expression, error) {
	return s.fn(args, env)
}

// NewSyntax construct a Syntax with custom name and SyntaxFunc
func NewSyntax(name string, fn SyntaxFunc) *Syntax {
	return &Syntax{fn, name}
}

func initSyntax() {
	SyntaxMap["define"] = NewSyntax("define", evalDefine)
	SyntaxMap["eval"] = NewSyntax("eval", evalEval)
	SyntaxMap["apply"] = NewSyntax("apply", evalApply)
	SyntaxMap["if"] = NewSyntax("if", evalIf)
	SyntaxMap["cond"] = NewSyntax("cond", evalCond)
	SyntaxMap["begin"] = NewSyntax("begin", evalBegin)
	SyntaxMap["lambda"] = NewSyntax("lambda", evalLambda)
	SyntaxMap["load"] = NewSyntax("load", evalLoad)
	SyntaxMap["delay"] = NewSyntax("delay", evalDelay)
	SyntaxMap["and"] = NewSyntax("and", evalAnd)
	SyntaxMap["or"] = NewSyntax("and", evalOr)
	SyntaxMap["let"] = NewSyntax("let", evalLet)
	SyntaxMap["let*"] = NewSyntax("let*", evalL2RLet)
	SyntaxMap["letrec"] = NewSyntax("letrec", evalLetRec)
	SyntaxMap["quote"] = NewSyntax("quote", evalQuote)
	SyntaxMap["set!"] = NewSyntax("set!", evalSet)
}

// Symbol represents the variable name in scheme.
type Symbol string

// Quote type in scheme
type Quote string

type commonFunction func(args ...Expression) (Expression, error)

// Builtin basic scheme function in pure go.
type Function struct {
	name     string
	function commonFunction
	minArgs  int
	maxArgs  int
}

// Call eval the function with args and returns the result.
func (f Function) Call(args ...Expression) (Expression, error) {
	if err := f.validateArgCount(args...); err != nil {
		return UndefObj, err
	}
	return f.function(args...)
}

func (f Function) validateArgCount(args ...Expression) error {
	if f.minArgs == -1 && f.maxArgs == -1 {
		return nil
	}
	c := len(args)
	if f.minArgs == f.maxArgs && f.maxArgs != c {
		return fmt.Errorf("%s requires %d arguments but %d arguments provided", f.name, f.maxArgs, c)
	}
	if f.minArgs != -1 && f.minArgs > c {
		return fmt.Errorf("%s requires at least %d arguments but %d arguments provided", f.name, f.minArgs, c)
	}
	if f.maxArgs != -1 && f.maxArgs < c {
		return fmt.Errorf("%s requires no more than %d arguments, but %d arguments provided", f.name, f.maxArgs, c)
	}
	return nil
}

// String returns the message to display
func (f Function) String() string {
	return "#[BuiltinFunction]"
}

// NewFunc return a Function struct init with arguments.
// minArgs, maxArgs define the arguments count limitation of Function. Set to -1 means no limitation.
func NewFunction(funcName string, f commonFunction, minArgs int, maxArgs int) Function {
	return Function{
		name:     funcName,
		function: f,
		minArgs:  minArgs,
		maxArgs:  maxArgs,
	}
}

// Thunk wraps expression for lazy execution
// Thunk should use as pointer
type Thunk struct {
	// expression to execute
	Exp Expression
	// expression result cache
	ret Expression
	// context to execute Exp
	Env *Env
}

// String returns the string represents the Thunk struct.
func (t Thunk) String() string {
	if t.ret != nil {
		return fmt.Sprintf("#[Thunk %s]", t.ret)
	}
	return fmt.Sprintf("#[Thunk exp: %s]", t.Exp)
}

// Value returns the actual value of the thunk
func (t *Thunk) Value() (Expression, error) {
	if t.ret != nil {
		return t.ret, nil
	}
	value, err := Eval(t.Exp, t.Env)
	if err != nil {
		return UndefObj, err
	}
	switch t2 := value.(type) {
	case *Thunk:
		value, err = t2.Value()
	default:
	}
	if err != nil {
		return UndefObj, err
	}
	t.ret = value
	return t.ret, nil
}

// IsThunk checks whether an expression is a thunk and return the result
func IsThunk(exp Expression) bool {
	switch exp.(type) {
	case *Thunk:
		return true
	default:
		return false
	}
}

// NewThunk creates a thunk and returns the pointer
func NewThunk(exp Expression, env *Env) *Thunk {
	return &Thunk{Env: env, Exp: exp}
}

// ActualValue returns the actual value of an expression.
// If the expression is a Thunk, eval and return the result, otherwise return the expression itself.
func ActualValue(exp Expression) (Expression, error) {
	switch p := exp.(type) {
	case *Thunk:
		return p.Value()
	default:
		return exp, nil
	}
}

// Represents Nil in scheme
type NilType struct{}

// Strings returns the string representing NilType.
func (n NilType) String() string {
	return "()"
}

// Object of NilType
var NilObj = NilType{}

// Undef represents undefined expression value.
type Undef struct{}

// String just implements the Stringer interface.
func (u Undef) String() string {
	return "<UNDEF>"
}

func extractList(expression Expression) (ret []Expression) {
	if !isList(expression) {
		return
	}
	switch v := expression.(type) {
	case *Pair:
		ret = append(ret, v.Car)
		ret = append(ret, extractList(v.Cdr)...)
		return
	default:
		return
	}
}

// Common Undef object.
var UndefObj = Undef{}

// IsNumber check whether the expression represents Number.
func IsNumber(exp Expression) bool {
	switch v := exp.(type) {
	case string:
		_, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return false
		}
		return true
	case Number:
		return true
	default:
		return false
	}
}

// IsString check whether the expression represents String in scheme.
func IsString(exp Expression) bool {
	switch v := exp.(type) {
	case string:
		ok, err := regexp.MatchString("\"(.|[\\r\\n])*\"", v)
		if ok && err == nil {
			return true
		}
		return false
	case String:
		return true
	default:
		return false
	}
}

// IsSyntaxExpression check whether the expression is a scheme syntax expression.
func IsSyntaxExpression(exp Expression) bool {
	ops, ok := exp.([]Expression)
	if !ok {
		return false
	}
	operator := ops[0]

	for key := range SyntaxMap {
		if key == operator {
			return true
		}
	}
	return false
}

// IsSymbol checks whether the expression is Symbol.
func IsSymbol(expression Expression) bool {
	_, ok := expression.([]Expression)
	if ok {
		return false
	}
	if _, ok := expression.(string); !ok {
		return false
	}
	if IsNumber(expression) || IsString(expression) || IsBoolean(expression) {
		return false
	}
	return true
}

// IsBoolean return true if the expression represents bool.
func IsBoolean(exp Expression) bool {
	_, ok := exp.(bool)
	if ok {
		return true
	}
	return exp == "#t" || exp == "#f"
}

// IsTrue check whether the condition is true. Return false when Exp is #f or false, otherwise return true
func IsTrue(exp Expression) bool {
	if exp == "#f" || exp == false {
		return false
	}
	return true
}

// IsNilObj returns true when the expression is NilTyp.
func IsNilObj(obj Expression) bool {
	switch obj.(type) {
	case NilType:
		return true
	default:
		return false
	}
}

// IsUndefObj returns true when the expression is Undef.
func IsUndefObj(obj Expression) bool {
	switch obj.(type) {
	case Undef:
		return true
	default:
		return false
	}
}

// IsPair checks whether the expression value is a *Pair.
func IsPair(obj Expression) bool {
	switch obj.(type) {
	case *Pair:
		return true
	default:
		return false
	}
}

// LambdaProcess wraps the body and env of a lambda expression
type LambdaProcess struct {
	params []Symbol
	body   []Expression // expressions of the lambda process
	env    *Env
}

// String implements the stringer interface
func (lambda *LambdaProcess) String() string {
	var buf bytes.Buffer
	buf.WriteString("(lambda (")
	for i, k := range lambda.params {
		buf.WriteString(string(k))
		if i != len(k)-1 {
			buf.WriteString(" ")
		}
	}
	buf.WriteString(") ")
	buf.WriteString(concatLambdaBodyToString(lambda.body))
	buf.WriteString(")")
	return buf.String()
}

// return the string represents the expression text
func expToPrintString(exp Expression) string {
	var buf bytes.Buffer
	switch v := exp.(type) {
	case []Expression:
		buf.WriteString("(")
		for i, exp := range v {
			buf.WriteString(expToPrintString(exp))
			if i != len(v)-1 {
				buf.WriteString(" ")
			}
		}
		buf.WriteString(")")
	default:
		buf.WriteString(fmt.Sprintf("%s", exp))
	}
	return buf.String()
}

func concatLambdaBodyToString(expressions []Expression) string {
	var buf bytes.Buffer
	for i, exp := range expressions {
		buf.WriteString(expToPrintString(exp))
		if i != len(expressions)-1 {
			buf.WriteString(" ")
		}
	}
	return buf.String()
}

// Return the expressions of body.
func (lambda *LambdaProcess) Body() Expression {
	if len(lambda.body) == 1 {
		return lambda.body[0]
	}
	return sequenceToExp(lambda.body)
}

// Pair combines the two values. Should only use with pointer
type Pair struct {
	Car, Cdr Expression
}

// IsNull checks whether the *Pair is null.
func (p *Pair) IsNull() bool {
	return p.Car == nil && p.Cdr == nil
}

// IsList check whether the *Pair is a well formed list.
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

// String returns the string representing the *Pair.
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

// check the result should print in console
func shouldPrint(exp Expression) bool {
	if exp == nil {
		return false
	}
	switch exp.(type) {
	case Undef:
		return false
	default:
		return true
	}
}

// Output string in interactive console that represents the expression value.
func valueToString(exp Expression) string {
	switch v := exp.(type) {
	case bool:
		if v == false {
			return "#f"
		}
		if v == true {
			return "#t"
		}
	default:
		return fmt.Sprintf("%v", exp)
	}
	return fmt.Sprintf("%v", exp)
}

// IsPrimitiveExpression checks whether the expressions value is the primitive types.
func IsPrimitiveExpression(exp Expression) bool {
	if IsNullExp(exp) || IsUndefObj(exp) ||
		IsQuote(exp) || IsNumber(exp) ||
		IsBoolean(exp) || IsString(exp) ||
		IsThunk(exp) || IsPair(exp) ||
		isList(exp) || IsLambdaType(exp) {
		return true
	}
	return false
}

// IsQuote check whether the value is Quote.
func IsQuote(exp Expression) bool {
	_, ok := exp.(Quote)
	return ok
}

// IsNullExp checks whether the expression represents Null(nil, NilType, blank list, blank expression).
func IsNullExp(exp Expression) bool {
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

// IsLambdaType checks whether this expression low level value is *LambdaProcess
func IsLambdaType(expression Expression) bool {
	_, ok := expression.(*LambdaProcess)
	return ok
}
