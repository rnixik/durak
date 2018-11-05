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
	pile                          *Pile // deck actually :/
	discardPileSize               int
	trumpSuit                     string
	trumpCard                     *Card
	trumpCardIsInPile             bool
	trumpCardIsOwnedByPlayerIndex int
	firstAttackerReasonCard       *Card
	attackerIndex                 int
	defenderIndex                 int
	battleground                  []*Card
	defendingCards                map[int]*Card
	defenderPickUp                bool
}

func newGame(id uint64, room *Room, players []*Player) *Game {
	return &Game{
		id:             id,
		room:           room,
		playerActions:  make(chan *PlayerAction),
		status:         GameStatusPreparing,
		players:        players,
		battleground:   make([]*Card, 0),
		defendingCards: make(map[int]*Card, 0),
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

func (g *Game) dealToPlayer(player *Player) {
	cardsLimit := 6
	for cardIndex := len(player.cards); cardIndex < cardsLimit; cardIndex = cardIndex + 1 {
		if len(player.cards) >= cardsLimit {
			break
		}
		card, err := g.pile.getCard()
		if err == nil {
			player.cards = append(player.cards, card)
		}
	}
}

func (g *Game) roundDeal(firstIndex int, lastIndex int) {
	g.dealToPlayer(g.players[firstIndex])
	for index, p := range g.players {
		if index != firstIndex && index != lastIndex {
			g.dealToPlayer(p)
		}
	}
	g.dealToPlayer(g.players[lastIndex])
}

func (g *Game) getGameStateInfo(player *Player) *GameStateInfo {
	gsi := &GameStateInfo{
		YourHand:         make([]*Card, 0),
		HandsSizes:       make([]int, len(g.players)),
		PileSize:         len(g.pile.cards),
		DiscardPileSize:  g.discardPileSize,
		Battleground:     g.battleground,
		DefendingCards:   g.defendingCards,
		CompletedPlayers: make(map[int]bool, 0),
		AttackerIndex:    g.attackerIndex,
		DefenderIndex:    g.defenderIndex,
	}

	for i, p := range g.players {
		if p == player {
			gsi.YourHand = p.cards
			gsi.CanYouComplete = g.canPlayerComplete(player)
			gsi.CanYouPickUp = g.canPlayerPickUp(player)
		}
		gsi.HandsSizes[i] = len(p.cards)
		gsi.CompletedPlayers[i] = p.IsCompleted
	}

	return gsi
}

func (g *Game) sendDealEvent() {
	de := GameDealEvent{
		TrumpSuit:                     g.trumpSuit,
		TrumpCard:                     g.trumpCard,
		TrumpCardIsInPile:             g.trumpCardIsInPile,
		TrumpCardIsOwnedByPlayerIndex: g.trumpCardIsOwnedByPlayerIndex,
	}

	for _, p := range g.players {
		de.GameStateInfo = g.getGameStateInfo(p)
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

func (g *Game) findNewAttacker(wasPickUp bool) (attackerIndex int, defenderIndex int) {
	attackerIndex = g.attackerIndex + 1
	defenderIndex = g.defenderIndex + 1

	if wasPickUp {
		attackerIndex = attackerIndex + 1
		defenderIndex = defenderIndex + 1
	}

	return g.adjustPlayerIndex(attackerIndex), g.adjustPlayerIndex(defenderIndex)
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
	if !player.hasCard(card) {
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

	if len(g.battleground) > 0 && (!g.hasBattlegroundSameValue(card) && !g.hasDefendingCardsSameValue(card)) {
		return false
	}
	if player.IsCompleted {
		return false
	}

	return true
}

func (g *Game) canPlayerDefendWithCard(player *Player, attackingCard *Card, defendingCard *Card) bool {
	if g.status != GameStatusPlaying {
		return false
	}
	if !player.hasCard(defendingCard) {
		return false
	}
	if g.defenderIndex != g.getPlayerIndex(player) {
		return false
	}
	if !g.hasBattlegroundCard(attackingCard) {
		return false
	}
	if g.trumpSuit != defendingCard.Suit && g.trumpSuit == attackingCard.Suit {
		return false
	}
	if g.trumpSuit == defendingCard.Suit && g.trumpSuit != attackingCard.Suit {
		return true
	}
	return defendingCard.gt(attackingCard)
}

func (g *Game) attack(player *Player, data AttackActionData) {
	card := data.Card
	log.Printf("attack card: %#v", card)
	canAttack := g.canPlayerAttackWithCard(player, card)
	if !canAttack {
		log.Printf("Cannot use card to attack")
		return
	}
	g.battleground = append(g.battleground, card)
	player.removeCard(card)

	gameAttackEvent := GameAttackEvent{
		AttackerIndex: g.getPlayerIndex(player),
		DefenderIndex: g.defenderIndex,
		Card:          card,
	}

	for _, p := range g.players {
		gameAttackEvent.GameStateInfo = g.getGameStateInfo(p)
		p.sendEvent(gameAttackEvent)
	}
}

func (g *Game) defend(player *Player, data DefendActionData) {
	canDefend := g.canPlayerDefendWithCard(player, data.AttackingCard, data.DefendingCard)
	if !canDefend {
		log.Printf("Cannot use card to defend")
		return
	}
	attackingIndex := g.getBattlegroundCardIndex(data.AttackingCard)
	g.defendingCards[attackingIndex] = data.DefendingCard
	player.removeCard(data.DefendingCard)

	gameAttackEvent := GameDefendEvent{
		DefenderIndex: g.getPlayerIndex(player),
		AttackingCard: data.AttackingCard,
		DefendingCard: data.DefendingCard,
	}

	for _, p := range g.players {
		gameAttackEvent.GameStateInfo = g.getGameStateInfo(p)
		p.sendEvent(gameAttackEvent)
	}
}

func (g *Game) canPlayerPickUp(player *Player) bool {
	if g.status != GameStatusPlaying {
		return false
	}
	if g.defenderIndex != g.getPlayerIndex(player) {
		return false
	}

	if len(g.battleground) == 0 {
		return false
	}
	if g.defenderPickUp {
		return false
	}

	return true
}

func (g *Game) pickUp(player *Player) {
	log.Printf("pick up")
	if !g.canPlayerPickUp(player) {
		log.Printf("Can't pick up")
		return
	}
	g.defenderPickUp = true
	if g.areAllPlayersCompleted() {
		g.endRound()
	} else {
		g.broadcastGameStateEvent()
	}
}

func (g *Game) canPlayerComplete(player *Player) bool {
	if g.status != GameStatusPlaying {
		return false
	}
	if g.defenderIndex == g.getPlayerIndex(player) &&
		(len(g.battleground) != len(g.defendingCards) || !g.areAllAttackersCompleted()) {
		return false
	}
	if g.attackerIndex == g.getPlayerIndex(player) && len(g.battleground) == 0 {
		return false
	}
	if player.IsCompleted {
		return false
	}

	return true
}

func (g *Game) complete(player *Player) {
	log.Printf("complete")
	if !g.canPlayerComplete(player) {
		log.Printf("Can't complete")
		return
	}

	player.IsCompleted = true

	if g.areAllPlayersCompleted() {
		g.endRound()
	} else {
		g.broadcastGameStateEvent()
	}
}

func (g *Game) endRound() {
	g.resetPlayersCompleteStatuses()
	g.roundDeal(g.attackerIndex, g.defenderIndex)
	g.attackerIndex, g.defenderIndex = g.findNewAttacker(g.defenderPickUp)

	if g.defenderPickUp {
		defenderPlayer := g.players[g.defenderIndex]
		for _, c := range g.battleground {
			defenderPlayer.cards = append(defenderPlayer.cards, c)
		}
		for _, c := range g.defendingCards {
			defenderPlayer.cards = append(defenderPlayer.cards, c)
		}
	} else {
		g.discardPileSize = g.discardPileSize + len(g.battleground) + len(g.defendingCards)
	}

	g.battleground = make([]*Card, 0)
	g.defendingCards = make(map[int]*Card, 0)
	g.defenderPickUp = false

	g.broadcastGameStateEvent()
}

func (g *Game) resetPlayersCompleteStatuses() {
	for _, p := range g.players {
		p.IsCompleted = false
	}
}

func (g *Game) broadcastGameStateEvent() {
	gameStateEvent := GameStateEvent{}

	for _, p := range g.players {
		gameStateEvent.GameStateInfo = g.getGameStateInfo(p)
		p.sendEvent(gameStateEvent)
	}
}

func (g *Game) onClientAction(action *PlayerAction) {
	if action.Name == PlayerActionNameAttack {
		data, ok := action.Data.(AttackActionData)
		if ok {
			g.attack(action.player, data)
		}
	} else if action.Name == PlayerActionNameDefend {
		data, ok := action.Data.(DefendActionData)
		if ok {
			g.defend(action.player, data)
		}
	} else if action.Name == PlayerActionNamePickUp {
		g.pickUp(action.player)
	} else if action.Name == PlayerActionNameComplete {
		g.complete(action.player)
	} else {
		log.Printf("Unknown game action: %s", action.Name)
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

func (g *Game) hasDefendingCardsSameValue(card *Card) bool {
	for _, c := range g.defendingCards {
		if c.Value == card.Value {
			return true
		}
	}
	return false
}

func (g *Game) hasBattlegroundCard(card *Card) bool {
	for _, c := range g.battleground {
		if c.equals(card) {
			return true
		}
	}
	return false
}

func (g *Game) areAllAttackersCompleted() bool {
	for _, p := range g.players {
		if p.IsActive == true && g.getPlayerIndex(p) != g.defenderIndex && !p.IsCompleted {
			return false
		}
	}
	return true
}

func (g *Game) areAllPlayersCompleted() bool {
	for _, p := range g.players {
		isDefenderAndPickUp := g.getPlayerIndex(p) == g.defenderIndex && g.defenderPickUp
		if p.IsActive == true && !p.IsCompleted && !isDefenderAndPickUp {
			return false
		}
	}
	return true
}

func (g *Game) adjustPlayerIndex(index int) int {
	activePlayersNum := 0
	for _, p := range g.players {
		if p.IsActive {
			activePlayersNum += 1
		}
	}
	if activePlayersNum < 2 {
		return -1
	}

	index = index % len(g.players)
	if !g.players[index].IsActive {
		return g.adjustPlayerIndex(index + 1)
	}

	return index
}

func (g *Game) getBattlegroundCardIndex(card *Card) int {
	for index, c := range g.battleground {
		if c.equals(card) {
			return index
		}
	}
	return -1
}
