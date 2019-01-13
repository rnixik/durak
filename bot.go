package main

import (
	"encoding/json"
	"log"
)

type Bot struct {
	botClient              *BotClient
	gameStateInfo          *GameStateInfo
	players                []*Player
	yourPlayerIndex        int
	gameIsOver             bool
	waitingForMyCardBeaten bool
}

func newBot(botClient *BotClient) *Bot {
	return &Bot{
		botClient:       botClient,
		yourPlayerIndex: -1,
		players:         make([]*Player, 0),
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
		log.Printf("cannot decode general bot event: %s", err)
		return
	}

	log.Printf("event for bot %+v", event)

	// Use this encoding to use following parsing data into struct
	eventDataJson, err := json.Marshal(event.Data)
	if err != nil {
		log.Printf("error at encoding event data: %s", err)
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
		log.Printf("error at parsing event data: %s", err)
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
	b.makeDecision()
}

func (b *Bot) onGameAttackEvent(event GameAttackEvent) {
	b.gameStateInfo = event.GameStateInfo
	b.makeDecision()
}

func (b *Bot) onGameDefendEvent(event GameDefendEvent) {
	b.gameStateInfo = event.GameStateInfo
	b.makeDecision()
}

func (b *Bot) onGameStateEvent(event GameStateEvent) {
	b.gameStateInfo = event.GameStateInfo
	b.makeDecision()
}

func (b *Bot) onNewRoundEvent(event NewRoundEvent) {
	b.gameStateInfo = event.GameStateInfo
	b.makeDecision()
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

	return true
}

func (b *Bot) canAttack() bool {
	if b.gameStateInfo.DefenderIndex == b.yourPlayerIndex {
		return false
	}
	if b.gameStateInfo.CompletedPlayers[b.yourPlayerIndex] {
		return false
	}
	if len(b.gameStateInfo.YourHand) == 0 {
		return false
	}
	if len(b.gameStateInfo.Battleground) == 0 && b.gameStateInfo.AttackerIndex != b.yourPlayerIndex {
		return false
	}
	if len(b.getAvailableCardsForAttack()) == 0 {
		log.Println("BOT: no cards for attack")
		return false
	}
	if b.waitingForMyCardBeaten {
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
	availableCards := b.getAvailableCardsForAttack()
	minimalValueCard := availableCards[0]
	trumpSuit := b.gameStateInfo.TrumpCard.Suit
	for _, avCard := range availableCards {
		if avCard.Suit == trumpSuit && minimalValueCard.Suit != trumpSuit {
			minimalValueCard = avCard
			continue
		}
		if avCard.getValueIndex() < minimalValueCard.getValueIndex() {
			minimalValueCard = avCard
		}
	}

	attackActionData := AttackActionData{Card: minimalValueCard}
	b.botClient.sendGameAction(PlayerActionNameAttack, attackActionData)
	b.waitingForMyCardBeaten = true
	log.Printf("BOT: attack with %+v", minimalValueCard)
}

func (b *Bot) makeDecision() {
	if !b.isGameStateValid() {
		log.Println("BOT: Invalid state for decision")
		return
	}
	if b.canAttack() {
		b.attack()
		return
	} else {
		log.Println("BOT: Can't attack")
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
