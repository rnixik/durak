package main

// PlayerAction contains command message from a player to a game.
type PlayerAction struct {
	Name   string      `json:"name"`
	Data   interface{} `json:"data"`
	player *Player
}

// UseCardActionData contains data of command message to use card from a player to a game.
type UseCardActionData struct {
	Card *Card `json:"card"`
}
