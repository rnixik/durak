package main

type RoomMemberInfo struct {
	Id         uint64 `json:"id"`
	Nickname   string `json:"nickname"`
	WantToPlay bool   `json:"want_to_play"`
}

// RoomInfo contains info about room where client is.
type RoomInfo struct {
	Id         uint64            `json:"id"`
	OwnerId    uint64            `json:"owner_id"`
	Name       string            `json:"name"`
	GameStatus string            `json:"game_status"`
	Members    []*RoomMemberInfo `json:"members"`
}

// RoomJoinedEvent contains info about room where client is
type RoomJoinedEvent struct {
	Room *RoomInfo `json:"room"`
}

// RoomUpdatedEvent contains info about updated room where client is
type RoomUpdatedEvent struct {
	Room *RoomInfo `json:"room"`
}
