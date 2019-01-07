package main

import (
	"math/rand"
)

type BotClient struct {
	nickname string
	id       uint64
	room     *Room

	incomingEvents chan []byte
}

func newBotClient(id uint64, room *Room) *BotClient {
	botClient := &BotClient{
		nickname:       generateBotName(),
		id:             id,
		room:           room,
		incomingEvents: make(chan []byte),
	}

	bot := newBot(botClient)
	go bot.run()

	return botClient
}

func (bl *BotClient) sendEvent(event interface{}) {
	jsonEvent, _ := eventToJSON(event)
	bl.sendMessage(jsonEvent)
}

func (bl *BotClient) sendMessage(message []byte) {
	bl.incomingEvents <- message
}

// Nickname returns nickname of the bot
func (bl *BotClient) Nickname() string {
	return bl.nickname
}

// Id returns id of the bot
func (bl *BotClient) Id() uint64 {
	return bl.id
}

func generateBotName() string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const botNameLength = 7
	b := make([]byte, botNameLength)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}

	return "bot-" + string(b)
}
