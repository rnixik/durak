package main

import (
	"encoding/json"
	"log"
)

type Bot struct {
	botClient                  *BotClient
	lastGamePlayersEvent       *GamePlayersEvent
	lastGameFirstAttackerEvent *GameFirstAttackerEvent
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
	}

	if err != nil {
		log.Printf("error at parsing event data: %s", err)
		return
	}

	b.makeDecision()
}

func (b *Bot) onGamePlayersEvent(event GamePlayersEvent) {
	b.lastGamePlayersEvent = &event
}

func (b *Bot) onGameFirstAttackerEvent(event GameFirstAttackerEvent) {
	b.lastGameFirstAttackerEvent = &event
}

func (b *Bot) makeDecision() {
	if b.lastGamePlayersEvent == nil {
		return
	}
	if b.lastGameFirstAttackerEvent == nil {
		return
	}

	if b.lastGameFirstAttackerEvent.AttackerIndex == b.lastGamePlayersEvent.YourPlayerIndex {
		log.Printf("Bot is attacker")
	} else {
		log.Printf("Bot is not attacker")
	}
}
