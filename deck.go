package main

import (
	"errors"
	"math/rand"
	"strings"
	"time"
)

// Deck represents the deck of cars - talon
type Deck struct {
	cards []*Card
}

func newDeck() *Deck {
	rand.Seed(time.Now().Unix())
	cards := make([]*Card, 36)
	i := 0
	for _, v := range cardValues {
		for _, s := range cardSuits {
			cards[i] = &Card{Value: v, Suit: s}
			i = i + 1
		}
	}
	return &Deck{
		cards: cards,
	}
}

func (p *Deck) shuffle() {
	for i := range p.cards {
		j := rand.Intn(i + 1)
		p.cards[i], p.cards[j] = p.cards[j], p.cards[i]
	}
}

func (p *Deck) getCard() (*Card, error) {
	n := len(p.cards)
	if n > 0 {
		card := p.cards[n-1]
		// delete
		p.cards = append(p.cards[:n-1], p.cards[n:]...)
		return card, nil
	}
	return &Card{}, errors.New("no cards left in deck")
}

func (p *Deck) asString() string {
	cardsStrings := []string{}
	for _, card := range p.cards {
		cardsStrings = append(cardsStrings, card.Value+card.Suit)
	}
	return strings.Join(cardsStrings, " ")
}
