package main

import (
	"log"
)

// Statuses of the game.
const (
	GameStatusPreparing = "preparing"
	GameStatusPlaying   = "playing"
	GameStatusFinished  = "finished"
)

// States of the game.
const (
	GameStateDealing    = "dealing"
	GameStateAttacking  = "attacking"
	GameStateThrowingIn = "throwing_in"
)

var (
	cardValues = []string{"6", "7", "8", "9", "10", "J", "Q", "K", "A"}
	cardSuits  = []string{"♣", "♦", "♥", "♠"}
)

// CardOnDesk represents a game card which was played.
type CardOnDesk struct {
	Card         *Card `json:"card"`
	SourcePlayer int64 `json:"source_player"`
	BeatByCard   *Card `json:"beat_by_card"`
}

// Desk contains played cards.
type Desk struct {
	Cards []*CardOnDesk `json:"cards"`
}

// Game represents status, state, etc of the game.
type Game struct {
	playerActions chan *PlayerAction
	owner         *Player
	id            uint64
	room          *Room
	// playing, finished
	status  string
	players []*Player
	// attack, throw_in
	state                         string
	pile                          *Pile
	trumpSuit                     string
	trumpCard                     *Card
	trumpCardIsInPile             bool
	trumpCardIsOwnedByPlayerIndex int
	firstAttackerReasonCard       *Card
	attackerIndex                 int
	defenderIndex                 int
	battleground                  []*Card
}

func newGame(id uint64, room *Room, players []*Player) *Game {
	return &Game{
		id:            id,
		room:          room,
		playerActions: make(chan *PlayerAction),
		status:        GameStatusPreparing,
		players:       players,
	}
}

func (g *Game) sendPlayersEvent() {
	pe := GamePlayersEvent{
		YourPlayerIndex: -1,
		Players:         g.players,
	}
	for i, p := range g.players {
		pe.YourPlayerIndex = i
		p.sendEvent(pe)
	}
	log.Println("end of pl event")
}

func (g *Game) deal() {
	cardsLimit := 6
	var lastCard *Card
	lastPlayerIndex := -1
	for cardIndex := 0; cardIndex < cardsLimit; cardIndex = cardIndex + 1 {
		for playerIndex, p := range g.players {
			if len(p.cards) >= cardsLimit {
				break
			}
			card, err := g.pile.getCard()
			if err == nil {
				p.cards = append(p.cards, card)
				lastCard = card
				lastPlayerIndex = playerIndex
			}
		}
	}
	if len(g.pile.cards) > 0 {
		lastCard = g.pile.cards[0]
		g.trumpCardIsInPile = true
		g.trumpCardIsOwnedByPlayerIndex = -1
	} else {
		g.trumpCardIsInPile = false
		g.trumpCardIsOwnedByPlayerIndex = lastPlayerIndex
	}

	g.trumpCard = lastCard
	g.trumpSuit = lastCard.Suit
}

func (g *Game) getGameStateInfo() *GameStateInfo {
	handSizes := make([]int, len(g.players))
	for i, p := range g.players {
		handSizes[i] = len(p.cards)
	}

	gsi := &GameStateInfo{
		YourHand:     make([]*Card, 0),
		HandsSizes:   handSizes,
		PileSize:     len(g.pile.cards),
		Battleground: g.battleground,
	}

	return gsi
}

func (g *Game) sendDealEvent() {
	gsi := g.getGameStateInfo()

	de := GameDealEvent{
		GameStateInfo:                 gsi,
		TrumpSuit:                     g.trumpSuit,
		TrumpCard:                     g.trumpCard,
		TrumpCardIsInPile:             g.trumpCardIsInPile,
		TrumpCardIsOwnedByPlayerIndex: g.trumpCardIsOwnedByPlayerIndex,
	}

	for _, p := range g.players {
		de.GameStateInfo.YourHand = p.cards
		p.sendEvent(de)
	}
}

func (g *Game) dumpHands() {
	for i, p := range g.players {
		log.Printf("player #%d has cards: %d\n", i, len(p.cards))
		str := ""
		for _, card := range p.cards {
			str += card.Value + card.Suit + " "
		}
		log.Printf("%s", str)
		log.Println("---")
	}
}

func (g *Game) dumpPile() {
	log.Printf("Pile has cards: %d\n", len(g.pile.cards))
	str := ""
	for _, card := range g.pile.cards {
		str += card.Value + card.Suit + " "
	}
	log.Printf("%s", str)
}

func (g *Game) dump() {
	log.Printf("%#v", g)
}

func (g *Game) prepare() {
	g.sendPlayersEvent()
	g.pile = newPile()
	g.pile.shuffle()
	g.deal()
	g.sendDealEvent()
	g.attackerIndex, g.defenderIndex, g.firstAttackerReasonCard = g.findFirstAttacker()
	g.dumpHands()
	g.dumpPile()
	g.dump()
	g.sendFirstAttackerEvent()
}

