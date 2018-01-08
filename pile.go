package main

import (
	"errors"
	"math/rand"
	"strings"
	"time"
)

// Pile represents the pile of cars - talon
type Pile struct {
	cards []*Card
}

func newPile() *Pile {
	rand.Seed(time.Now().Unix())
	cards := make([]*Card, 36)
	i := 0
	for _, v := range cardValues {
		for _, s := range cardSuits {
			cards[i] = &Card{Value: v, Suit: s}
			i = i + 1
		}
	}
	return &Pile{
		cards: cards,
	}
}

func (p *Pile) shuffle() {
	for i := range p.cards {
		j := rand.Intn(i + 1)
		p.cards[i], p.cards[j] = p.cards[j], p.cards[i]
	}
}

func (p *Pile) getCard() (*Card, error) {
	n := len(p.cards)
	if n > 0 {
		card := p.cards[n-1]
		// delete
		p.cards = append(p.cards[:n-1], p.cards[n:]...)
		return card, nil
	}
	return &Card{}, errors.New("No cards left in pile")
}

func (p *Pile) asString() string {
	cardsStrings := []string{}
	for _, card := range p.cards {
		cardsStrings = append(cardsStrings, card.Value+card.Suit)
	}
	return strings.Join(cardsStrings, " ")
}
