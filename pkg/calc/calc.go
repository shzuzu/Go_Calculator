package calc

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

type Result struct {
	Expression string
	Value      float64
	Err        error
}

type WorkerPool struct {
	jobs     chan string
	results  chan Result
	workers  int
	stopChan chan struct{}
	wg       sync.WaitGroup
}

func NewWorkerPool(numWorkers int) *WorkerPool {
	return &WorkerPool{
		jobs:     make(chan string, numWorkers),
		results:  make(chan Result, numWorkers),
		workers:  numWorkers,
		stopChan: make(chan struct{}),
	}
}

func (wp *WorkerPool) Start() {
	log.Printf("WorkerPool: starting %d workers", wp.workers)
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker()
	}
}

func (wp *WorkerPool) worker() {
	defer wp.wg.Done()
	log.Println("Worker: started")
	for {
		select {
		case job, ok := <-wp.jobs:
			if !ok {
				log.Println("Worker: jobs channel closed, exiting")
				return
			}
			log.Printf("Worker: received job: %s", job)
			value, err := Calc(job)
			log.Printf("Worker: calculated result for job %s: %v, error: %v", job, value, err)
			wp.results <- Result{
				Expression: job,
				Value:      value,
				Err:        err,
			}
			log.Printf("Worker: result sent for job %s", job)
		case <-wp.stopChan:
			log.Println("Worker: stop signal received, exiting")
			return
		}
	}
}

func (wp *WorkerPool) Submit(expression string) {
	log.Printf("WorkerPool: submitting expression: %s", expression)
	wp.jobs <- expression
	log.Printf("WorkerPool: expression submitted: %s", expression)
}

func (wp *WorkerPool) GetResult() Result {
	log.Println("WorkerPool: waiting for result")
	result := <-wp.results
	log.Printf("WorkerPool: received result: %v, error: %v", result.Value, result.Err)
	return result
}
func (wp *WorkerPool) Close() {
	close(wp.jobs)
	close(wp.stopChan)
	wp.wg.Wait()
	close(wp.results)
}

func Calc(expression string) (float64, error) {
	log.Printf("Calc: parsing expression: %s", expression)
	node, err := parser.ParseExpr(expression)
	if expression == "" {
		log.Println("Calc: empty expression")
		return 0, ErrEOF
	}
	if err != nil {
		log.Printf("Calc: parsing error: %v", err)
		return 0, ErrInvalidExpression
	}
	return evalNode(node)
}
func (wp *WorkerPool) ValidateExpression(expression string) error {
	if strings.TrimSpace(expression) == "" {
		return ErrEOF
	}

	expr, err := parser.ParseExpr(expression)
	if err != nil {
		return ErrInvalidExpression
	}

	return validateNode(expr)
}

func validateNode(node ast.Node) error {
	switch n := node.(type) {
	case *ast.BinaryExpr:
		err := validateNode(n.X)
		if err != nil {
			return err
		}

		err = validateNode(n.Y)
		if err != nil {
			return err
		}

		if n.Op == token.QUO {
			if right, ok := n.Y.(*ast.BasicLit); ok {
				if right.Kind == token.INT || right.Kind == token.FLOAT {
					value, err := strconv.ParseFloat(right.Value, 64)
					if err == nil && value == 0 {
						return ErrDivisionByZero
					}
				}
			}
		}

		switch n.Op {
		case token.ADD, token.SUB, token.MUL, token.QUO:
			return nil
		default:
			return ErrInvalidExpression
		}

	case *ast.BasicLit:
		if n.Kind != token.FLOAT && n.Kind != token.INT {
			return ErrInvalidExpression
		}

		_, err := strconv.ParseFloat(n.Value, 64)
		if err != nil {
			return ErrInvalidExpression
		}

	case *ast.ParenExpr:
		return validateNode(n.X)

	case *ast.UnaryExpr:
		err := validateNode(n.X)
		if err != nil {
			return err
		}

		switch n.Op {
		case token.SUB, token.ADD:
			return nil
		default:
			return ErrInvalidExpression
		}

	default:
		return ErrInvalidExpression
	}

	return nil
}

