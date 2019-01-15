package goscheme

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/c-bata/go-prompt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
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

// InterpreterMode represents mode the interpreter will run
type InterpreterMode uint8

const (
	// Interactive set the interpreter running as interactive shell
	Interactive InterpreterMode = iota
	// NoneInteractive set the interpreter running in normal mode without shell.
	NoneInteractive
)

// Interpreter read from source and evaluate them.
type Interpreter struct {
	currentFragment   []byte
	currentLineScript []byte
	input             io.Reader
	prompt            *prompt.Prompt
	exit              chan os.Signal
	mode              InterpreterMode
	consoleWriter     prompt.ConsoleWriter
	env               *Env
}

// Run start the interpreter and evaluate the input.
func (i *Interpreter) Run() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("error raised: %v", r)
			log.Println(err)
		}
	}()
	if i.mode == Interactive {
		i.runInInteractiveMode()
		return nil
	}
	i.runNormal()
	return nil
}

func (i *Interpreter) runNormal() error {
	go i.checkExit()
	i.check()
	scanner := bufio.NewScanner(i.input)
	for {
		if eof := !scanner.Scan(); eof {
			if i.indents() != 0 {
				return errors.New("syntax error: missing )")
			}
			return nil
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
				return err
			}
			_, err = EvalAll(expTokens, i.env)
			if err != nil {
				return err
			}
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
	fmt.Println("\nExiting...")
	os.Exit(0)
}

func (i *Interpreter) checkExit() {
	signal.Notify(i.exit, os.Interrupt)
	for {
		select {
		case <-i.exit:
			i.exitProcess()
		}
	}
}

func (i *Interpreter) printTips() {
	fmt.Println(
		`Welcome to goscheme.
Enter '(exit)' or CTRL+D to exit.`)
}

func (i *Interpreter) indents() int {
	codeFragment := i.currentFragment
	return neededIndents(bytes.NewReader(codeFragment))
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
	}, i.completer, prompt.OptionLivePrefix(func() (prefix string, useLivePrefix bool) {
		prefix = ">>>"
		useLivePrefix = true
		if i.indents() > 0 {
			prefix = ""
		}
		return
	}), prompt.OptionWriter(i.writer()))
	i.prompt = p
}

func (i *Interpreter) writer() prompt.ConsoleWriter {
	if i.consoleWriter == nil {
		i.consoleWriter = prompt.NewStdoutWriter()
	}
	return i.consoleWriter
}

func (i *Interpreter) print(text string, color prompt.Color) {
	writer := i.writer()
	writer.SetColor(color, prompt.DefaultColor, true)
	writer.WriteStr(text)
	writer.SetColor(prompt.DefaultColor, prompt.DefaultColor, false)
}

func (i *Interpreter) completer(d prompt.Document) (ret []prompt.Suggest) {
	source := d.GetWordBeforeCursor()
	key := strings.Trim(d.GetWordBeforeCursor(), "(")
	completeText := func(s string) string {
		return strings.Replace(source, key, s, 1)
	}
	if key == "" {
		return
	}
	for _, s := range i.env.Symbols() {
		if strings.HasPrefix(string(s), key) {
			ret = append(ret, prompt.Suggest{Text: completeText(string(s)), Description: string(s)})
		}
	}
	return
}

func (i *Interpreter) evalPromptInput(input string) {
	// if exec failed should re-init the interpreter's internal value
	defer func() {
		if r := recover(); r != nil {
			i.currentLineScript = make([]byte, 0, 10)
			i.currentFragment = make([]byte, 0, 10)
		}
	}()

	i.currentLineScript = []byte(input)
	i.currentFragment = append(i.currentFragment, '\n')
	i.currentFragment = append(i.currentFragment, i.currentLineScript...)
	if i.indents() <= 0 {
		tokenizer := NewTokenizerFromReader(bytes.NewReader(i.currentFragment))
		tokens := tokenizer.Tokens()
		expTokens, err := Parse(&tokens)
		if err != nil {
			i.print(fmt.Sprintf("%s\n", err), prompt.Red)
			return
		}
		ret, err := EvalAll(expTokens, i.env)
		if err != nil {
			i.print(fmt.Sprintf("err:=>%s\n", err), prompt.Red)
		}
		if shouldPrint(ret) && err == nil {
			i.print(fmt.Sprintf("#=>%s\n", valueToString(ret)), prompt.Green)
		}
		i.currentFragment = make([]byte, 0, 10)
	}
	i.printIndents()
}

// NewFileInterpreter construct a *Interpreter from file.
func NewFileInterpreter(reader io.Reader) *Interpreter {
	return &Interpreter{input: reader, exit: exit, mode: NoneInteractive, env: setupBuiltinEnv()}
}

// NewFileInterpreterWithEnv construct a *Interpreter from io.reader init with env.
func NewFileInterpreterWithEnv(reader io.Reader, env *Env) *Interpreter {
	return &Interpreter{input: reader, exit: exit, mode: NoneInteractive, env: env}
}

// NewREPLInterpreter construct a REPL *Interpreter.
func NewREPLInterpreter() *Interpreter {
	i := &Interpreter{exit: exit, mode: Interactive, env: setupBuiltinEnv()}
	i.initPromote()
	return i
}
