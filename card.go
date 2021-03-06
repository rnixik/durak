package main

var (
	cardValues = []string{"6", "7", "8", "9", "10", "J", "Q", "K", "A"}
	cardSuits  = []string{"♣", "♦", "♥", "♠"}
)

// Card represents a card from French playing cards with Value And Suit.
type Card struct {
	Value string `json:"value"`
	Suit  string `json:"suit"`
}

func (c *Card) getValueIndex() int {
	index := -1
	for i, v := range cardValues {
		if v == c.Value {
			index = i
		}
	}
	return index
}

func (c *Card) gt(otherCard *Card) bool {
	if c.Suit != otherCard.Suit {
		return false
	}
	thisIndex := c.getValueIndex()
	otherIndex := otherCard.getValueIndex()
	return thisIndex > otherIndex
}

func (c *Card) gte(otherCard *Card) bool {
	if c.Suit != otherCard.Suit {
		return false
	}
	thisIndex := c.getValueIndex()
	otherIndex := otherCard.getValueIndex()
	return thisIndex >= otherIndex
}

func (c *Card) lt(otherCard *Card) bool {
	if c.Suit != otherCard.Suit {
		return false
	}
	thisIndex := c.getValueIndex()
	otherIndex := otherCard.getValueIndex()
	return thisIndex < otherIndex
}

func (c *Card) lte(otherCard *Card) bool {
	if c.Suit != otherCard.Suit {
		return false
	}
	thisIndex := c.getValueIndex()
	otherIndex := otherCard.getValueIndex()
	return thisIndex <= otherIndex
}

func (c *Card) equals(otherCard *Card) bool {
	return c.Value == otherCard.Value && c.Suit == otherCard.Suit
}
