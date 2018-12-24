package goscheme

import (
	"io"
)

// return the indents current input string should add
func neededIndents(reader io.RuneReader) int {
	stack := make([]rune, 0, 3)

	for ch, _, err := reader.ReadRune(); err == nil; {
		if ch == '(' {
			stack = append(stack, ch)
		} else if ch == ')' {
			stack = stack[:len(stack)-1]
		}
		ch, _, err = reader.ReadRune()
	}
	return len(stack)
}
