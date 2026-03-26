package builtin

import (
	"context"
	"fmt"
	"math"
)

type Calculator struct{}

func NewCalculator() *Calculator { return &Calculator{} }

func (c *Calculator) Name() string { return "calculator" }
func (c *Calculator) Description() string {
	return "Evaluate a mathematical expression. Input: {\"expression\": \"2+2\"} or {\"operation\": \"add\", \"a\": 2, \"b\": 2}"
}

func (c *Calculator) Execute(ctx context.Context, args map[string]any) (any, error) {
	if expr, ok := args["expression"].(string); ok && expr != "" {
		return c.evaluateExpression(expr)
	}
	return c.evaluateByOperation(args)
}

func (c *Calculator) evaluateExpression(expr string) (any, error) {
	result := evalSimpleMath(expr)
	if result == nil {
		return nil, fmt.Errorf("cannot evaluate expression: %s", expr)
	}
	return result, nil
}

func (c *Calculator) evaluateByOperation(args map[string]any) (any, error) {
	op, _ := args["operation"].(string)
	a, _ := toFloat64(args["a"])
	b, _ := toFloat64(args["b"])

	switch op {
	case "add":
		return a + b, nil
	case "subtract":
		return a - b, nil
	case "multiply":
		return a * b, nil
	case "divide":
		if b == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return a / b, nil
	case "power":
		return math.Pow(a, b), nil
	case "sqrt":
		return math.Sqrt(a), nil
	default:
		return nil, fmt.Errorf("unknown operation: %s", op)
	}
}

func toFloat64(v any) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case float64:
		return n, true
	case string:
		var f float64
		fmt.Sscanf(n, "%f", &f)
		return f, true
	default:
		return 0, false
	}
}

func evalSimpleMath(expr string) any {
	var a, b float64
	var op byte

	if n, err := fmt.Sscanf(expr, "%f %c %f", &a, &op, &b); n == 3 && err == nil {
		switch op {
		case '+':
			return a + b
		case '-':
			return a - b
		case '*':
			return a * b
		case '/':
			if b != 0 {
				return a / b
			}
		case '^':
			return math.Pow(a, b)
		}
	}

	return nil
}
