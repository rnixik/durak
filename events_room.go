package main

// RoomInfo contains info about room where client is.
type RoomInfo struct {
	Id         uint64          `json:"id"`
	OwnerId    uint64          `json:"owner_id"`
	Name       string          `json:"name"`
	GameStatus string          `json:"game_status"`
	Clients    []*ClientInList `json:"clients"`
}

// RoomJoinedEvent contains info about room where client is
type RoomJoinedEvent struct {
	Room *RoomInfo `json:"room"`
}

// RoomUpdatedEvent contains info about updated room where client is
type RoomUpdatedEvent struct {
	Room *RoomInfo `json:"room"`
}
