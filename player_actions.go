package main

const PlayerActionNameAttack = "attack"
const PlayerActionNameDefend = "defend"
const PlayerActionNamePickUp = "pick_up"
const PlayerActionNameComplete = "complete"

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

// DefendActionData contains data of command message to defend a card with card from a player to a game.
type DefendActionData struct {
	AttackingCard *Card `json:"attackingCard"`
	DefendingCard *Card `json:"defendingCard"`
}
