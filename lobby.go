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

	// Rooms created by clients
	rooms map[*Client]*Room
}

func newLobby() *Lobby {
	return &Lobby{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		games:      make([]*Game, 0),
		rooms:      make(map[*Client]*Room),
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
			client.isValid = true
			l.clients[client] = true
		case client := <-l.unregister:
			leftId := client.id
			if _, ok := l.clients[client]; ok {
				client.isValid = false
				delete(l.clients, client)
				close(client.send)
			}
			l.onLeft(leftId)
		case message := <-l.broadcast:
			for client := range l.clients {
				select {
				case client.send <- message:
				default:
					client.isValid = false
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

func (l *Lobby) onJoinCommand(c *Client, nickname string) {
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

	roomsInList := make([]*RoomInList, 0)
	for _, room := range l.rooms {
		roomInList := room.toRoomInList()
		roomsInList = append(roomsInList, roomInList)
	}

	event := &ClientJoinedEvent{
		YourId:       c.id,
		YourNickname: nickname,
		Clients:      clientsInList,
		Rooms:        roomsInList,
	}
	c.sendEvent(event)
}

func (l *Lobby) onLeft(leftId uint64) {
	leftEvent := &ClientLeftEvent{
		Id: leftId,
	}
	l.broadcastEvent(leftEvent)
}

func (l *Lobby) onCreateNewRoomCommand(c *Client) {
	_, roomExists := l.rooms[c]
	if roomExists {
		errEvent := &ClientCommandError{"You can create 1 room only"}
		c.sendEvent(errEvent)
		return
	}
	room := newRoom(c)
	l.rooms[c] = room
	roomInList := room.toRoomInList()
	event := &ClientCreatedRoomEvent{roomInList}
	l.broadcastEvent(event)
}

func (l *Lobby) onClientCommand(cc *ClientCommand) {
	if cc.Type == "lobby" {
		if cc.SubType == "join" {
			nickname, ok := cc.Data.(string)
			if ok {
				l.onJoinCommand(cc.client, nickname)
				l.createNewGame(cc.client)
			}
		} else if cc.SubType == "create_room" {
			l.onCreateNewRoomCommand(cc.client)
		}
	} else if cc.Type == "game" {
		// demo
		actionData := AttackActionData{
			Card:        &Card{"6", "♦"},
			TargetIndex: 1,
		}
		playerAction := &PlayerAction{Name: "attack", Data: actionData, player: l.games[0].players[0]}
		l.games[0].playerActions <- playerAction
	}
}
