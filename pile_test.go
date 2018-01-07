package main

import "testing"

func TestNewPile(t *testing.T) {
	pile := newPile()
	got := len(pile.cards)
	expected := 36
	if got != expected {
		t.Errorf("TestNewPile expected: %v, got: %v", expected, got)
	}
}

func TestShuffle(t *testing.T) {
	pile := newPile()
	card1 := pile.cards[0]
	card2 := pile.cards[1]
	card3 := pile.cards[2]
	pile.shuffle()
	scard1 := pile.cards[0]
	scard2 := pile.cards[1]
	scard3 := pile.cards[2]
	if card1.equals(scard1) && card2.equals(scard2) && card3.equals(scard3) {
		t.Errorf("TestShuffle: it was expected that at least one of first 3 cards was moved somewhere")
		t.Logf("Pile: %s", pile.asString())
	}
}

func TestGetCard(t *testing.T) {
	pile := newPile()
	card, err := pile.getCard()
	if err != nil {
		t.Errorf("TestGetCard got error: %s", err)
		return
	}
	if len(card.Value) == 0 {
		t.Errorf("TestGetCard got empty value of card")
	}
	if len(card.Suit) == 0 {
		t.Errorf("TestGetCard got empty suit of card")
	}
}

func TestGetCardOnEmptyPile(t *testing.T) {
	pile := newPile()
	pile.cards = pile.cards[:0]
	_, err := pile.getCard()
	if err == nil {
		t.Errorf("TestGetCard must be error")
	}
}

func TestAsString(t *testing.T) {
	pile := newPile()
	pile.cards = pile.cards[:4]
	got := pile.asString()
	expected := "6♣ 6♦ 6♥ 6♠"
	if got != expected {
		t.Errorf("TestAsString expected: %v, got: %v", expected, got)
	}
}
