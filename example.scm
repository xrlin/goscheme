; Created by xrlin
; Usage: (load "example") or (load "example.scm")
; This example file contains the procedures to calculate Fibonacci number, calculate and display pascal triangle number.
(define (fib n)
 (if (<= n 2) 1 (+ (fib (- n 1)) (fib (- n 2)))))

(define (fib2 n)
 (begin (define (fib-iter a b n)
         (if (= n 0) b (fib-iter b (+ a b) (- n 1))))
  (fib-iter 0 1 (- n 1))))

(define (display-pascal n)
 (begin
  (define lst '())))

(define (add-num a b)
 (cond
  ((null? a) b)
  ((null? b) a)
  (else (+ a b))))

(define (calc-pascal lst k)
 (if (<= k 0) lst
  (begin (list-set! lst k (add-num (list-ref lst (- k 1)) (list-ref lst k)))
   (calc-pascal lst (- k 1)))))

(define (generate-pascal-helper sequence k n)
 (cond ((= k n) sequence)
  (else (generate-pascal-helper (append (calc-pascal sequence (- k 1)) 1) (+ k 1) n))))

(define (generate-pascal n)
 (generate-pascal-helper '() 0 n))

(define (display-pascal-helper sequence current-level max-level width)
 (begin
  (display-pascal-indents (- max-level current-level) width)
  (display-pascal-sequence sequence width)))

(define (display-blank n)
 (if (> n 0) (display " ")))

(define (display-tab n)
 (if (> n 0) (begin (display "\t") (display-tab (- n 1)))))

(define (display-pascal-indents indent-count width)
 (if (> indent-count 0) (begin (display-tab 1) (display-pascal-indents (- indent-count 1) width))))

(define (display-pascal-sequence sequence width)
 (if (null? sequence) (displayln "") (begin (display-pascal-num (car sequence) width) (display-pascal-sequence (cdr sequence) width))))

(define (max-pascal-number level)
 (reduce (lambda (x y) (if (> x y) x y)) (generate-pascal level)))

(define (number-char-width num)
 (begin
  (define (number-width num)
   (if (< num 10) 1 (+ (number-width (/ num 10)) 1)))
  (number-width num)))

(define (display-pascal-num num width)
 (if (<= (number-char-width num) width)
  (begin (display num) (display-blank (- width (number-char-width num))) (display-tab 2))))

(define (display-pascal-triangle-helper start-level max-level char-width)
 (if (<= start-level max-level) (begin
                                 (display-pascal-helper (generate-pascal start-level) start-level max-level char-width)
                                 (display-pascal-triangle-helper (+ start-level 1) max-level char-width))))

(define (display-pascal-triangle start-level max-level)
 (display-pascal-triangle-helper start-level max-level (number-char-width (max-pascal-number max-level))))
