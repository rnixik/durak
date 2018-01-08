package main

import (
	"log"
	"sync/atomic"
)

var lastGameId uint64
var lastClientId uint64

// Lobby is the first place for connected clients. It passes commands to games.
type Lobby struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	// Started games
	games []*Game
}

func newLobby() *Lobby {
	return &Lobby{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		games:      make([]*Game, 0),
	}
}

func (l *Lobby) run() {
	log.Println("Go lobby")
	for {
		select {
		case client := <-l.register:
			atomic.AddUint64(&lastClientId, 1)
			lastClientIdSafe := atomic.LoadUint64(&lastClientId)
			client.id = lastClientIdSafe
			l.clients[client] = true
		case client := <-l.unregister:
			leftId := client.id
			if _, ok := l.clients[client]; ok {
				delete(l.clients, client)
				close(client.send)
			}
			l.onLeft(leftId)
		case message := <-l.broadcast:
			for client := range l.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(l.clients, client)
				}
			}
		}
	}
}

func (l *Lobby) createNewGame(ownerClient *Client) {
	players := make([]*Player, 3)
	players[0] = newPlayer(ownerClient, true)
	players[1] = newPlayer(&Client{nickname: "pl2"}, true)
	players[2] = newPlayer(&Client{nickname: "pl3"}, false)

	atomic.AddUint64(&lastGameId, 1)
	lastGameIdSafe := atomic.LoadUint64(&lastGameId)

	game := newGame(lastGameIdSafe, players[0], players)
	l.games = append(l.games, game)
	go game.begin()
}

func (l *Lobby) broadcastEvent(event interface{}) {
	json, _ := eventToJSON(event)
	for client := range l.clients {
		client.sendMessage(json)
	}
}

func (l *Lobby) onJoin(c *Client, nickname string) {
	c.nickname = nickname

	broadcastEvent := &ClientBroadCastJoinedEvent{
		Id:       c.id,
		Nickname: c.nickname,
	}
	l.broadcastEvent(broadcastEvent)

	clientsInList := make([]*ClientInList, 0)
	for client := range l.clients {
		clientInList := &ClientInList{
			Id:       client.id,
			Nickname: client.Nickname(),
		}
		clientsInList = append(clientsInList, clientInList)
	}

	gamesInList := make([]*GameInList, 0)
	for _, game := range l.games {
		if game.status != GameStatusFinished {
			gameInList := &GameInList{
				Id:   game.id,
				Name: game.getName(),
			}
			gamesInList = append(gamesInList, gameInList)
		}
	}

	event := &ClientJoinedEvent{
		YourId:       c.id,
		YourNickname: nickname,
		Clients:      clientsInList,
		Games:        gamesInList,
	}
	c.sendEvent(event)
}

func (l *Lobby) onLeft(leftId uint64) {
	leftEvent := &ClientLeftEvent{
		Id: leftId,
	}
	l.broadcastEvent(leftEvent)
}

func (l *Lobby) onClientMessage(m *ClientMessage) {
	if m.Type == "lobby" {
		if m.SubType == "join" {
			nickname, ok := m.Data.(string)
			if ok {
				l.onJoin(m.client, nickname)
				l.createNewGame(m.client)
			}
		}
	} else if m.Type == "game" {
		// demo
		actionData := AttackActionData{
			Card:        &Card{"6", "♦"},
			TargetIndex: 1,
		}
		playerAction := &PlayerAction{Name: "attack", Data: actionData, player: l.games[0].players[0]}
		l.games[0].playerActions <- playerAction
	}
}
