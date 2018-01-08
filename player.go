package main

// ClientInterface represents interface which sends events to connected players.
type ClientInterface interface {
	sendEvent(event interface{})
	sendMessage(message []byte)
	Nickname() string
}

// Player represents connected to a game client which can have cards.
type Player struct {
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
	client   ClientInterface
	cards    []*Card
}

func (p *Player) sendEvent(event interface{}) {
	p.client.sendEvent(event)
}

func (p *Player) sendMessage(message []byte) {
	p.client.sendMessage(message)
}

func newPlayer(client ClientInterface, isActive bool) *Player {
	return &Player{
		Name:     client.Nickname(),
		IsActive: isActive,
		client:   client,
	}
}
