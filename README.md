<h1 align="center">GoScheme</h1><br>
<p align="center">Just another shceme interpreter written in Go.</p>
<p align="center"><img src="https://travis-ci.org/xrlin/goscheme.svg?branch=master">&nbsp;<img src="https://goreportcard.com/badge/github.com/xrlin/goscheme"></p>

<p align="center">
  <img src="https://raw.githubusercontent.com/xrlin/goscheme/master/screenshots/repl.gif">
</p>

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

## Example

* calculate nth fibonacci number

    ```scheme
    ; calculate nth fibonacci number
    (define (fib n)
       (if (<= n 2) 1 (+ (fib (- n 1)) (fib (- n 2)))))
    
    (fib 10)
     
    ;#=> 55
     
    ; calculate nth fibnacci number in tail recursion
    (define (fib2 n)
       (begin (define (fib-iter a b n)
             (if (= n 0) b (fib-iter b (+ a b) (- n 1))))
       (fib-iter 0 1 (- n 1))))
    (fib2 30)
    ;#=>832040
    ```

* Mutually recursion

    ```scheme
    (letrec (
        (zero? (lambda (x) (= x 0)))
        (even?
        (lambda (n)
        (if (zero? n)
            #t
            (odd? (- n 1)))))
        (odd?
            (lambda (n)
            (if (zero? n)
                #f
                (even? (- n 1))))))
    (even? 88))
    ;#=>#t
    ```

Explore `example.scm` for more examples.


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


