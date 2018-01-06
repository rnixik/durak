package main

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
	thisIndex := c.getValueIndex()
	otherIndex := otherCard.getValueIndex()
	return thisIndex > otherIndex
}

func (c *Card) gte(otherCard *Card) bool {
	thisIndex := c.getValueIndex()
	otherIndex := otherCard.getValueIndex()
	return thisIndex >= otherIndex
}

func (c *Card) lt(otherCard *Card) bool {
	thisIndex := c.getValueIndex()
	otherIndex := otherCard.getValueIndex()
	return thisIndex < otherIndex
}

func (c *Card) lte(otherCard *Card) bool {
	thisIndex := c.getValueIndex()
	otherIndex := otherCard.getValueIndex()
	return thisIndex <= otherIndex
}
