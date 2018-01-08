package main

// PlayerAction contains command message from a player to a game.
type PlayerAction struct {
	Name   string      `json:"name"`
	Data   interface{} `json:"data"`
	player *Player
}

// AttackActionData contains data of command message from a player to a game.
type AttackActionData struct {
	Card        *Card `json:"card"`
	TargetIndex int   `json:"target_index"`
}
