package main

import "math/rand"

type Bot struct {
	nickname string
	id       uint64
	room     *Room
}

func newBot(id uint64) *Bot {
	return &Bot{
		nickname: generateBotName(),
		id:       id,
	}
}

func (b *Bot) sendEvent(event interface{}) {
	jsonEvent, _ := eventToJSON(event)
	b.sendMessage(jsonEvent)
}

func (b *Bot) sendMessage(message []byte) {

}

// Nickname returns nickname of the bot
func (b *Bot) Nickname() string {
	return b.nickname
}

// Id returns id of the bot
func (b *Bot) Id() uint64 {
	return b.id
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
