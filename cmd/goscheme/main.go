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
		interpreter = goscheme.NewREPLInterpreter()
	} else {
		file, err := os.Open(filePath)
		if err != nil {
			fmt.Println(err)
			return
		}
		interpreter = goscheme.NewFileInterpreter(file)
	}
	interpreter.Run()
}
