package main

// Player represents connected to a game client which can have cards.
type Player struct {
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
	client   ClientSender
	cards    []*Card
}

func (p *Player) sendEvent(event interface{}) {
	p.client.sendEvent(event)
}

func (p *Player) sendMessage(message []byte) {
	p.client.sendMessage(message)
}

func newPlayer(client ClientSender, isActive bool) *Player {
	return &Player{
		Name:     client.Nickname(),
		IsActive: isActive,
		client:   client,
	}
}

func (p *Player) removeCard(card *Card) {
	cards := make([]*Card, 0)
	for _, c := range p.cards {
		if !c.equals(card) {
			cards = append(cards, c)
		}
	}
	p.cards = cards
}

func (p *Player) hasCard(card *Card) bool {
	for _, c := range p.cards {
		if c.equals(card) {
			return true
		}
	}
	return false
}
