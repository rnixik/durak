package main

import (
	"fmt"
	"log"
	"time"
)

// Statuses of the game.
const (
	GameStatusPreparing = "preparing"
	GameStatusPlaying   = "playing"
	GameStatusEnd       = "end"

	AfkTimeoutSeconds = 120
)

// Game represents status, state, etc of the game.
type Game struct {
	id                            string
	playerActions                 chan *PlayerAction
	owner                         *Player
	room                          *Room
	status                        string
	players                       []*Player
	deck                          *Deck
	discardPileSize               int
	trumpSuit                     string
	trumpCard                     *Card
	trumpCardIsOwnedByPlayerIndex int
	firstAttackerReasonCard       *Card
	attackerIndex                 int
	defenderIndex                 int
	battleground                  []*Card
	defendingCards                map[int]*Card
	defenderPickUp                bool
	afkTimers                     map[int]*time.Timer
	gameLogger                    GameLogger
}

// GameLogger stores game events
type GameLogger interface {
	// Save event when game starts
	LogGameBegins(game *Game)
	// Save event when a player attacks with card
	LogPlayerActionAttack(game *Game, data AttackActionData)
	// Save event when a player defends card
	LogPlayerActionDefend(game *Game, data DefendActionData)
	// Save event when a player picks up cards from desk
	LogPlayerActionPickUp(game *Game)
	// Save event when a players completes a round
	LogPlayerActionComplete(game *Game)
	// Save event when game ends
	LogGameEnds(game *Game, hasLoser bool, loserIndex int) error
}

