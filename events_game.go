package main

// GamePlayersEvent contains list of players which were connected to a game.
type GamePlayersEvent struct {
	YourPlayerIndex int       `json:"your_player_index"`
	Players         []*Player `json:"players"`
}

// GameStateInfo contains info about card for each player, cards in deck, card on battleground.
type GameStateInfo struct {
	YourHand     []*Card `json:"your_hand"`
	HandsSizes   []int   `json:"hands_sizes"`
	PileSize     int     `json:"pile_size"`
	Battleground []*Card `json:"battleground"`
}

// GameDealEvent contains info about game after the deal. It includes list of cards for each player.
type GameDealEvent struct {
	GameStateInfo                 *GameStateInfo `json:"game_state_info"`
	TrumpSuit                     string         `json:"trump_suit"`
	TrumpCard                     *Card          `json:"trump_card"`
	TrumpCardIsInPile             bool           `json:"trump_card_is_in_pile"`
	TrumpCardIsOwnedByPlayerIndex int            `json:"trump_card_is_owned_by_player_index"`
}

// GameFirstAttackerEvent contains info who is the first attacker and why.
type GameFirstAttackerEvent struct {
	ReasonCard    *Card `json:"reason_card"`
	AttackerIndex int   `json:"attacker_index"`
	DefenderIndex int   `json:"defender_index"`
}

// GamePlayerLeftEvent contains index of player who left the game
type GamePlayerLeftEvent struct {
	PlayerIndex int `json:"player_index"`
}

// GameEndEvent contains index of player who left the game
type GameEndEvent struct {
	WinnerIndex int `json:"winner_index"`
}

// GameAttackEvent contains info about attack with card
type GameAttackEvent struct {
	GameStateInfo *GameStateInfo `json:"game_state_info"`
	AttackerIndex int            `json:"attacker_index"`
	DefenderIndex int            `json:"defender_index"`
	Card          *Card          `json:"card"`
}
