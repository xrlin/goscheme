package goscheme

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/signal"
)

var exit = make(chan os.Signal, 1)

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

type InterpreterMode uint8

const (
	Interactive InterpreterMode = iota
	NoneInteractive
)

type Interpreter struct {
	currentFragment   []byte
	currentLineScript []byte
	interactive       bool
	scanner           *bufio.Scanner
	exit              chan os.Signal
	mode              InterpreterMode
	env               *Env
}

func (i *Interpreter) clean() {

}

func (i *Interpreter) Run() {
	if i.mode == Interactive {
		i.runInInteractiveMode()
		return
	}
	i.runNormal()

}

func (i *Interpreter) runNormal() {
	go i.checkAndExit()
	for {
		if eof := !i.scanner.Scan(); eof {
			if i.indents() != 0 {
				panic("syntax error: missing )")
			}
			return
		}
		b := i.scanner.Bytes()
		i.currentLineScript = b
		i.currentFragment = append(i.currentFragment, '\n')
		i.currentFragment = append(i.currentFragment, i.currentLineScript...)
		if i.indents() == 0 {
			tokenizer := NewTokenizerFromReader(bytes.NewReader(i.currentFragment))
			tokens := tokenizer.Tokens()
			expTokens, err := Parse(&tokens)
			if err != nil {
				fmt.Printf("%s\n", err)
				continue
			}
			EvalAll(expTokens, i.env)
			i.currentFragment = make([]byte, 0, 10)
		}
	}
}

func (i *Interpreter) runInInteractiveMode() {
	go i.checkAndExit()
	i.printTips()
	for {
		if eof := !i.scanner.Scan(); eof {
			i.exitProcess()
		}
		//i.scanner.Scan()
		b := i.scanner.Bytes()
		i.currentLineScript = b
		i.currentFragment = append(i.currentFragment, '\n')
		i.currentFragment = append(i.currentFragment, i.currentLineScript...)
		if i.indents() == 0 {
			tokenizer := NewTokenizerFromReader(bytes.NewReader(i.currentFragment))
			tokens := tokenizer.Tokens()
			expTokens, err := Parse(&tokens)
			if err != nil {
				fmt.Printf("%s\n", err)
				continue
			}
			ret := EvalAll(expTokens, i.env)
			if shouldPrint(ret) {
				fmt.Println(ret)
			}
			i.currentFragment = make([]byte, 0, 10)
		}
		i.printPrompt()
	}
}

func (i *Interpreter) exitProcess() {
	i.clean()
	fmt.Println("\nExiting...")
	os.Exit(0)
}

func (i *Interpreter) checkAndExit() {
	signal.Notify(i.exit, os.Interrupt)
	for {
		select {
		case <-i.exit:
			i.exitProcess()
		default:

		}
	}
}

func (i *Interpreter) printTips() {
	fmt.Println("Welcome to goscheme")
	i.printPrompt()
}

func (i *Interpreter) indents() int {
	return neededIndents(bytes.NewReader(i.currentFragment))
}

func (i *Interpreter) printPrompt() {
	prompt := ">"
	for c := 0; c < i.indents(); c++ {
		prompt += "\t"
	}
	fmt.Print(prompt)
}

func NewInterpreter(reader io.Reader, mode InterpreterMode) *Interpreter {
	scanner := bufio.NewScanner(reader)
	return &Interpreter{scanner: scanner, exit: exit, mode: mode, env: setupBuiltinEnv()}
}
