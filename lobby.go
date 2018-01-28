package main

import (
	"fmt"
	"log"
	"strconv"
	"sync/atomic"
)

var lastGameId uint64
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
}

func newLobby() *Lobby {
	return &Lobby{
		broadcast:      make(chan []byte),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		clients:        make(map[*Client]bool),
		clientCommands: make(chan *ClientCommand),
		games:          make([]*Game, 0),
		rooms:          make(map[*Client]*Room),
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
			if _, ok := l.clients[client]; ok {
				l.onClientLeft(client)
				client.isValid = false
				delete(l.clients, client)
				close(client.send)
			}
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
		case clientCommand := <-l.clientCommands:
			l.onClientCommand(clientCommand)
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
		// TODO: add check whether client is in a room or in a game
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
		errEvent := &ClientCommandError{"You can create 1 room only"}
		c.sendEvent(errEvent)
		return
	}

	oldRoomJoined := c.room
	if oldRoomJoined != nil {
		l.onLeftRoom(c, oldRoomJoined)
	}

	atomic.AddUint64(&lastRoomId, 1)
	lastRoomIdSafe := atomic.LoadUint64(&lastRoomId)

	room := newRoom(lastRoomIdSafe, c)
	l.rooms[c] = room

	event := &ClientCreatedRoomEvent{room.toRoomInList()}
	l.broadcastEvent(event)
}

func (l *Lobby) getRoomById(roomId uint64) (room *Room, err error) {
	for _, r := range l.rooms {
		if r.Id() == roomId {
			return r, nil
		}
	}
	return nil, fmt.Errorf("Room not found by id = %d", roomId)
}

func (l *Lobby) onLeftRoom(c *Client, room *Room) {
	changedOwner, roomBecameEmpty := room.removeClient(c)
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
		errEvent := &ClientCommandError{fmt.Sprintf("Room does not exists: %d", roomId)}
		c.sendEvent(errEvent)
	}
}

func (l *Lobby) onClientCommand(cc *ClientCommand) {
	if cc.Type == "lobby" {
		if cc.SubType == "join" {
			nickname, ok := cc.Data.(string)
			if ok {
				l.onJoinCommand(cc.client, nickname)
				//l.createNewGame(cc.client)
			}
		} else if cc.SubType == "create_room" {
			l.onCreateNewRoomCommand(cc.client)
		} else if cc.SubType == "join_room" {
			roomIdStr := fmt.Sprintf("%v", cc.Data)
			roomId, err := strconv.ParseUint(roomIdStr, 10, 64)
			if err == nil {
				l.onJoinRoomCommand(cc.client, roomId)
			}
		}
	} else if cc.Type == "game" {
		// demo
		actionData := AttackActionData{
			Card:        &Card{"6", "♦"},
			TargetIndex: 1,
		}
		playerAction := &PlayerAction{Name: "attack", Data: actionData, player: l.games[0].players[0]}
		l.games[0].playerActions <- playerAction
	} else if cc.Type == "room" {
		if cc.client.room == nil {
			return
		}
		cc.client.room.onClientCommand(cc)
	}
}
