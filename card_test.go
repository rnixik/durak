package main

import "testing"

func TestGetValueIndex(t *testing.T) {
	card := &Card{"9", "♦"}
	got := card.getValueIndex()
	expected := 3
	if got != expected {
		t.Errorf("getValueIndex expected: %v, got: %v", expected, got)
	}
}

func TestGtFalse(t *testing.T) {
	card := &Card{"9", "♦"}
	otherCard := &Card{"10", "♠"}
	got := card.gt(otherCard)
	expected := false
	if got != expected {
		t.Errorf("TestGtFalse expected: %v, got: %v", expected, got)
	}
}

func TestGtTrue(t *testing.T) {
	card := &Card{"9", "♦"}
	otherCard := &Card{"8", "♦"}
	got := card.gt(otherCard)
	expected := true
	if got != expected {
		t.Errorf("TestGtTrue expected: %v, got: %v", expected, got)
	}
}

func TestGtWithEqual(t *testing.T) {
	card := &Card{"9", "♦"}
	otherCard := &Card{"9", "♠"}
	got := card.gt(otherCard)
	expected := false
	if got != expected {
		t.Errorf("TestGtWithEqual expected: %v, got: %v", expected, got)
	}
}

func TestGte(t *testing.T) {
	card := &Card{"9", "♦"}
	otherCard := &Card{"9", "♦"}
	got := card.gte(otherCard)
	expected := true
	if got != expected {
		t.Errorf("TestGte expected: %v, got: %v", expected, got)
	}
}

func TestGteFalse(t *testing.T) {
	card := &Card{"J", "♦"}
	otherCard := &Card{"Q", "♠"}
	got := card.gte(otherCard)
	expected := false
	if got != expected {
		t.Errorf("TestGteFalse expected: %v, got: %v", expected, got)
	}
}

func TestLtFalse(t *testing.T) {
	card := &Card{"9", "♦"}
	otherCard := &Card{"10", "♦"}
	got := card.lt(otherCard)
	expected := true
	if got != expected {
		t.Errorf("TestLtFalse expected: %v, got: %v", expected, got)
	}
}

func TestLtTrue(t *testing.T) {
	card := &Card{"9", "♦"}
	otherCard := &Card{"8", "♠"}
	got := card.lt(otherCard)
	expected := false
	if got != expected {
		t.Errorf("TestLtTrue expected: %v, got: %v", expected, got)
	}
}

func TestLtWithEqual(t *testing.T) {
	card := &Card{"9", "♦"}
	otherCard := &Card{"9", "♠"}
	got := card.lt(otherCard)
	expected := false
	if got != expected {
		t.Errorf("TestLtWithEqual expected: %v, got: %v", expected, got)
	}
}

func TestLte(t *testing.T) {
	card := &Card{"9", "♦"}
	otherCard := &Card{"9", "♦"}
	got := card.lte(otherCard)
	expected := true
	if got != expected {
		t.Errorf("TestLte expected: %v, got: %v", expected, got)
	}
}

func TestLteFalse(t *testing.T) {
	card := &Card{"K", "♦"}
	otherCard := &Card{"Q", "♠"}
	got := card.lte(otherCard)
	expected := false
	if got != expected {
		t.Errorf("TestLteFalse expected: %v, got: %v", expected, got)
	}
}

func TestEqual(t *testing.T) {
	card := &Card{"K", "♦"}
	otherCard := &Card{"K", "♦"}
	got := card.equals(otherCard)
	expected := true
	if got != expected {
		t.Errorf("TestEqual expected: %v, got: %v", expected, got)
	}
}

func TestNotEqualByValue(t *testing.T) {
	card := &Card{"K", "♦"}
	otherCard := &Card{"Q", "♦"}
	got := card.equals(otherCard)
	expected := false
	if got != expected {
		t.Errorf("TestNotEqualByValue expected: %v, got: %v", expected, got)
	}
}

func TestNotEqualBySuit(t *testing.T) {
	card := &Card{"K", "♦"}
	otherCard := &Card{"K", "♠"}
	got := card.equals(otherCard)
	expected := false
	if got != expected {
		t.Errorf("TestNotEqualBySuit expected: %v, got: %v", expected, got)
	}
}
