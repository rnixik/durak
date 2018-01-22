package main

// Room represents place where some of clients want to start a new game.
type Room struct {
	id      uint64
	owner   *Client
	clients map[*Client]bool
	game    *Game
}

func newRoom(roomId uint64, owner *Client) *Room {
	clients := make(map[*Client]bool, 0)
	clients[owner] = true
	room := &Room{roomId, owner, clients, nil}
	owner.room = room

	roomJoinedEvent := RoomJoinedEvent{room.toRoomInfo()}
	owner.sendEvent(roomJoinedEvent)

	return room
}

// Name returns name of the room by its owner.
func (r *Room) Name() string {
	return r.owner.Nickname()
}

// Id returns id of the room
func (r *Room) Id() uint64 {
	return r.id
}

func (r *Room) removeClient(client *Client) (changedOwner bool, roomBecameEmpty bool) {
	client.room = nil
	delete(r.clients, client)

	roomUpdatedEvent := &RoomUpdatedEvent{r.toRoomInfo()}
	r.broadcastEvent(roomUpdatedEvent, nil)

	if len(r.clients) == 0 {
		roomBecameEmpty = true
		return
	}
	if r.owner == client {
		for ic := range r.clients {
			r.owner = ic
			changedOwner = true
			return
		}
	}
	return
}

func (r *Room) addClient(client *Client) {
	r.clients[client] = true
	client.room = r

	roomUpdatedEvent := &RoomUpdatedEvent{r.toRoomInfo()}
	r.broadcastEvent(roomUpdatedEvent, client)

	roomJoinedEvent := RoomJoinedEvent{r.toRoomInfo()}
	client.sendEvent(roomJoinedEvent)
}

func (r *Room) broadcastEvent(event interface{}, exceptClient *Client) {
	json, _ := eventToJSON(event)
	for c := range r.clients {
		if exceptClient == nil || c.id != exceptClient.id {
			c.sendMessage(json)
		}
	}
}

func (r *Room) toRoomInList() *RoomInList {
	gameStatus := ""
	if r.game != nil {
		gameStatus = r.game.status
	}
	roomInList := &RoomInList{
		Id:         r.Id(),
		OwnerId:    r.owner.id,
		Name:       r.Name(),
		GameStatus: gameStatus,
		ClientsNum: len(r.clients),
	}
	return roomInList
}

func (r *Room) toRoomInfo() *RoomInfo {
	gameStatus := ""
	if r.game != nil {
		gameStatus = r.game.status
	}

	clientsInList := make([]*ClientInList, 0)
	for client := range r.clients {
		clientInList := &ClientInList{
			Id:       client.id,
			Nickname: client.Nickname(),
		}
		clientsInList = append(clientsInList, clientInList)
	}

	roomInfo := &RoomInfo{
		Id:         r.Id(),
		OwnerId:    r.owner.id,
		Name:       r.Name(),
		GameStatus: gameStatus,
		Clients:    clientsInList,
	}
	return roomInfo
}
