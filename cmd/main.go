package main

import (
	"fmt"
	"github.com/xrlin/goscheme"
	"os"
)

func main() {
	var filePath string
	if len(os.Args) >= 2 {
		filePath = os.Args[1]
	}
	var interpreter *goscheme.Interpreter
	if filePath == "" {
		interpreter = goscheme.NewInterpreter(os.Stdin, goscheme.Interactive)
	} else {
		file, err := os.Open(filePath)
		if err != nil {
			fmt.Println(err)
			return
		}
		interpreter = goscheme.NewInterpreter(file, goscheme.NoneInteractive)
	}
	interpreter.Run()
}
