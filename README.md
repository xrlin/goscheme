# GoScheme
Yet, just an another shceme interpreter written in Go.

## Installation

```bash
go install github.com/xrlin/goscheme/cmd/goscheme/...
```

Or you can download the corresponding pre-compiled executable file in [release page](https://github.com/xrlin/goscheme/releases).

## Usage

```shell
# Just run goscheme to enter interactive shell
goscheme

# Run a scheme file
goscheme test.scm
```

## Features

* Interactive REPL shell

* Tail recursion optimization

* Lazy evaluation

* Short circut logic

* Type: `String`, `Number`, `Quote`, `LambdaProcess`, `Pair`, `Bool` ...

* syntax, builtin functions and procedures

    `load` 
    `define`
    `let`
    `let*`
    `letrec`
    `begin`
    `lambda`
    `and`
    `or`
    `not`
    `if`
    `cond`
    `delay`
    `map`
    `reduce`
    `force`
    `+`
    `-`
    `*`
    `/`
    `=`
    `cons`
    `list`
    `append`
    `list-length`
    `list-ref`
    `quote`
    `null?`
    `'`
    `eval`
    `apply`
    `set!`
    `set-cdr!`
    `set-car!`
    ... etc

Though it is a toy project just for fun and practive, Feel free to open an issue or make a merge request is you find bugs or have some suggestions.

Happy coding...


