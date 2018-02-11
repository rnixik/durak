package main

// GamePlayersEvent contains list of players which were connected to a game.
type GamePlayersEvent struct {
	YourPlayerIndex int       `json:"your_player_index"`
	Players         []*Player `json:"players"`
}

// GameDealEvent contains info about game after the deal. It includes list of cards for each player.
type GameDealEvent struct {
	YourHand                      []*Card `json:"your_hand"`
	HandsSizes                    []int   `json:"hands_sizes"`
	PileSize                      int     `json:"pile_size"`
	TrumpSuit                     string  `json:"trump_suit"`
	TrumpCard                     *Card   `json:"trump_card"`
	TrumpCardIsInPile             bool    `json:"trump_card_is_in_pile"`
	TrumpCardIsOwnedByPlayerIndex int     `json:"trump_card_is_owned_by_player_index"`
}

// GameFirstAttackerEvent contains info who is the first attacker and why.
type GameFirstAttackerEvent struct {
	ReasonCard    *Card `json:"reason_card"`
	AttackerIndex int   `json:"attacker_index"`
}
