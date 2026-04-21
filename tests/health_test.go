package health_test

import (
	"testing"
	"strings"
	"math"
)

func TestTrueIsTrue(t *testing.T) {
	if !true {
		t.Fatal("expected true to be true")
	}
}

func TestStringContains(t *testing.T) {
	s := "billing-service"
	if !strings.Contains(s, "billing") {
		t.Errorf("expected string to contain 'billing', got: %s", s)
	}
}

func TestAddition(t *testing.T) {
	result := 2 + 2
	if result != 4 {
		t.Errorf("expected 4, got %d", result)
	}
}

func TestSqrtIsPositive(t *testing.T) {
	result := math.Sqrt(16)
	if result <= 0 {
		t.Errorf("expected positive sqrt, got %f", result)
	}
}

func TestSliceLength(t *testing.T) {
	items := []string{"a", "b", "c"}
	if len(items) != 3 {
		t.Errorf("expected length 3, got %d", len(items))
	}
}