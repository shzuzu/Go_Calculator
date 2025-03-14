package calc

import "errors"

var (
	ErrInvalidExpression = errors.New("invalid expression")
	ErrDivisionByZero    = errors.New("division by zero")
	ErrEOF               = errors.New("empty expression ")
	// ErrUnsupportedLiteral  = errors.New("unsupported literal type")
	// ErrUnsupportedOperator = errors.New("unsupported operator")
	// ErrUnsupportedNode     = errors.New("unsupported node type")
)
