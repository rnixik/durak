package main

import (
	"encoding/json"
	"log"
)

type Bot struct {
	botClient            *BotClient
	lastGamePlayersEvent *GamePlayersEvent
}

func newBot(botClient *BotClient) *Bot {
	return &Bot{
		botClient: botClient,
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
	switch event.Name {
	case "GamePlayersEvent":
		type GamePlayersEventBot struct {
			Event GamePlayersEvent `json:"Data"`
		}
		var parsedEvent GamePlayersEventBot
		err = json.Unmarshal(message, &parsedEvent)
		if err != nil {
			return
		}
		b.onGamePlayersEvent(parsedEvent.Event)
	}

	b.makeDecision()
}

func (b *Bot) onGamePlayersEvent(event GamePlayersEvent) {
	b.lastGamePlayersEvent = &event
}

func (b *Bot) makeDecision() {
	if b.lastGamePlayersEvent == nil {
		return
	}

	log.Printf("My index is %d", b.lastGamePlayersEvent.YourPlayerIndex)
}
