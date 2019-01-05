package main

import "testing"

func TestNewDeck(t *testing.T) {
	deck := newDeck()
	got := len(deck.cards)
	expected := 36
	if got != expected {
		t.Errorf("TestNewDeck expected: %v, got: %v", expected, got)
	}
}

func TestShuffle(t *testing.T) {
	deck := newDeck()
	card1 := deck.cards[0]
	card2 := deck.cards[1]
	card3 := deck.cards[2]
	deck.shuffle()
	scard1 := deck.cards[0]
	scard2 := deck.cards[1]
	scard3 := deck.cards[2]
	if card1.equals(scard1) && card2.equals(scard2) && card3.equals(scard3) {
		t.Errorf("TestShuffle: it was expected that at least one of first 3 cards was moved somewhere")
		t.Logf("Deck: %s", deck.asString())
	}
}

func TestGetCard(t *testing.T) {
	deck := newDeck()
	card, err := deck.getCard()
	if err != nil {
		t.Fatalf("TestGetCard got error: %s", err)
	}
	if len(card.Value) == 0 {
		t.Errorf("TestGetCard got empty value of card")
	}
	if len(card.Suit) == 0 {
		t.Errorf("TestGetCard got empty suit of card")
	}
}

func TestGetCardOnEmptyDeck(t *testing.T) {
	deck := newDeck()
	deck.cards = deck.cards[:0]
	_, err := deck.getCard()
	if err == nil {
		t.Errorf("TestGetCard must be error")
	}
}

func TestAsString(t *testing.T) {
	deck := newDeck()
	deck.cards = deck.cards[:4]
	got := deck.asString()
	expected := "6♣ 6♦ 6♥ 6♠"
	if got != expected {
		t.Errorf("TestAsString expected: %v, got: %v", expected, got)
	}
}
