package goscheme

import "strings"

func Tokenize(inputScript string) (tokens []string) {
	tokenStr := strings.Replace(strings.Replace(inputScript, "(", "( ", -1), ")", " )", -1)
	for _, token := range strings.Split(tokenStr, " ") {
		if token != "" {
			tokens = append(tokens, token)
		}
	}
	return
}

func Parse(tokens []string) []Expression {
	return []Expression{}
}
