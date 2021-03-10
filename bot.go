package main

import (
	"encoding/json"
	"log"
)

const additionalAttackStopIndex = float64(0.25)

// Bot represents a AI player
type Bot struct {
	botClient         *BotClient
	gameStateInfo     *GameStateInfo
	players           []*Player
	yourPlayerIndex   int
	gameWasStarted    bool
	gameIsOver        bool
	myUnbeatenCards   map[Card]bool
	iAmPickingUp      bool
	initialPlayersNum int
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

	eventDataJson, err := json.Marshal(event.Data)
	if err != nil {
		log.Printf("BOT: error at encoding event data: %s", err)
		return
	}

	eventHandlers := map[string]func(b *Bot) (func(), error){
		"GamePlayersEvent": func(b *Bot) (func(), error) {
			var parsedEvent GamePlayersEvent
			err = json.Unmarshal(eventDataJson, &parsedEvent)

			return func() {
				b.onGamePlayersEvent(parsedEvent)
			}, err
		},
		"GameFirstAttackerEvent": func(b *Bot) (func(), error) {
			var parsedEvent GameFirstAttackerEvent
			err = json.Unmarshal(eventDataJson, &parsedEvent)

			return func() {
				b.onGameFirstAttackerEvent(parsedEvent)
			}, err
		},
		"GameStartedEvent": func(b *Bot) (func(), error) {
			var parsedEvent GameStartedEvent
			err = json.Unmarshal(eventDataJson, &parsedEvent)

			return func() {
				b.onGameStartedEvent(parsedEvent)
			}, err
		},
		"GameAttackEvent": func(b *Bot) (func(), error) {
			var parsedEvent GameAttackEvent
			err = json.Unmarshal(eventDataJson, &parsedEvent)

			return func() {
				b.onGameAttackEvent(parsedEvent)
			}, err
		},
		"GameDefendEvent": func(bot *Bot) (func(), error) {
			var parsedEvent GameDefendEvent
			err = json.Unmarshal(eventDataJson, &parsedEvent)

			return func() {
				b.onGameDefendEvent(parsedEvent)
			}, err
		},
		"GameStateEvent": func(bot *Bot) (func(), error) {
			var parsedEvent GameStateEvent
			err = json.Unmarshal(eventDataJson, &parsedEvent)

			return func() {
				b.onGameStateEvent(parsedEvent)
			}, err
		},
		"NewRoundEvent": func(bot *Bot) (func(), error) {
			var parsedEvent NewRoundEvent
			err = json.Unmarshal(eventDataJson, &parsedEvent)

			return func() {
				b.onNewRoundEvent(parsedEvent)
			}, err
		},
		"GameEndEvent": func(bot *Bot) (func(), error) {
			var parsedEvent GameEndEvent
			err = json.Unmarshal(eventDataJson, &parsedEvent)

			return func() {
				b.onGameEndEvent()
			}, err
		},
	}

	handler, ok := eventHandlers[event.Name]
	if !ok {
		return
	}

	finishExec, err := handler(b)

	if err != nil {
		log.Printf("BOT: error at parsing event data: %s", err)
		return
	}

	finishExec()
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
	b.gameIsOver = false
	b.initialPlayersNum = len(b.players)
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

func (b *Bot) onGameEndEvent() {
	b.gameIsOver = true
}

func (b *Bot) isGameStateValid() bool {
	if len(b.players) < 2 {
		// log.Println("BOT: not enough players")
		return false
	}
	if b.yourPlayerIndex < 0 {
		// log.Println("BOT: wrong your player index")
		return false
	}
	if b.gameIsOver {
		// log.Println("BOT: game is over")
		return false
	}
	if b.gameStateInfo == nil {
		// log.Println("BOT: game state info is nil")
		return false
	}
	if !b.gameWasStarted {
		// log.Println("BOT: game was not started")
		return false
	}

	return true
}

func (b *Bot) canAttack() bool {
	if !b.gameStateInfo.CanYouAttack {
		return false
	}
	if len(b.getAvailableCardsForAttack()) == 0 {
		return false
	}
	if len(b.myUnbeatenCards) > 0 {
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

func (b *Bot) attack() bool {
	availableCards := b.getAvailableCardsForAttack()
	minimalValueCard := b.findLowestCard(availableCards)

	// Should bot add card to strong cards on table?
	battlegroundPickUpValue := b.getTablePickUpValue(minimalValueCard)
	if battlegroundPickUpValue > additionalAttackStopIndex {
		return false
	}

	attackActionData := AttackActionData{Card: minimalValueCard}
	b.botClient.sendGameAction(PlayerActionNameAttack, attackActionData)
	b.myUnbeatenCards[*minimalValueCard] = true

	return true
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
			b.pickUp()
			return
		}

		minimalValueCard := b.findLowestCard(defendCandidates)
		defendActionData := DefendActionData{AttackingCard: attackCard, DefendingCard: minimalValueCard}
		b.botClient.sendGameAction(PlayerActionNameDefend, defendActionData)
	}
}

func (b *Bot) complete() {
	b.botClient.sendGameAction(PlayerActionNameComplete, nil)
}

func (b *Bot) pickUp() {
	if b.iAmPickingUp {
		return
	}
	b.iAmPickingUp = true
	b.botClient.sendGameAction(PlayerActionNamePickUp, nil)
}

func (b *Bot) makeDecision() {
	if !b.isGameStateValid() {
		return
	}

	if b.gameStateInfo.DefenderPickUp {
		b.myUnbeatenCards = make(map[Card]bool, 0)
	}

	// log.Println(b.getTablePickUpValue(nil))

	if b.canAttack() {
		if len(b.gameStateInfo.Battleground) == 0 {
			if b.attack() {
				return
			}

		}
		// Should bot add card to strong cards on table?
		battlegroundPickUpValue := b.getTablePickUpValue(nil)
		if battlegroundPickUpValue < additionalAttackStopIndex {
			if b.attack() {
				return
			}
		}
	}

	if b.canDefend() {
		b.defend()
		return
	}

	if b.gameStateInfo.CanYouComplete {
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

		sameSuitType := avCard.Suit == trumpSuit && minimalValueCard.Suit == trumpSuit ||
			avCard.Suit != trumpSuit && minimalValueCard.Suit != trumpSuit

		if sameSuitType && avCard.getValueIndex() < minimalValueCard.getValueIndex() {
			minimalValueCard = avCard
		}
	}

	return minimalValueCard
}

func (b *Bot) getTablePickUpValue(additionalCard *Card) float64 {
	if len(b.gameStateInfo.Battleground) == 0 {
		return float64(0)
	}

	deckRemainsIndex := b.getDeckRemainsIndex()
	battlegroundAttackRateIndex := b.getCardsOnTablePowerRate(additionalCard)
	tripletsIndex, quartetsIndex := b.getPairsOnTableIndex(additionalCard)

	cardsPowerIndex := (10*battlegroundAttackRateIndex + tripletsIndex + 2*quartetsIndex) / float64(13)

	return deckRemainsIndex * cardsPowerIndex
}

// How many cards left in deck: 0..1: 0 - empty deck; 1 - full deck.
func (b *Bot) getDeckRemainsIndex() float64 {
	deckRemainsIndex := float64(0)
	if b.initialPlayersNum < 6 {
		deckRemainsIndex = float64(b.gameStateInfo.DeckSize) / float64(36-b.initialPlayersNum*6)
	}

	return deckRemainsIndex
}

// Calculate power of cards on table in range 0..1, where 1 is maximum possible cards power
func (b *Bot) getCardsOnTablePowerRate(additionalCard *Card) float64 {
	// Each card has attack rate from 0 ("6") to 9 ("A")
	// Each trump card has attack rate from 10 (trump "6") to 18 (trump "A")
	// maxAttackRate - cards with highest value: trump "A", "K", "Q", ...
	// We need this value to get attack rate of battleground in range of 0..1
	maxAttackRate := 0
	maxAttackRatePerCurrentCard := 18 // Trump "A" has maximum attack rate

	totalCardsOnTable := len(b.gameStateInfo.Battleground) + len(b.gameStateInfo.DefendingCards)
	if additionalCard != nil {
		totalCardsOnTable += 1
	}
	for i := 0; i < totalCardsOnTable; i++ {
		maxAttackRate += maxAttackRatePerCurrentCard
		maxAttackRatePerCurrentCard -= 1
	}

	getCardAttackRate := func(card *Card) int {
		if card.Suit == b.gameStateInfo.TrumpCard.Suit {
			return card.getValueIndex() + 9
		}
		return card.getValueIndex()
	}

	battlegroundAttackRate := 0
	for _, card := range b.gameStateInfo.Battleground {
		battlegroundAttackRate += getCardAttackRate(card)
	}
	for _, card := range b.gameStateInfo.DefendingCards {
		battlegroundAttackRate += getCardAttackRate(card)
	}

	if additionalCard != nil {
		battlegroundAttackRate += getCardAttackRate(additionalCard)
	}

	battlegroundAttackRateIndex := float64(0)
	if maxAttackRate > 0 {
		battlegroundAttackRateIndex = float64(battlegroundAttackRate) / float64(maxAttackRate)
	}

	return battlegroundAttackRateIndex
}

// Returns indexes based on how many pairs on table
func (b *Bot) getPairsOnTableIndex(additionalCard *Card) (tripletsIndex float64, quartetsIndex float64) {
	pairNumberCards := b.getCardsNumberOnTable()
	if additionalCard != nil {
		if _, ok := pairNumberCards[additionalCard.Value]; ok {
			pairNumberCards[additionalCard.Value] += 1
		} else {
			pairNumberCards[additionalCard.Value] = 1
		}
	}

	tripletsIndex = getCardsPairsIndex(pairNumberCards, 3)
	quartetsIndex = getCardsPairsIndex(pairNumberCards, 4)

	return
}

func (b *Bot) getCardsNumberOnTable() map[string]int {
	cards := make(map[string]int, 0)
	for _, card := range b.gameStateInfo.Battleground {
		if _, ok := cards[card.Value]; ok {
			cards[card.Value] += 1
		} else {
			cards[card.Value] = 1
		}
	}
	for _, card := range b.gameStateInfo.DefendingCards {
		if _, ok := cards[card.Value]; ok {
			cards[card.Value] += 1
		} else {
			cards[card.Value] = 1
		}
	}

	return cards
}

func getCardsPairsIndex(cards map[string]int, pairSize int) float64 {
	totalCards := 0
	pairsNumber := 0
	for _, number := range cards {
		totalCards += number
		if number == pairSize {
			pairsNumber += 1
		}
	}

	maxPairsNumber := totalCards / pairSize

	index := float64(0)
	if maxPairsNumber > 0 {
		index = float64(pairsNumber) / float64(maxPairsNumber)
	}

	return index
}
