package main

type Player struct {
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
	client   *Client
	cards    []*Card
}

func (p *Player) sendEvent(event interface{}) {
	p.client.sendEvent(event)
}

func (p *Player) sendMessage(message []byte) {
	p.client.sendMessage(message)
}

func newPlayer(client *Client, isActive bool) *Player {
	return &Player{
		Name:     client.nickname,
		IsActive: isActive,
		client:   client,
	}
}
