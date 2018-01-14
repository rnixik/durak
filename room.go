package main

// Room represents place where some of clients want to start a new game.
type Room struct {
	owner   *Client
	clients map[*Client]bool
	game    *Game
}

func newRoom(owner *Client) *Room {
	clients := make(map[*Client]bool, 0)
	clients[owner] = true
	return &Room{owner, clients, nil}
}

// Name returns name of the room by its owner.
func (r *Room) Name() string {
	return r.owner.Nickname()
}

// Id returns id of the room
func (r *Room) Id() uint64 {
	// Just use id of onwer because one client can't have more than 1 room
	return r.owner.id
}

func (r *Room) removeClient(client *Client) (changedOwner bool, roomBecameEmpty bool) {
	client.room = nil
	delete(r.clients, client)
	if len(r.clients) == 0 {
		roomBecameEmpty = true
		return
	}
	if r.owner == client {
		for ic, _ := range r.clients {
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