func newGame(room *Room, players []*Player, gameLogger GameLogger) *Game {
	currentTime := time.Now()
	gameId := fmt.Sprintf(
		"%d%02d%02d_%02d%02d%02d_%d",
		currentTime.Year(),
		currentTime.Month(),
		currentTime.Day(),
		currentTime.Hour(),
		currentTime.Minute(),
		currentTime.Second(),
		room.id,
	)
	return &Game{
		id:             gameId,
		room:           room,
		playerActions:  make(chan *PlayerAction),
		status:         GameStatusPreparing,
		players:        players,
		battleground:   make([]*Card, 0),
		defendingCards: make(map[int]*Card, 0),
		afkTimers:      make(map[int]*time.Timer, 0),
		gameLogger:     gameLogger,
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
}

func (g *Game) deal() {
	cardsLimit := 6
	var lastCard *Card
	lastPlayerIndex := -1
	for cardIndex := 0; cardIndex < cardsLimit; cardIndex = cardIndex + 1 {
		for playerIndex, p := range g.players {
			if !p.IsActive || len(p.cards) >= cardsLimit {
				break
			}
			card, err := g.deck.getCard()
			if err == nil {
				p.cards = append(p.cards, card)
				lastCard = card
				lastPlayerIndex = playerIndex
			}
		}
	}
	if len(g.deck.cards) > 0 {
		lastCard = g.deck.cards[0]
		g.trumpCardIsOwnedByPlayerIndex = -1
	} else {
		g.trumpCardIsOwnedByPlayerIndex = lastPlayerIndex
	}

	g.trumpCard = lastCard
	if lastCard != nil {
		g.trumpSuit = lastCard.Suit
	}
}

func (g *Game) dealToPlayer(player *Player) {
	if !player.IsActive {
		return
	}
	cardsLimit := 6
	for cardIndex := len(player.cards); cardIndex < cardsLimit; cardIndex = cardIndex + 1 {
		if len(player.cards) >= cardsLimit {
			break
		}
		card, err := g.deck.getCard()
		if err == nil {
			player.cards = append(player.cards, card)
		}
	}
	if len(player.cards) == 0 {
		player.IsActive = false
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
		DeckSize:         len(g.deck.cards),
		DiscardPileSize:  g.discardPileSize,
		TrumpCard:        g.trumpCard,
		Battleground:     g.battleground,
		DefendingCards:   g.defendingCards,
		CompletedPlayers: make(map[int]bool, 0),
		DefenderPickUp:   g.defenderPickUp,
		AttackerIndex:    g.attackerIndex,
		DefenderIndex:    g.defenderIndex,
	}

	for i, p := range g.players {
		if p == player {
			gsi.YourHand = p.cards
			gsi.CanYouComplete = g.canPlayerComplete(player)
			gsi.CanYouAttack = g.canPlayerAttack(player)
			gsi.CanYouPickUp = g.canPlayerPickUp(player)
		}
		gsi.HandsSizes[i] = len(p.cards)
		gsi.CompletedPlayers[i] = p.IsCompleted
	}

	return gsi
}

func (g *Game) sendDealEvent() {
	de := GameDealEvent{
		TrumpCardIsOwnedByPlayerIndex: g.trumpCardIsOwnedByPlayerIndex,
	}

	for _, p := range g.players {
		de.GameStateInfo = g.getGameStateInfo(p)
		p.sendEvent(de)
	}
}

func (g *Game) prepare() {
	g.sendPlayersEvent()
	g.deck = newDeck()
	g.deck.shuffle()
	g.deal()
	g.sendDealEvent()
	g.attackerIndex, g.defenderIndex, g.firstAttackerReasonCard = g.findFirstAttacker()
	g.sendFirstAttackerEvent()
}

func (g *Game) findFirstAttacker() (firstAttackerIndex int, defenderIndex int, lowestTrumpCard *Card) {
	firstAttackerIndex = -1
	defenderIndex = -1
	lowestTrumpCard = &Card{"A", g.trumpSuit}

	for playerIndex, p := range g.players {
		for _, c := range p.cards {
			if c.Suit == g.trumpSuit && c.lte(lowestTrumpCard) {
				firstAttackerIndex = g.adjustPlayerIndex(playerIndex)
				defenderIndex = g.adjustPlayerIndex(firstAttackerIndex + 1)
				lowestTrumpCard = c
			}
		}
	}

	if firstAttackerIndex >= 0 {
		return
	}

	// fallback
	for playerIndex, p := range g.players {
		if !p.IsActive {
			break
		}
		for _, c := range p.cards {
			if c.lte(lowestTrumpCard) {
				firstAttackerIndex = playerIndex
				defenderIndex = g.adjustPlayerIndex(firstAttackerIndex + 1)
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
		if g.getActivePlayersNum() == 2 {
			attackerIndex = g.attackerIndex
			defenderIndex = g.defenderIndex
		} else {
			attackerIndex = attackerIndex + 1
			defenderIndex = defenderIndex + 1
		}

	}

	attackerIndex, defenderIndex = g.adjustPlayerIndex(attackerIndex), g.adjustPlayerIndex(defenderIndex)
	if attackerIndex == defenderIndex {
		attackerIndex = g.adjustPlayerIndex(attackerIndex + 1)
	}

	return
}

func (g *Game) sendFirstAttackerEvent() {
	fae := GameFirstAttackerEvent{
		ReasonCard: g.firstAttackerReasonCard,
	}
	for _, p := range g.players {
		fae.GameStateInfo = g.getGameStateInfo(p)
		p.sendEvent(fae)
	}
}

func (g *Game) begin() {
	g.prepare()
	g.status = GameStatusPlaying

	gse := &GameStartedEvent{}
	for _, p := range g.players {
		gse.GameStateInfo = g.getGameStateInfo(p)
		p.sendEvent(gse)
	}

	g.gameLogger.LogGameBegins(g)
	g.room.onGameStarted()
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

func (g *Game) canPlayerAttack(player *Player) bool {
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
	if len(g.battleground) >= len(g.players[g.defenderIndex].cards)+len(g.defendingCards) {
		return false
	}
	if player.IsCompleted {
		return false
	}

	return true
}

func (g *Game) canPlayerAttackWithCard(player *Player, card *Card) bool {
	if !g.canPlayerAttack(player) {
		return false
	}
	if !player.hasCard(card) {
		return false
	}
	if len(g.battleground) > 0 && (!g.hasBattlegroundSameValue(card) && !g.hasDefendingCardsSameValue(card)) {
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
	g.gameLogger.LogPlayerActionAttack(g, data)

	gameAttackEvent := GameAttackEvent{
		AttackerIndex: g.getPlayerIndex(player),
		DefenderIndex: g.defenderIndex,
		Card:          card,
	}

	for _, p := range g.players {
		gameAttackEvent.GameStateInfo = g.getGameStateInfo(p)
		p.sendEvent(gameAttackEvent)
	}

	g.restartAfkTimers(g.defenderIndex)
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

	g.gameLogger.LogPlayerActionDefend(g, data)

	// Allow other players to attack in the middle of a round
	g.resetPlayersCompleteStatuses()

	gameAttackEvent := GameDefendEvent{
		DefenderIndex: g.getPlayerIndex(player),
		AttackingCard: data.AttackingCard,
		DefendingCard: data.DefendingCard,
	}

	for _, p := range g.players {
		gameAttackEvent.GameStateInfo = g.getGameStateInfo(p)
		p.sendEvent(gameAttackEvent)
	}

	if len(g.defendingCards) == len(g.battleground) {
		// All cards are beaten, wait for attacker
		g.restartAfkTimers(g.attackerIndex)
	} else {
		// Not all cards are beaten, wait for defender
		g.restartAfkTimers(g.defenderIndex)
	}
}

func (g *Game) canPlayerPickUp(player *Player) bool {
	if g.status != GameStatusPlaying {
		return false
	}
	if !player.IsActive {
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
	g.gameLogger.LogPlayerActionPickUp(g)

	if g.areAllPlayersCompleted() {
		g.endRound()
	} else {
		g.broadcastGameStateEvent()
	}

	g.restartAfkTimers(g.defenderIndex)
}

func (g *Game) canPlayerComplete(player *Player) bool {
	if g.status != GameStatusPlaying {
		return false
	}
	if !player.IsActive {
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

	if len(g.battleground) != len(g.defendingCards) && !g.defenderPickUp {
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
	g.gameLogger.LogPlayerActionComplete(g)

	if g.areAllPlayersCompleted() {
		g.endRound()
	} else {
		g.broadcastGameStateEvent()
	}
}

func (g *Game) endRound() {
	g.resetPlayersCompleteStatuses()
	// Can change number of active players to end game
	g.roundDeal(g.attackerIndex, g.defenderIndex)

	defenderPlayer := g.players[g.defenderIndex]
	wasAttackSuccessful := g.defenderPickUp

	if g.defenderPickUp {
		for _, c := range g.battleground {
			defenderPlayer.cards = append(defenderPlayer.cards, c)
		}
		for _, c := range g.defendingCards {
			defenderPlayer.cards = append(defenderPlayer.cards, c)
		}
	} else {
		g.discardPileSize = g.discardPileSize + len(g.battleground) + len(g.defendingCards)
	}

	g.attackerIndex, g.defenderIndex = g.findNewAttacker(g.defenderPickUp)
	g.battleground = make([]*Card, 0)
	g.defendingCards = make(map[int]*Card, 0)
	g.defenderPickUp = false

	activePlayers := g.getActivePlayers()

	if len(activePlayers) < 2 {
		// End of game
		hasLoser := false
		loserIndex := -1
		if len(activePlayers) == 1 {
			hasLoser = true
			loserIndex = g.getPlayerIndex(activePlayers[0])
		}
		g.endGame(hasLoser, loserIndex)
		g.restartAfkTimers(-1)
	} else {
		nrd := NewRoundEvent{WasAttackSuccessful: wasAttackSuccessful}
		for _, p := range g.players {
			nrd.GameStateInfo = g.getGameStateInfo(p)
			p.sendEvent(nrd)
		}
		g.restartAfkTimers(g.attackerIndex)
	}
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

func (g *Game) endGame(hasLoser bool, loserIndex int) {
	if g.status != GameStatusPlaying {
		return
	}
	g.status = GameStatusEnd
	g.broadcastGameStateEvent()
	gameEndEvent := &GameEndEvent{
		HasLoser:   hasLoser,
		LoserIndex: loserIndex,
	}
	err := g.gameLogger.LogGameEnds(g, hasLoser, loserIndex)
	if err != nil {
		log.Printf("Error writing log: %s", err)
	}
	g.room.broadcastEvent(gameEndEvent, nil)
	close(g.playerActions)
	g.room.onGameEnded()
}

func (g *Game) onActivePlayerLeft(playerIndex int, isAfk bool) {
	gamePlayerLeft := &GamePlayerLeftEvent{playerIndex, isAfk}
	g.room.broadcastEvent(gamePlayerLeft, nil)

	if g.getActivePlayersNum() == 2 {
		g.endGame(true, playerIndex)
	}
}

func (g *Game) onLatePlayerJoin(player *Player) {
	g.sendPlayersEvent()

	gameStateEvent := GameStateEvent{}
	gameStateEvent.GameStateInfo = g.getGameStateInfo(player)
	player.sendEvent(gameStateEvent)
}

func (g *Game) onClientRemoved(client ClientSender) {
	for index, p := range g.players {
		if p.client.Id() == client.Id() && p.IsActive {
			g.onActivePlayerLeft(index, false)
			return
		}
	}
}

func (g *Game) restartAfkTimers(waitingPlayerIndex int) {
	for plIndex, timer := range g.afkTimers {
		timer.Stop()
		delete(g.afkTimers, plIndex)
	}

	if waitingPlayerIndex < 0 {
		return
	}

	timer, ok := g.afkTimers[waitingPlayerIndex]
	if ok {
		log.Println("stopping timer", waitingPlayerIndex)
		timer.Stop()
	}

	timer = time.AfterFunc(time.Second*AfkTimeoutSeconds, func() {
		log.Println("time after func", waitingPlayerIndex)
		delete(g.afkTimers, waitingPlayerIndex)
		g.onActivePlayerLeft(waitingPlayerIndex, true)
	})

	g.afkTimers[waitingPlayerIndex] = timer
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
		if p.IsActive && !p.IsCompleted && !isDefenderAndPickUp {
			return false
		}
	}
	return true
}

func (g *Game) getActivePlayers() (players []*Player) {
	for _, p := range g.players {
		if p.IsActive {
			players = append(players, p)
		}
	}
	return
}

func (g *Game) getActivePlayersNum() int {
	return len(g.getActivePlayers())
}

func (g *Game) adjustPlayerIndex(index int) int {
	activePlayersNum := g.getActivePlayersNum()
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
