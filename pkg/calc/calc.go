package calc

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
)

func Calc(expression string) (float64, error) {
	node, err := parser.ParseExpr(expression)
	if err != nil {
		return 0, ErrInvalidExpression
	}
	return evalNode(node)
}

func evalNode(node ast.Node) (float64, error) {
	switch n := node.(type) {
	case *ast.BinaryExpr:
		left, err := evalNode(n.X)
		if err != nil {
			return 0, err
		}
		right, err := evalNode(n.Y)
		if err != nil {
			return 0, err
		}

		switch n.Op {
		case token.ADD:
			return left + right, nil
		case token.SUB:
			return left - right, nil
		case token.MUL:
			return left * right, nil
		case token.QUO:
			if right == 0 {
				return 0, ErrDivisionByZero
			}
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

	default:
		return 0, ErrInvalidExpression
	}
}
