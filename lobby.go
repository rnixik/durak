package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sync/atomic"
)

var lastClientId uint64
var lastRoomId uint64

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

	// Commands from clients
	clientCommands chan *ClientCommand

	// Started games
	games []*Game

	// Rooms created by clients
	rooms map[*Client]*Room

	gameLogger GameLogger
}

func newLobby(gameLogger GameLogger) *Lobby {
	return &Lobby{
		broadcast:      make(chan []byte),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		clients:        make(map[*Client]bool),
		clientCommands: make(chan *ClientCommand),
		games:          make([]*Game, 0),
		rooms:          make(map[*Client]*Room),
		gameLogger:     gameLogger,
	}
}

func (l *Lobby) run() {
	log.Println("Go lobby")

	go func() {
		for {
			select {
			case message, ok := <-l.broadcast:
				if !ok {
					continue
				}
				for client := range l.clients {
					client.sendMessage(message)
				}
			}
		}
	}()

	for {
		select {
		case client := <-l.register:
			atomic.AddUint64(&lastClientId, 1)
			lastClientIdSafe := atomic.LoadUint64(&lastClientId)
			client.id = lastClientIdSafe
			client.isValid = true
			l.clients[client] = true
		case client := <-l.unregister:
			if _, ok := l.clients[client]; ok {
				l.onClientLeft(client)
				client.isValid = false
				delete(l.clients, client)
				close(client.send)
			}
		case clientCommand := <-l.clientCommands:
			l.onClientCommand(clientCommand)
		}
	}
}

func (l *Lobby) broadcastEvent(event interface{}) {
	jsonData, _ := eventToJSON(event)
	l.broadcast <- jsonData
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

func (l *Lobby) onClientLeft(client *Client) {
	room := client.room
	if room != nil {
		l.onLeftRoom(client, room)
	}
	leftEvent := &ClientLeftEvent{
		Id: client.id,
	}
	l.broadcastEvent(leftEvent)
}

func (l *Lobby) onCreateNewRoomCommand(c *Client) {
	_, roomExists := l.rooms[c]
	if roomExists {
		errEvent := &ClientCommandError{errorYouCanCreateOneRoomOnly}
		c.sendEvent(errEvent)
		return
	}

	oldRoomJoined := c.room
	if oldRoomJoined != nil {
		l.onLeftRoom(c, oldRoomJoined)
	}

	atomic.AddUint64(&lastRoomId, 1)
	lastRoomIdSafe := atomic.LoadUint64(&lastRoomId)

	room := newRoom(lastRoomIdSafe, c, l)
	l.rooms[c] = room

	event := &ClientCreatedRoomEvent{room.toRoomInList()}
	l.broadcastEvent(event)

	roomJoinedEvent := RoomJoinedEvent{room.toRoomInfo()}
	c.sendEvent(roomJoinedEvent)
}

func (l *Lobby) getRoomById(roomId uint64) (room *Room, err error) {
	for _, r := range l.rooms {
		if r.Id() == roomId {
			return r, nil
		}
	}
	return nil, fmt.Errorf("room not found by id = %d", roomId)
}

func (l *Lobby) onLeftRoom(c *Client, room *Room) {
	changedOwner, roomBecameEmpty := room.removeClient(c)
	c.room = nil
	if roomBecameEmpty {
		roomInListRemovedEvent := &RoomInListRemovedEvent{room.Id()}
		l.broadcastEvent(roomInListRemovedEvent)
		l.rooms[c] = nil
		delete(l.rooms, c)
		return
	}
	if changedOwner {
		roomOwnerClient, ok := room.owner.client.(*Client)
		if !ok {
			return
		}
		l.rooms[roomOwnerClient] = room
		delete(l.rooms, c)
	}
	roomInListUpdatedEvent := &RoomInListUpdatedEvent{room.toRoomInList()}
	l.broadcastEvent(roomInListUpdatedEvent)
}

func (l *Lobby) onJoinRoomCommand(c *Client, roomId uint64) {
	log.Printf("OnJoinRoomCommand: %s, %d", c.Nickname(), roomId)
	oldRoomJoined := c.room
	if oldRoomJoined != nil && oldRoomJoined.Id() == roomId {
		return
	}
	if oldRoomJoined != nil {
		l.onLeftRoom(c, oldRoomJoined)
	}
	room, err := l.getRoomById(roomId)
	if err == nil {
		room.addClient(c)
		roomInListUpdatedEvent := &RoomInListUpdatedEvent{room.toRoomInList()}
		l.broadcastEvent(roomInListUpdatedEvent)
		log.Printf("Client %s joined room %d", c.Nickname(), roomId)
	} else {
		errEvent := &ClientCommandError{errorRoomDoesNotExist}
		c.sendEvent(errEvent)
	}
}

func (l *Lobby) onClientCommand(cc *ClientCommand) {
	if cc.Type == ClientCommandTypeLobby {
		if cc.SubType == ClientCommandLobbySubTypeJoin {
			var nickname string
			if err := json.Unmarshal(cc.Data, &nickname); err != nil {
				return
			}
			l.onJoinCommand(cc.client, nickname)
		} else if cc.SubType == ClientCommandLobbySubTypeCreateRoom {
			l.onCreateNewRoomCommand(cc.client)
		} else if cc.SubType == ClientCommandLobbySubTypeJoinRoom {
			var roomId uint64
			if err := json.Unmarshal(cc.Data, &roomId); err != nil {
				return
			}
			l.onJoinRoomCommand(cc.client, roomId)
		}
	} else if cc.Type == ClientCommandTypeGame {
		l.dispatchGameEvent(cc)
	} else if cc.Type == ClientCommandTypeRoom {
		if cc.client.room == nil {
			return
		}
		cc.client.room.onClientCommand(cc)
	}
}

func (l *Lobby) dispatchGameEvent(cc *ClientCommand) {
	if cc.client.room == nil {
		return
	}
	if cc.client.room.game == nil {
		return
	}
	game := cc.client.room.game
	if game.status != GameStatusPlaying {
		return
	}

	player := game.findPlayerOfClient(cc.client)
	if player == nil {
		return
	}

	var playerAction *PlayerAction

	if cc.SubType == ClientCommandGameSubTypeAttack {
		var attackActionData AttackActionData
		if err := json.Unmarshal(cc.Data, &attackActionData); err != nil {
			return
		}
		playerAction = &PlayerAction{Name: PlayerActionNameAttack, Data: attackActionData, player: player}
	} else if cc.SubType == ClientCommandGameSubTypeDefend {
		var defendActionData DefendActionData
		if err := json.Unmarshal(cc.Data, &defendActionData); err != nil {
			return
		}
		playerAction = &PlayerAction{Name: PlayerActionNameDefend, Data: defendActionData, player: player}
	} else if cc.SubType == ClientCommandGameSubTypePickUp {
		playerAction = &PlayerAction{Name: PlayerActionNamePickUp, player: player}
	} else if cc.SubType == ClientCommandGameSubTypeComplete {
		playerAction = &PlayerAction{Name: PlayerActionNameComplete, player: player}
	}

	if playerAction != nil {
		game.playerActions <- playerAction
	}
}

func (l *Lobby) sendRoomUpdate(room *Room) {
	roomInListUpdatedEvent := &RoomInListUpdatedEvent{room.toRoomInList()}
	l.broadcastEvent(roomInListUpdatedEvent)
}
