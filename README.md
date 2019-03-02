<h1 align="center">GoScheme</h1><br>
<p align="center">Just another shceme interpreter written in Go.</p>
<p align="center">
  <a href="https://travis-ci.org/xrlin/goscheme"><img src="https://travis-ci.org/xrlin/goscheme.svg?branch=master"><a/>
  <a href="https://godoc.org/github.com/xrlin/goscheme" rel="nofollow">
    <img src="https://camo.githubusercontent.com/4953dcce3ef06016a8f872b20e3bf6cd65e99621/68747470733a2f2f696d672e736869656c64732e696f2f62616467652f676f646f632d7265666572656e63652d3532373242342e737667" alt="godoc" style="max-width:100%;">
  </a>
  <a href="https://goreportcard.com/report/github.com/xrlin/goscheme"><img src="https://goreportcard.com/badge/github.com/xrlin/goscheme"></a>
</p>

<p align="center">
  <img src="https://raw.githubusercontent.com/xrlin/goscheme/master/screenshots/repl.gif">
</p>

## Installation

```bash
go get github.com/xrlin/goscheme/cmd/goscheme
```

Or you can download the corresponding pre-compiled executable file in [release page](https://github.com/xrlin/goscheme/releases).

## Usage

```shell
# Just run goscheme to enter interactive shell
goscheme

# Run a scheme file
goscheme test.scm
```

## Examples

* Calculate nth fibonacci number

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

Though it is a toy project just for fun and practice, Feel free to open an issue or make a merge request is you find bugs or have some suggestions.

Happy coding...


