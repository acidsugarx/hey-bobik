package calc

import (
	"math"
	"testing"
)

func TestEvalBasicOperations(t *testing.T) {
	c := New()

	tests := []struct {
		expr     string
		expected float64
	}{
		{"2+2", 4},
		{"10-3", 7},
		{"5*6", 30},
		{"20/4", 5},
		{"10%3", 1},
		{"2+3*4", 14},
		{"(2+3)*4", 20},
		{"-5+10", 5},
		{"100/4/5", 5},
	}

	for _, tt := range tests {
		result, err := c.Eval(tt.expr)
		if err != nil {
			t.Errorf("Eval(%q) error: %v", tt.expr, err)
			continue
		}
		if math.Abs(result-tt.expected) > 0.0001 {
			t.Errorf("Eval(%q) = %v, want %v", tt.expr, result, tt.expected)
		}
	}
}

func TestEvalFloats(t *testing.T) {
	c := New()

	result, err := c.Eval("3.14*2")
	if err != nil {
		t.Fatalf("Eval error: %v", err)
	}
	if math.Abs(result-6.28) > 0.0001 {
		t.Errorf("expected 6.28, got %v", result)
	}
}

func TestEvalDivisionByZero(t *testing.T) {
	c := New()

	_, err := c.Eval("10/0")
	if err == nil {
		t.Error("expected division by zero error")
	}
}

func TestEvalInvalidExpression(t *testing.T) {
	c := New()

	_, err := c.Eval("abc+def")
	if err == nil {
		t.Error("expected error for invalid expression")
	}
}

func TestPercentage(t *testing.T) {
	c := New()

	result := c.Percentage(15, 2500)
	if result != 375 {
		t.Errorf("Percentage(15, 2500) = %v, want 375", result)
	}

	result = c.Percentage(50, 200)
	if result != 100 {
		t.Errorf("Percentage(50, 200) = %v, want 100", result)
	}
}

func TestFormatResult(t *testing.T) {
	c := New()

	tests := []struct {
		val      float64
		expected string
	}{
		{100, "100"},
		{3.14, "3.14"},
		{2.5, "2.5"},
		{1.00, "1"},
		{0.1, "0.1"},
	}

	for _, tt := range tests {
		result := c.FormatResult(tt.val)
		if result != tt.expected {
			t.Errorf("FormatResult(%v) = %q, want %q", tt.val, result, tt.expected)
		}
	}
}

func TestPreprocess(t *testing.T) {
	c := New()

	// Test comma to dot conversion
	result, err := c.Eval("3,14*2")
	if err != nil {
		t.Fatalf("Eval error: %v", err)
	}
	if math.Abs(result-6.28) > 0.0001 {
		t.Errorf("expected 6.28, got %v", result)
	}
}
