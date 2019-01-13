package main

import (
	"encoding/json"
	"log"
)

type Bot struct {
	botClient       *BotClient
	gameStateInfo   *GameStateInfo
	players         []*Player
	yourPlayerIndex int
	gameWasStarted  bool
	gameIsOver      bool
	myUnbeatenCards map[Card]bool
	iAmPickingUp    bool
}

func newBot(botClient *BotClient) *Bot {
	return &Bot{
		botClient:       botClient,
		yourPlayerIndex: -1,
		players:         make([]*Player, 0),
		myUnbeatenCards: make(map[Card]bool, 0),
	}
}

func (b *Bot) run() {
	for {
		select {
		case event := <-b.botClient.incomingEvents:
			b.dispatchEvent(event)
		}
	}
}

func (b *Bot) dispatchEvent(message []byte) {
	var event JSONEvent
	err := json.Unmarshal(message, &event)
	if err != nil {
		log.Printf("BOT: cannot decode general event: %s", err)
		return
	}

	log.Printf("BOT: event %+v", event)

	// Use this encoding to use following parsing data into struct
	eventDataJson, err := json.Marshal(event.Data)
	if err != nil {
		log.Printf("BOT: error at encoding event data: %s", err)
		return
	}

	switch event.Name {
	case "GamePlayersEvent":
		var parsedEvent GamePlayersEvent
		err = json.Unmarshal(eventDataJson, &parsedEvent)
		if err == nil {
			b.onGamePlayersEvent(parsedEvent)
		}
	case "GameFirstAttackerEvent":
		var parsedEvent GameFirstAttackerEvent
		err = json.Unmarshal(eventDataJson, &parsedEvent)
		if err == nil {
			b.onGameFirstAttackerEvent(parsedEvent)
		}
	case "GameStartedEvent":
		var parsedEvent GameStartedEvent
		err = json.Unmarshal(eventDataJson, &parsedEvent)
		if err == nil {
			b.onGameStartedEvent(parsedEvent)
		}
	case "GameAttackEvent":
		var parsedEvent GameAttackEvent
		err = json.Unmarshal(eventDataJson, &parsedEvent)
		if err == nil {
			b.onGameAttackEvent(parsedEvent)
		}
	case "GameDefendEvent":
		var parsedEvent GameDefendEvent
		err = json.Unmarshal(eventDataJson, &parsedEvent)
		if err == nil {
			b.onGameDefendEvent(parsedEvent)
		}
	case "GameStateEvent":
		var parsedEvent GameStateEvent
		err = json.Unmarshal(eventDataJson, &parsedEvent)
		if err == nil {
			b.onGameStateEvent(parsedEvent)
		}
	case "NewRoundEvent":
		var parsedEvent NewRoundEvent
		err = json.Unmarshal(eventDataJson, &parsedEvent)
		if err == nil {
			b.onNewRoundEvent(parsedEvent)
		}
	case "GameEndEvent":
		var parsedEvent GameEndEvent
		err = json.Unmarshal(eventDataJson, &parsedEvent)
		if err == nil {
			b.onGameEndEvent(parsedEvent)
		}
	}

	if err != nil {
		log.Printf("BOT: error at parsing event data: %s", err)
		return
	}

	b.makeDecision()
}

func (b *Bot) onGamePlayersEvent(event GamePlayersEvent) {
	b.players = event.Players
	b.yourPlayerIndex = event.YourPlayerIndex
}

func (b *Bot) onGameFirstAttackerEvent(event GameFirstAttackerEvent) {
	b.gameStateInfo = event.GameStateInfo
}

func (b *Bot) onGameStartedEvent(event GameStartedEvent) {
	b.gameStateInfo = event.GameStateInfo
	b.gameWasStarted = true
}

func (b *Bot) onGameAttackEvent(event GameAttackEvent) {
	b.gameStateInfo = event.GameStateInfo
}

func (b *Bot) onGameDefendEvent(event GameDefendEvent) {
	b.gameStateInfo = event.GameStateInfo
	delete(b.myUnbeatenCards, *event.AttackingCard)
}

func (b *Bot) onGameStateEvent(event GameStateEvent) {
	b.gameStateInfo = event.GameStateInfo
}

func (b *Bot) onNewRoundEvent(event NewRoundEvent) {
	b.myUnbeatenCards = make(map[Card]bool, 0)
	b.iAmPickingUp = false
	b.gameStateInfo = event.GameStateInfo
}

func (b *Bot) onGameEndEvent(event GameEndEvent) {
	b.gameIsOver = true
}

func (b *Bot) isGameStateValid() bool {
	if len(b.players) < 2 {
		log.Println("BOT: not enough players")
		return false
	}
	if b.yourPlayerIndex < 0 {
		log.Println("BOT: wrong your player index")
		return false
	}
	if b.gameIsOver {
		log.Println("BOT: game is over")
		return false
	}
	if b.gameStateInfo == nil {
		log.Println("BOT: game state info is nil")
		return false
	}
	if !b.gameWasStarted {
		log.Println("BOT: game was not started")
		return false
	}

	return true
}

