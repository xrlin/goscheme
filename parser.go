package goscheme

import "fmt"

// Parse read and parse the tokens to construct a syntax tree represents in nested slices.
func Parse(tokens *[]string) (ret []Expression, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s", r)
		}
	}()

	for len(*tokens) > 0 {
		ret = append(ret, readTokens(tokens))
	}
	return
}

func readTokens(tokens *[]string) Expression {
	if len(*tokens) == 0 {
		return nil
	}
	token := (*tokens)[0]
	*tokens = (*tokens)[1:]

	switch token {
	case "(":
		ret := make([]Expression, 0)
		for len(*tokens) >= 0 && (*tokens)[0] != ")" {
			nextPart := readTokens(tokens)
			ret = append(ret, nextPart)
		}
		if len(*tokens) == 0 {
			panic("syntax error: missing ')'")
		}
		*tokens = (*tokens)[1:]
		return ret
	case ")":
		panic("syntax error: unexpected ')'")
	case "'":
		ret := make([]Expression, 0, 4)
		ret = append(ret, "quote")
		nextPart := readTokens(tokens)
		ret = append(ret, nextPart)
		return ret
	default:
		return token
	}
}