func (g *Game) findFirstAttacker() (firstAttackerIndex int, defenderIndex int, lowestTrumpCard *Card) {
	firstAttackerIndex = -1
	defenderIndex = -1
	lowestTrumpCard = &Card{"A", g.trumpSuit}

	for playerIndex, p := range g.players {
		for _, c := range p.cards {
			if c.Suit == g.trumpSuit && c.lte(lowestTrumpCard) {
				firstAttackerIndex = playerIndex
				defenderIndex = g.getNextPlayerIndex(firstAttackerIndex)
				lowestTrumpCard = c
			}
		}
	}

	if firstAttackerIndex >= 0 {
		return
	}

	// fallback
	for playerIndex, p := range g.players {
		for _, c := range p.cards {
			if c.lte(lowestTrumpCard) {
				firstAttackerIndex = playerIndex
				defenderIndex = g.getNextPlayerIndex(firstAttackerIndex)
				lowestTrumpCard = c
			}
		}
	}

	return
}

func (g *Game) getNextPlayerIndex(playerIndex int) (nextPlayerIndex int) {
	nextPlayerIndex = -1
	if len(g.players) > playerIndex+1 {
		nextPlayerIndex = playerIndex + 1
		return
	}
	if len(g.players) > 1 {
		nextPlayerIndex = 0
		return
	}
	return
}

func (g *Game) sendFirstAttackerEvent() {
	fae := GameFirstAttackerEvent{
		ReasonCard:    g.firstAttackerReasonCard,
		AttackerIndex: g.attackerIndex,
		DefenderIndex: g.defenderIndex,
	}
	for _, p := range g.players {
		p.sendEvent(fae)
	}
}

func (g *Game) begin() {
	g.prepare()
	g.status = GameStatusPlaying
	for {
		select {
		case action, ok := <-g.playerActions:
			if !ok {
				return
			}
			log.Printf("action: %#v", action)
			g.onClientAction(action)
		}
	}
}

func (g *Game) canPlayerAttackWithCard(player *Player, card *Card) bool {
	if g.status != GameStatusPlaying {
		return false
	}
	if g.defenderIndex == g.getPlayerIndex(player) {
		return false
	}
	if len(g.battleground) == 0 && g.attackerIndex != g.getPlayerIndex(player) {
		return false
	}
	if len(g.battleground) >= 6 {
		return false
	}

	if len(g.battleground) > 0 && !g.hasBattlegroundSameValue(card) {
		return false
	}

	return true
}

func (g *Game) attack(player *Player, data AttackActionData) {
	card := data.Card
	log.Printf("attack card: %#v", card)
	canAttack := g.canPlayerAttackWithCard(player, card)
	if !canAttack {
		log.Printf("Cannot use card")
		return
	}
	g.battleground = append(g.battleground, card)
	player.removeCard(card)

	gameAttackEvent := GameAttackEvent{
		GameStateInfo: g.getGameStateInfo(),
		AttackerIndex: g.getPlayerIndex(player),
		DefenderIndex: g.defenderIndex,
		Card:          card,
	}

	for _, p := range g.players {
		gameAttackEvent.GameStateInfo.YourHand = p.cards
		p.sendEvent(gameAttackEvent)
	}
}

func (g *Game) onClientAction(action *PlayerAction) {
	if action.Name == PlayerActionNameAttack {
		data, ok := action.Data.(AttackActionData)
		if ok {
			g.attack(action.player, data)
		} else {
			log.Printf("Cannot cast to UseCardActionData: %#v", action.Data)
		}
	}
}

func (g *Game) broadcastEvent(event interface{}) {
	json, _ := eventToJSON(event)
	for _, p := range g.players {
		p.sendMessage(json)
	}
}

func (g *Game) onGameEnded(winnerIndex int) {
	gameEndEvent := &GameEndEvent{winnerIndex}
	g.room.broadcastEvent(gameEndEvent, nil)
	close(g.playerActions)
	g.room.onGameEnded()
}

func (g *Game) onPlayerLeft(playerIndex int) {
	gamePlayerLeft := &GamePlayerLeftEvent{playerIndex}
	g.room.broadcastEvent(gamePlayerLeft, nil)

	winnerIndex := -1
	if len(g.players) == 2 {
		if playerIndex == 0 {
			winnerIndex = 1
		} else {
			winnerIndex = 0
		}
		g.onGameEnded(winnerIndex)
	}
}

func (g *Game) onClientRemoved(client *Client) {
	for index, p := range g.players {
		if p.client.Id() == client.Id() {
			g.onPlayerLeft(index)
			return
		}
	}
}

func (g *Game) findPlayerOfClient(client *Client) *Player {
	for _, p := range g.players {
		if p.client.Id() == client.Id() {
			return p
		}
	}
	return nil
}

func (g *Game) getPlayerIndex(player *Player) int {
	for index, p := range g.players {
		if p.client.Id() == player.client.Id() {
			return index
		}
	}
	return -1
}

func (g *Game) hasBattlegroundSameValue(card *Card) bool {
	for _, c := range g.battleground {
		if c.Value == card.Value {
			return true
		}
	}
	return false
}
