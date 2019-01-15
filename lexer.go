package goscheme

import (
	"bufio"
	"io"
	"strings"
	"unicode"
)

// Tokenize return the scheme tokens of input string
func Tokenize(inputScript string) []string {
	t := NewTokenizerFromString(inputScript)
	return t.Tokens()
}

// Tokenizer wraps the input to generate tokens.
type Tokenizer struct {
	Source       *bufio.Reader
	EOF          bool
	currentCh    rune
	currentToken string
}

// NewTokenizerFromString construct *Tokenizer from string
func NewTokenizerFromString(input string) *Tokenizer {
	return &Tokenizer{Source: bufio.NewReader(strings.NewReader(input)), currentCh: -1}
}

// NewTokenizerFromReader construct *Tokenizer from io.Reader
func NewTokenizerFromReader(input io.Reader) *Tokenizer {
	return &Tokenizer{Source: bufio.NewReader(input), currentCh: -1}
}

func (t *Tokenizer) readAhead() {
	if t.EOF {
		return
	}
	r, _, err := t.Source.ReadRune()
	if err == io.EOF {
		t.EOF = true
		return
	}
	t.currentCh = r
}

func (t *Tokenizer) readString() (string, bool) {
	buf := make([]rune, 0, 10)
	buf = append(buf, '"')
	t.readAhead()
	for !t.EOF && t.currentCh != '"' {
		if t.currentCh == '\\' {
			t.readAhead()
			if t.currentCh == 'n' {
				buf = append(buf, '\n')
			} else if t.currentCh == 't' {
				buf = append(buf, '\t')
			} else {
				buf = append(buf, t.currentCh)
			}
			t.readAhead()
			continue
		}
		buf = append(buf, t.currentCh)
		t.readAhead()
	}
	if t.EOF {
		return "", !t.EOF
	}
	buf = append(buf, '"')
	t.readAhead()
	return string(buf), true
}

func (t *Tokenizer) readSymbol() (string, bool) {
	buf := make([]rune, 0, 1)
	if t.EOF {
		return "", false
	}
	for !t.EOF && isSymbolCh(t.currentCh) {
		buf = append(buf, t.currentCh)
		t.readAhead()
	}
	return string(buf), true
}

func isSymbolCh(r rune) bool {
	return !unicode.IsSpace(r) && !strings.ContainsRune("()'", r)
}

func (t *Tokenizer) skipComment() {
	for t.currentCh == ';' {
		for t.currentCh != '\n' {
			t.readAhead()
			if t.EOF {
				return
			}
		}
		t.readAhead()
	}
}

func (t *Tokenizer) readNextToken() (string, bool) {

	if t.EOF {
		t.currentCh = 0
		t.currentToken = ""
		return "", false
	}

	for t.currentCh == -1 || unicode.IsSpace(t.currentCh) {
		t.readAhead()
		if t.EOF {
			t.currentToken = ""
			t.currentCh = 0
			t.EOF = true
			return "", false
		}
	}

	if t.currentCh == ';' {
		t.skipComment()
		return t.readNextToken()
	}
	if t.currentCh == '"' {
		return t.readString()
	}
	if t.currentCh == '(' {
		t.readAhead()
		return "(", true
	}
	if t.currentCh == ')' {
		t.readAhead()
		return ")", true
	}
	if isSymbolCh(t.currentCh) {
		return t.readSymbol()
	}
	if t.currentCh == '\'' {
		t.readAhead()
		return "'", true
	}
	return "", false
}

// NextToken read ahead and returns the next valid token.
func (t *Tokenizer) NextToken() (string, bool) {
	token, ok := t.readNextToken()
	t.currentToken = token
	return t.currentToken, ok
}

// Tokens returns all the tokens
func (t *Tokenizer) Tokens() []string {
	var ret []string
	token, ok := t.NextToken()
	for ok {
		ret = append(ret, token)
		token, ok = t.NextToken()
	}
	return ret
}
