package goscheme

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/c-bata/go-prompt"
	"io"
	"os"
	"os/signal"
)

var exit = make(chan os.Signal, 1)

// return the indents current input string should add
// if result > 0 missing ) , if result < 0 missing (, if result == 0 syntax check passed.
func neededIndents(reader io.RuneReader) int {
	stack := make([]rune, 0, 3)

	for ch, _, err := reader.ReadRune(); err == nil; {
		if ch == '(' {
			stack = append(stack, ch)
		} else if ch == ')' {
			if len(stack)-1 < 0 {
				return len(stack) - 1
			}
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
	input             io.Reader
	prompt            *prompt.Prompt
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
	go i.checkExit()
	i.check()
	scanner := bufio.NewScanner(i.input)
	for {
		if eof := !scanner.Scan(); eof {
			if i.indents() != 0 {
				panic("syntax error: missing )")
			}
			return
		}
		b := scanner.Bytes()
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

// check whether the input has syntax error
func (i *Interpreter) check() {
	var buf bytes.Buffer
	teeReader := io.TeeReader(i.input, &buf)

	defer func() {
		i.input = &buf
	}()

	reader := bufio.NewReader(teeReader)
	indents := neededIndents(reader)
	if indents != 0 {
		if indents < 0 {
			fmt.Println("syntax error: missing (")
		} else {
			fmt.Println("syntax error: missing )")
		}
		i.exit <- os.Interrupt
	}
}

func (i *Interpreter) runInInteractiveMode() {
	i.printTips()
	go i.checkExit()
	i.prompt.Run()
}

func (i *Interpreter) exitProcess() {
	i.clean()
	fmt.Println("\nExiting...")
	os.Exit(0)
}

func (i *Interpreter) checkExit() {
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
}

func (i *Interpreter) indents() int {
	return neededIndents(bytes.NewReader(i.currentFragment))
}

func (i *Interpreter) printIndents() {
	var s string
	for c := 0; c < i.indents(); c++ {
		s += "\t"
	}
	fmt.Print(s)
}

func (i *Interpreter) initPromote() {
	p := prompt.New(func(s string) {
		i.evalPromptInput(s)
	}, func(document prompt.Document) []prompt.Suggest {
		return []prompt.Suggest{}
	}, prompt.OptionLivePrefix(func() (prefix string, useLivePrefix bool) {
		prefix = ">>>"
		useLivePrefix = true
		if i.indents() > 0 {
			prefix = ""
		}
		return
	}))
	i.prompt = p
}

func (i *Interpreter) evalPromptInput(input string) {
	// after each exec should init the interpreter status
	defer func() {
		i.currentLineScript = make([]byte, 0, 10)
		i.currentFragment = make([]byte, 0, 10)
	}()

	i.currentLineScript = []byte(input)
	i.currentFragment = append(i.currentFragment, '\n')
	i.currentFragment = append(i.currentFragment, i.currentLineScript...)
	if i.indents() <= 0 {
		tokenizer := NewTokenizerFromReader(bytes.NewReader(i.currentFragment))
		tokens := tokenizer.Tokens()
		expTokens, err := Parse(&tokens)
		if err != nil {
			fmt.Printf("%s\n", err)
			return
		}
		ret := EvalAll(expTokens, i.env)
		if shouldPrint(ret) {
			fmt.Println(toString(ret))
		}
		i.currentFragment = make([]byte, 0, 10)
	}
	i.printIndents()
}

func NewFileInterpreter(reader io.Reader) *Interpreter {
	return &Interpreter{input: reader, exit: exit, mode: NoneInteractive, env: setupBuiltinEnv()}
}

func NewFileInterpreterWithEnv(reader io.Reader, env *Env) *Interpreter {
	return &Interpreter{input: reader, exit: exit, mode: NoneInteractive, env: env}
}

func NewREPLInterpreter() *Interpreter {
	i := &Interpreter{exit: exit, mode: Interactive, env: setupBuiltinEnv()}
	i.initPromote()
	return i
}
