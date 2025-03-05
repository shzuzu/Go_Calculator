package calc

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strconv"
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
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker()
	}
}

func (wp *WorkerPool) worker() {
	defer wp.wg.Done()
	for {
		select {
		case job, ok := <-wp.jobs:
			if !ok {
				return
			}
			value, err := Calc(job)
			wp.results <- Result{
				Expression: job,
				Value:      value,
				Err:        err,
			}
		case <-wp.stopChan:
			return
		}
	}
}

func (wp *WorkerPool) Submit(expression string) {
	wp.jobs <- expression
}

func (wp *WorkerPool) GetResult() Result {
	return <-wp.results
}

func (wp *WorkerPool) Close() {
	close(wp.jobs)
	close(wp.stopChan)
	wp.wg.Wait()
	close(wp.results)
}

func Calc(expression string) (float64, error) {
	node, err := parser.ParseExpr(expression)
	if expression == "" {
		fmt.Println("\nYou shold enter an expression")
		return 0, ErrEOF
	}
	if err != nil {
		return 0, ErrInvalidExpression
	}
	return evalNode(node)
}

func evalNode(node ast.Node) (float64, error) {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	switch n := node.(type) {
	case *ast.BinaryExpr:
		left, err := evalNode(n.X)
		if err != nil {
			return 0, ErrInvalidExpression
		}
		right, err := evalNode(n.Y)
		if err != nil {
			return 0, ErrInvalidExpression
		}

		switch n.Op {
		case token.ADD:
			ta, _ := strconv.Atoi(os.Getenv("TIME_ADDITION_MS"))
			sleepTime := time.Millisecond * time.Duration(ta)
			time.Sleep(sleepTime)

			return left + right, nil
		case token.SUB:
			ta, _ := strconv.Atoi(os.Getenv("TIME_SUBTRACTION_MS"))
			sleepTime := time.Millisecond * time.Duration(ta)
			time.Sleep(sleepTime)

			return left - right, nil
		case token.MUL:
			ta, _ := strconv.Atoi(os.Getenv("TIME_MULTIPLICATIONS_MS"))
			sleepTime := time.Millisecond * time.Duration(ta)
			time.Sleep(sleepTime)

			return left * right, nil
		case token.QUO:
			if right == 0 {
				return 0, ErrDivisionByZero
			}
			ta, _ := strconv.Atoi(os.Getenv("TIME_DIVISIONS_MS"))
			sleepTime := time.Millisecond * time.Duration(ta)
			time.Sleep(sleepTime)

			return left / right, nil
		default:
			return 0, ErrInvalidExpression
		}

	case *ast.BasicLit:
		if n.Kind != token.FLOAT && n.Kind != token.INT {
			return 0, ErrInvalidExpression
		}
		return strconv.ParseFloat(n.Value, 64)

	case *ast.ParenExpr:
		// Вычисляем выражение внутри скобок
		return evalNode(n.X)

	case *ast.UnaryExpr:
		// учет унарных операторов
		value, err := evalNode(n.X)
		if err != nil {
			return 0, ErrInvalidExpression
		}
		switch n.Op {
		case token.SUB:
			return -value, nil
		case token.ADD:
			return value, nil
		default:
			return 0, ErrInvalidExpression
		}

	default:
		return 0, ErrInvalidExpression
	}
}

//я пытался
// package calc

// import (
// 	"fmt"
// 	"go/ast"
// 	"go/parser"
// 	"go/token"
// 	"strconv"
// 	"sync"
// )

// type Task struct {
// 	ID            string  `json:"id"`
// 	Arg1          float64 `json:"arg1"`
// 	Arg2          float64 `json:"arg2"`
// 	Operation     string  `json:"operation"`
// 	OperationTime int     `json:"operation_time"`
// }

// type Tasks struct {
// 	Tasks []Task `json:"tasks"`
// }

// func createTask(node ast.Node) ([]Task, error) {
// 	var tasks []Task

// 	switch n := node.(type) {
// 	case *ast.BinaryExpr:
// 		leftTasks, err := createTask(n.X)
// 		if err != nil {
// 			return nil, err
// 		}
// 		tasks = append(tasks, leftTasks...)

// 		rightTasks, err := createTask(n.Y)
// 		if err != nil {
// 			return nil, err
// 		}
// 		tasks = append(tasks, rightTasks...)

// 		task := Task{
// 			Operation: n.Op.String(),
// 		}

// 		if lit, ok := n.X.(*ast.BasicLit); ok {
// 			arg1, err := strconv.ParseFloat(lit.Value, 64)
// 			if err != nil {
// 				return nil, ErrInvalidExpression
// 			}
// 			task.Arg1 = arg1
// 		}

// 		if lit, ok := n.Y.(*ast.BasicLit); ok {
// 			arg2, err := strconv.ParseFloat(lit.Value, 64)
// 			if err != nil {
// 				return nil, ErrInvalidExpression
// 			}
// 			task.Arg2 = arg2
// 		}
// 		task.OperationTime = 1

// 		tasks = append(tasks, task)

// 	case *ast.BasicLit:
// 		if n.Kind == token.INT || n.Kind == token.FLOAT {
// 			value, err := strconv.ParseFloat(n.Value, 64)
// 			if err != nil {
// 				return nil, ErrInvalidExpression
// 			}
// 			tasks = append(tasks, Task{
// 				Arg1: value,
// 			})
// 		}

// 	case *ast.ParenExpr:
// 		return createTask(n.X)

// 	case *ast.UnaryExpr:
// 		valueTasks, err := createTask(n.X)
// 		if err != nil {
// 			return nil, err
// 		}
// 		tasks = append(tasks, valueTasks...)

// 	default:
// 		return nil, ErrInvalidExpression
// 	}

// 	return tasks, nil
// }

// func executeTask(task Task) float64 {
// 	switch task.Operation {
// 	case "+":
// 		return task.Arg1 + task.Arg2
// 	case "-":
// 		return task.Arg1 - task.Arg2
// 	case "*":
// 		return task.Arg1 * task.Arg2
// 	case "/":
// 		if task.Arg2 == 0 {
// 			panic("division by zero")
// 		}
// 		return task.Arg1 / task.Arg2
// 	default:
// 		return task.Arg1
// 	}
// }

// func worker(tasks <-chan Task, results chan<- float64, wg *sync.WaitGroup) {
// 	defer wg.Done()
// 	for task := range tasks {
// 		results <- executeTask(task)
// 	}
// }

// func Calc(expression string, numWorkers int) (float64, error) {
// 	node, err := parser.ParseExpr(expression)
// 	if expression == "" {
// 		return 0, fmt.Errorf("you should enter an expression")
// 	}
// 	if err != nil {
// 		return 0, ErrInvalidExpression
// 	}

// 	tasks, err := createTask(node)
// 	if err != nil {
// 		return 0, err
// 	}

// 	taskChan := make(chan Task, len(tasks))
// 	resultChan := make(chan float64, len(tasks))

// 	var wg sync.WaitGroup

// 	for i := 0; i < numWorkers; i++ {
// 		wg.Add(1)
// 		go worker(taskChan, resultChan, &wg)
// 	}

// 	for _, task := range tasks {
// 		taskChan <- task
// 	}
// 	close(taskChan)

// 	wg.Wait()
// 	close(resultChan)

// 	var result float64
// 	for res := range resultChan {
// 		result += res
// 	}

// 	return result, nil
// }
