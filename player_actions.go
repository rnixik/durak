package main

const PlayerActionNameAttack = "attack"

// PlayerAction contains command message from a player to a game.
type PlayerAction struct {
	Name   string      `json:"name"`
	Data   interface{} `json:"data"`
	player *Player
}

// AttackActionData contains data of command message to attack with card from a player to a game.
type AttackActionData struct {
	Card *Card `json:"card"`
}
