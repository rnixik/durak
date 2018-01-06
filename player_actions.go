package main

type PlayerAction struct {
	Name   string      `json:"name"`
	Data   interface{} `json:"data"`
	player *Player
}

type AttackActionData struct {
	Card        *Card `json:"card"`
	TargetIndex int   `json:"target_index"`
}
