package goscheme

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Expression interface{}

type Number float64

type String string

func (s String) String() string {
	return "\"" + string(s) + "\""
}

type Symbol string

type Quote string

type Boolean bool

type commonFunction func(args ...Expression) Expression

type Function struct {
	name     string
	function commonFunction
	minArgs  int
	maxArgs  int
}

func (f Function) Call(args ...Expression) Expression {
	if err := f.validateArgCount(args...); err != nil {
		panic(err)
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

// Thunk wraps expressoin for lazy execution
// Thunk should use as pointer
type Thunk struct {
	// expression to execute
	Exp Expression
	// expression result cache
	ret Expression
	// context to execute Exp
	Env *Env
}

func (t Thunk) String() string {
	if t.ret != nil {
		return fmt.Sprintf("#[Thunk %s]", t.ret)
	}
	return fmt.Sprintf("#[Thunk exp: %s]", t.Exp)
}

// Value returns the actual value of the thunk
func (t *Thunk) Value() Expression {
	if t.ret != nil {
		return t.ret
	}
	value := Eval(t.Exp, t.Env)
	switch t2 := value.(type) {
	case *Thunk:
		value = t2.Value()
	default:
	}
	t.ret = value
	return t.ret
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
func ActualValue(exp Expression) Expression {
	switch p := exp.(type) {
	case *Thunk:
		return p.Value()
	default:
		return exp
	}
}

type NilType struct{}

func (n NilType) String() string {
	return "()"
}

var syntaxes = [...]string{"define", "lambda", "if", "let", "cond", "begin", "quote", "eval", "apply", "load", "delay", "and", "or", "let", "let*", "letrec", "set!"}

var NilObj = NilType{}

type Undef struct{}

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

var undefObj = Undef{}

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
	return false
}

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
	return false
}

func IsSpecialSyntaxExpression(exp Expression, name string) bool {
	ops, ok := exp.([]Expression)
	if !ok {
		return false
	}
	operator := ops[0]
	return operator == name
}

func IsSyntaxExpression(exp Expression) bool {
	ops, ok := exp.([]Expression)
	if !ok {
		return false
	}
	operator := ops[0]

	for _, s := range syntaxes {
		if s == operator {
			return true
		}
	}
	return false
}

func IsSymbol(expression Expression) bool {
	_, ok := expression.([]Expression)
	if ok {
		return false
	}
	if _, ok := expression.(string); !ok {
		return false
	}
	if IsNumber(expression) || IsString(expression) || IsBoolean(expression) || IsSyntaxExpression(expression) {
		return false
	}
	return true
}

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

func IsNilObj(obj Expression) bool {
	switch obj.(type) {
	case NilType:
		return true
	default:
		return false
	}
}

func isUndefObj(obj Expression) bool {
	switch obj.(type) {
	case Undef:
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

type LambdaProcess struct {
	params []Symbol
	body   []Expression // expressions of the lambda process
	env    *Env
}

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

func (lambda *LambdaProcess) Body() Expression {
	if len(lambda.body) == 1 {
		return lambda.body[0]
	}
	return sequenceToExp(lambda.body)
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
