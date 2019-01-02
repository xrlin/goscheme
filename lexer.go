package goscheme

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode"
)

func Tokenize(inputScript string) []string {
	t := NewTokenizerFromString(inputScript)
	return t.Tokens()
}

type Tokenizer struct {
	Source       *bufio.Reader
	Eof          bool
	currentCh    rune
	currentToken string
}

func NewTokenizerFromString(input string) *Tokenizer {
	return &Tokenizer{Source: bufio.NewReader(strings.NewReader(input)), currentCh: -1}
}

func NewTokenizerFromReader(input io.Reader) *Tokenizer {
	return &Tokenizer{Source: bufio.NewReader(input), currentCh: -1}
}

func (t *Tokenizer) readAhead() {
	if t.Eof {
		return
	}
	r, _, err := t.Source.ReadRune()
	if err == io.EOF {
		t.Eof = true
		return
	}
	t.currentCh = r
}

func (t *Tokenizer) readString() (string, bool) {
	buf := make([]rune, 0, 10)
	buf = append(buf, '"')
	t.readAhead()
	for !t.Eof && t.currentCh != '"' {
		if t.currentCh == '\\' {
			t.readAhead()
			if t.currentCh == 'n' {
				buf = append(buf, '\n')
			} else {
				buf = append(buf, t.currentCh)
			}
			t.readAhead()
			continue
		}
		buf = append(buf, t.currentCh)
		t.readAhead()
	}
	if t.Eof {
		return "", !t.Eof
	}
	buf = append(buf, '"')
	t.readAhead()
	return string(buf), true
}

func (t *Tokenizer) readSymbol() (string, bool) {
	buf := make([]rune, 0, 1)
	if t.Eof {
		return "", false
	}
	for !t.Eof && isSymbolCh(t.currentCh) {
		buf = append(buf, t.currentCh)
		t.readAhead()
	}
	return string(buf), true
}

func isSymbolCh(r rune) bool {
	return !unicode.IsSpace(r) && !strings.ContainsRune("()'", r)
}

func (t *Tokenizer) readNextToken() (string, bool) {
	if t.Eof {
		t.currentCh = 0
		t.currentToken = ""
		return "", false
	}

	for t.currentCh == -1 || unicode.IsSpace(t.currentCh) {
		t.readAhead()
		if t.Eof {
			t.currentToken = ""
			t.currentCh = 0
			t.Eof = true
			return "", false
		}
	}

	if t.currentCh == '"' {
		return t.readString()
	} else if t.currentCh == '(' {
		t.readAhead()
		return "(", true
	} else if t.currentCh == ')' {
		t.readAhead()
		return ")", true
	} else if isSymbolCh(t.currentCh) {
		return t.readSymbol()
	} else if t.currentCh == '\'' {
		t.readAhead()
		return "'", true
	} else {
		return "", false
	}
}

func (t *Tokenizer) NextToken() (string, bool) {
	token, ok := t.readNextToken()
	t.currentToken = token
	return t.currentToken, ok
}

func (t *Tokenizer) Tokens() []string {
	var ret []string
	token, ok := t.NextToken()
	for ok {
		ret = append(ret, token)
		token, ok = t.NextToken()
	}
	return ret
}

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
