package main

// Room represents place where some of clients want to start a new game.
type Room struct {
	owner   *Client
	clients []*Client
	game    *Game
}

func newRoom(owner *Client) *Room {
	return &Room{owner, make([]*Client, 0), nil}
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
	}
	return roomInList
}