func (b *Bot) canAttack() bool {
	if b.gameStateInfo.DefenderIndex == b.yourPlayerIndex {
		log.Println("BOT: is defender")
		return false
	}
	if b.gameStateInfo.CompletedPlayers[b.yourPlayerIndex] {
		log.Println("BOT: is completed")
		return false
	}
	if len(b.gameStateInfo.YourHand) == 0 {
		log.Println("BOT: no cards")
		return false
	}
	if len(b.gameStateInfo.Battleground) == 0 && b.gameStateInfo.AttackerIndex != b.yourPlayerIndex {
		log.Println("BOT: not first attacker")
		return false
	}
	if len(b.getAvailableCardsForAttack()) == 0 {
		log.Println("BOT: no cards for attack")
		return false
	}
	if len(b.myUnbeatenCards) > 0 {
		log.Printf("BOT: waiting for cards beaten: %+v", b.myUnbeatenCards)
		return false
	}

	return true
}

func (b *Bot) getAvailableCardsForAttack() (cards []*Card) {
	if len(b.gameStateInfo.Battleground) == 0 {
		return b.gameStateInfo.YourHand
	}

	for _, cardOnHand := range b.gameStateInfo.YourHand {
		if b.hasBattlegroundSameValue(cardOnHand) || b.hasDefendingCardsSameValue(cardOnHand) {
			cards = append(cards, cardOnHand)
		}
	}

	return cards
}

func (b *Bot) attack() {
	b.printHand()
	availableCards := b.getAvailableCardsForAttack()
	minimalValueCard := b.findLowestCard(availableCards)
	attackActionData := AttackActionData{Card: minimalValueCard}
	b.botClient.sendGameAction(PlayerActionNameAttack, attackActionData)
	log.Printf("BOT: attack with %+v", minimalValueCard)
	b.myUnbeatenCards[*minimalValueCard] = true
}

func (b *Bot) canDefend() bool {
	if b.gameStateInfo.DefenderIndex != b.yourPlayerIndex {
		return false
	}
	if len(b.getAttackingCardsToDefend()) == 0 {
		return false
	}

	return true
}

func (b *Bot) getAttackingCardsToDefend() (attackingCards []*Card) {
	for bgIndex, bgCard := range b.gameStateInfo.Battleground {
		if _, ok := b.gameStateInfo.DefendingCards[bgIndex]; !ok {
			attackingCards = append(attackingCards, bgCard)
		}
	}

	return
}

func (b *Bot) defend() {
	b.printHand()
	attackingCards := b.getAttackingCardsToDefend()
	trumpSuit := b.gameStateInfo.TrumpCard.Suit
	for _, attackCard := range attackingCards {
		defendCandidates := make([]*Card, 0)
		for _, hCard := range b.gameStateInfo.YourHand {
			if hCard.Suit == trumpSuit && attackCard.Suit != trumpSuit || hCard.gt(attackCard) {
				defendCandidates = append(defendCandidates, hCard)
			}
		}
		if len(defendCandidates) == 0 {
			log.Printf("BOT: can't beat card %+v", attackCard)
			b.pickUp()
			return
		}

		minimalValueCard := b.findLowestCard(defendCandidates)
		attackActionData := DefendActionData{AttackingCard: attackCard, DefendingCard: minimalValueCard}
		b.botClient.sendGameAction(PlayerActionNameDefend, attackActionData)
	}
}

func (b *Bot) complete() {
	log.Println("BOT: complete")
	b.botClient.sendGameAction(PlayerActionNameComplete, nil)
}

func (b *Bot) pickUp() {
	log.Println("BOT: pick up")
	if b.iAmPickingUp {
		log.Println("BOT: already pick up")
		return
	}
	b.iAmPickingUp = true
	b.botClient.sendGameAction(PlayerActionNamePickUp, nil)
}

func (b *Bot) makeDecision() {
	if !b.isGameStateValid() {
		log.Println("BOT: Invalid state for decision")
		return
	}

	if b.gameStateInfo.DefenderPickUp {
		b.myUnbeatenCards = make(map[Card]bool, 0)
	}

	if b.canAttack() {
		log.Println("BOT: Can attack")
		b.attack()
		return
	} else {
		log.Println("BOT: Can't attack")
	}

	if b.canDefend() {
		b.defend()
		return
	}

	if b.gameStateInfo.CanYouComplete && len(b.myUnbeatenCards) == 0 {
		b.complete()
	}
}

func (b *Bot) hasBattlegroundSameValue(card *Card) bool {
	for _, c := range b.gameStateInfo.Battleground {
		if c.Value == card.Value {
			return true
		}
	}
	return false
}

func (b *Bot) hasDefendingCardsSameValue(card *Card) bool {
	for _, c := range b.gameStateInfo.DefendingCards {
		if c.Value == card.Value {
			return true
		}
	}
	return false
}

func (b *Bot) findLowestCard(cards []*Card) *Card {
	if len(cards) == 0 {
		return nil
	}
	minimalValueCard := cards[0]
	trumpSuit := b.gameStateInfo.TrumpCard.Suit
	for _, avCard := range cards {
		if avCard.Suit != trumpSuit && minimalValueCard.Suit == trumpSuit {
			minimalValueCard = avCard
			continue
		}
		if avCard.getValueIndex() < minimalValueCard.getValueIndex() {
			minimalValueCard = avCard
		}
	}

	return minimalValueCard
}

func (b *Bot) printHand() {
	cardsStr := ""
	for _, card := range b.gameStateInfo.YourHand {
		cardsStr += " " + card.Value + card.Suit
	}
	log.Printf("Cards: %s", cardsStr)
}
