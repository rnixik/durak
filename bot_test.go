package main

import "testing"

func TestFindLowestCard1(t *testing.T) {
	bot := &Bot{gameStateInfo: &GameStateInfo{TrumpCard: &Card{"9", "♦"}}}

	cards := []*Card{
		&Card{"7", "♥"},
		&Card{"6", "♦"},
	}

	got := bot.findLowestCard(cards)
	expected := &Card{"7", "♥"}
	if !got.equals(expected) {
		t.Errorf("findLowestCard expected: %v, got: %v", expected, got)
	}
}

func TestFindLowestCard2(t *testing.T) {
	bot := &Bot{gameStateInfo: &GameStateInfo{TrumpCard: &Card{"9", "♦"}}}

	cards := []*Card{
		&Card{"7", "♥"},
		&Card{"6", "♥"},
	}

	got := bot.findLowestCard(cards)
	expected := &Card{"6", "♥"}
	if !got.equals(expected) {
		t.Errorf("findLowestCard expected: %v, got: %v", expected, got)
	}
}

func TestFindLowestCard3(t *testing.T) {
	bot := &Bot{gameStateInfo: &GameStateInfo{TrumpCard: &Card{"9", "♦"}}}

	cards := []*Card{
		&Card{"Q", "♥"},
		&Card{"A", "♥"},
	}

	got := bot.findLowestCard(cards)
	expected := &Card{"Q", "♥"}
	if !got.equals(expected) {
		t.Errorf("findLowestCard expected: %v, got: %v", expected, got)
	}
}

func TestFindLowestCard4(t *testing.T) {
	bot := &Bot{gameStateInfo: &GameStateInfo{TrumpCard: &Card{"9", "♦"}}}

	cards := []*Card{
		&Card{"K", "♦"},
		&Card{"Q", "♦"},
		&Card{"A", "♦"},
	}

	got := bot.findLowestCard(cards)
	expected := &Card{"Q", "♦"}
	if !got.equals(expected) {
		t.Errorf("findLowestCard expected: %v, got: %v", expected, got)
	}
}
