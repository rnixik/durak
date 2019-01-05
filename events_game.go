package main

// GamePlayersEvent contains list of players which were connected to a game.
type GamePlayersEvent struct {
	YourPlayerIndex int       `json:"yourPlayerIndex"`
	Players         []*Player `json:"players"`
}

// GameStateInfo contains info about card for each player, cards in deck, card on battleground.
type GameStateInfo struct {
	YourHand         []*Card       `json:"yourHand"`
	CanYouPickUp     bool          `json:"canYouPickUp"`
	CanYouComplete   bool          `json:"canYouComplete"`
	HandsSizes       []int         `json:"handsSizes"`
	DeckSize         int           `json:"deckSize"`
	DiscardPileSize  int           `json:"discardPileSize"`
	TrumpCard        *Card         `json:"trumpCard"`
	Battleground     []*Card       `json:"battleground"`
	DefendingCards   map[int]*Card `json:"defendingCards"`
	CompletedPlayers map[int]bool  `json:"completedPlayers"`
	DefenderPickUp   bool          `json:"defenderPickUp"`
	AttackerIndex    int           `json:"attackerIndex"`
	DefenderIndex    int           `json:"defenderIndex"`
}

// GameDealEvent contains info about game after the deal. It includes list of cards for each player.
type GameDealEvent struct {
	GameStateInfo                 *GameStateInfo `json:"gameStateInfo"`
	TrumpSuit                     string         `json:"trumpSuit"`
	TrumpCard                     *Card          `json:"trumpCard"`
	TrumpCardIsInDeck             bool           `json:"trumpCardIsInDeck"`
	TrumpCardIsOwnedByPlayerIndex int            `json:"trumpCardIsOwnedByPlayerIndex"`
}

// GameFirstAttackerEvent contains info who is the first attacker and why.
type GameFirstAttackerEvent struct {
	ReasonCard    *Card `json:"reasonCard"`
	AttackerIndex int   `json:"attackerIndex"`
	DefenderIndex int   `json:"defenderIndex"`
}

// GamePlayerLeftEvent contains index of player who left the game
type GamePlayerLeftEvent struct {
	PlayerIndex int `json:"playerIndex"`
}

// GameAttackEvent contains info about attack with card
type GameAttackEvent struct {
	GameStateInfo *GameStateInfo `json:"gameStateInfo"`
	AttackerIndex int            `json:"attackerIndex"`
	DefenderIndex int            `json:"defenderIndex"`
	Card          *Card          `json:"card"`
}

// GameDefendEvent contains info about defending card with other card
type GameDefendEvent struct {
	GameStateInfo *GameStateInfo `json:"gameStateInfo"`
	DefenderIndex int            `json:"defenderIndex"`
	AttackingCard *Card          `json:"attackingCard"`
	DefendingCard *Card          `json:"defendingCard"`
}

// GameStateEvent contains info about state only
type GameStateEvent struct {
	GameStateInfo *GameStateInfo `json:"gameStateInfo"`
}

// GameEndEvent contains info about winner
type GameEndEvent struct {
	HasLoser   bool `json:"hasLoser"`
	LoserIndex int  `json:"loserIndex"`
}
