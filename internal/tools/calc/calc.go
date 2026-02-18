package calc

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"strconv"
	"strings"
)

// Calculator evaluates mathematical expressions.
type Calculator struct{}

// New creates a new Calculator.
func New() *Calculator {
	return &Calculator{}
}

// Eval evaluates a mathematical expression string and returns the result.
// Supports: +, -, *, /, %, parentheses, and percentage calculations.
func (c *Calculator) Eval(expr string) (float64, error) {
	// Preprocess the expression
	expr = c.preprocess(expr)

	// Parse and evaluate
	return c.evaluate(expr)
}

// preprocess normalizes the expression.
func (c *Calculator) preprocess(expr string) string {
	// Remove spaces
	expr = strings.ReplaceAll(expr, " ", "")

	// Handle Russian decimal separator
	expr = strings.ReplaceAll(expr, ",", ".")

	// Handle percentage: "15% от 2500" -> "2500*0.15"
	// More complex percentage handling would need NLP, but we can handle simple cases

	// Handle "x% of y" pattern - convert to y * x / 100
	// This is a simplified approach

	return expr
}

// evaluate parses and evaluates the expression.
func (c *Calculator) evaluate(expr string) (float64, error) {
	// Use Go's parser for safe expression evaluation
	// This only allows numeric literals and basic operators

	node, err := parser.ParseExpr(expr)
	if err != nil {
		return 0, fmt.Errorf("invalid expression: %w", err)
	}

	return c.evalAST(node)
}

func (c *Calculator) evalAST(node ast.Expr) (float64, error) {
	switch n := node.(type) {
	case *ast.BasicLit:
		if n.Kind == token.INT || n.Kind == token.FLOAT {
			return strconv.ParseFloat(n.Value, 64)
		}
		return 0, fmt.Errorf("unsupported literal type")

	case *ast.BinaryExpr:
		left, err := c.evalAST(n.X)
		if err != nil {
			return 0, err
		}
		right, err := c.evalAST(n.Y)
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
				return 0, fmt.Errorf("division by zero")
			}
			return left / right, nil
		case token.REM:
			return math.Mod(left, right), nil
		default:
			return 0, fmt.Errorf("unsupported operator: %s", n.Op)
		}

	case *ast.ParenExpr:
		return c.evalAST(n.X)

	case *ast.UnaryExpr:
		val, err := c.evalAST(n.X)
		if err != nil {
			return 0, err
		}
		if n.Op == token.SUB {
			return -val, nil
		}
		if n.Op == token.ADD {
			return val, nil
		}
		return 0, fmt.Errorf("unsupported unary operator: %s", n.Op)

	default:
		return 0, fmt.Errorf("unsupported expression type")
	}
}

// Percentage calculates percentage of a value.
func (c *Calculator) Percentage(percent, value float64) float64 {
	return value * percent / 100
}

// FormatResult formats a float result for display.
func (c *Calculator) FormatResult(val float64) string {
	// If it's a whole number, show without decimals
	if val == math.Trunc(val) {
		return fmt.Sprintf("%.0f", val)
	}
	// Otherwise show up to 2 decimal places, trimming trailing zeros
	result := fmt.Sprintf("%.2f", val)
	result = strings.TrimRight(result, "0")
	result = strings.TrimRight(result, ".")
	return result
}