func evalNode(node ast.Node) (float64, error) {
	// Формируем абсолютный путь к .env
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	envPath := filepath.Join(dir, "../../.env")

	err := godotenv.Load(envPath)
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	switch n := node.(type) {
	case *ast.BinaryExpr:
		log.Printf("evalNode: evaluating binary expression: %v", n)
		left, err := evalNode(n.X)
		if err != nil {
			log.Printf("evalNode: error evaluating left operand: %v", err)
			return 0, ErrInvalidExpression
		}
		right, err := evalNode(n.Y)
		if err != nil {
			log.Printf("evalNode: error evaluating right operand: %v", err)
			return 0, ErrInvalidExpression
		}

		switch n.Op {
		case token.ADD:
			ta, _ := strconv.Atoi(os.Getenv("TIME_ADDITION_MS"))
			sleepTime := time.Millisecond * time.Duration(ta)
			log.Printf("evalNode: sleeping for addition: %v", sleepTime)
			time.Sleep(sleepTime)
			log.Printf("evalNode: addition result: %f", left+right)
			return left + right, nil
		case token.SUB:
			ts, _ := strconv.Atoi(os.Getenv("TIME_SUBTRACTION_MS"))
			sleepTime := time.Millisecond * time.Duration(ts)
			log.Printf("evalNode: sleeping for subtraction: %v", sleepTime)
			time.Sleep(sleepTime)
			log.Printf("evalNode: subtraction result: %f", left-right)
			return left - right, nil
		case token.MUL:
			tm, _ := strconv.Atoi(os.Getenv("TIME_MULTIPLICATIONS_MS"))
			sleepTime := time.Millisecond * time.Duration(tm)
			log.Printf("evalNode: sleeping for multiplication: %v", sleepTime)
			time.Sleep(sleepTime)
			log.Printf("evalNode: multiplication result: %f", left*right)
			return left * right, nil
		case token.QUO:
			if right == 0 {
				log.Println("evalNode: division by zero")
				return 0, ErrDivisionByZero
			}
			td, _ := strconv.Atoi(os.Getenv("TIME_DIVISIONS_MS"))
			sleepTime := time.Millisecond * time.Duration(td)
			log.Printf("evalNode: sleeping for division: %v", sleepTime)
			time.Sleep(sleepTime)
			log.Printf("evalNode: division result: %f", left/right)
			return left / right, nil
		default:
			log.Printf("evalNode: unsupported binary operator: %v", n.Op)
			return 0, ErrInvalidExpression
		}

	case *ast.BasicLit:
		log.Printf("evalNode: evaluating literal: %v", n)
		if n.Kind != token.FLOAT && n.Kind != token.INT {
			log.Printf("evalNode: unsupported literal type: %v", n.Kind)
			return 0, ErrInvalidExpression
		}
		value, err := strconv.ParseFloat(n.Value, 64)
		if err != nil {
			log.Printf("evalNode: error parsing literal: %v", err)
			return 0, ErrInvalidExpression
		}
		log.Printf("evalNode: parsed literal: %f", value)
		return value, nil

	case *ast.ParenExpr:
		log.Println("evalNode: evaluating parenthesized expression")
		return evalNode(n.X)

	case *ast.UnaryExpr:
		log.Println("evalNode: evaluating unary expression")
		value, err := evalNode(n.X)
		if err != nil {
			log.Printf("evalNode: error evaluating unary operand: %v", err)
			return 0, ErrInvalidExpression
		}
		switch n.Op {
		case token.SUB:
			log.Printf("evalNode: unary negation result: %f", -value)
			return -value, nil
		case token.ADD:
			log.Printf("evalNode: unary plus result: %f", value)
			return value, nil
		default:
			log.Printf("evalNode: unsupported unary operator: %v", n.Op)
			return 0, ErrInvalidExpression
		}

	default:
		log.Printf("evalNode: unsupported node type: %T", node)
		return 0, ErrInvalidExpression
	}
}
