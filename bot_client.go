package main

import (
	"log"
	"math/rand"
)

// BotClient represents a connection to the game for an AI player
type BotClient struct {
	nickname string
	id       uint64
	room     *Room

	incomingEvents  chan []byte
	outgoingActions chan *PlayerAction
}

func newBotClient(id uint64, room *Room) *BotClient {
	botClient := &BotClient{
		nickname:        generateBotName(),
		id:              id,
		room:            room,
		incomingEvents:  make(chan []byte),
		outgoingActions: make(chan *PlayerAction),
	}

	bot := newBot(botClient)
	go botClient.sendingActionsToGame()
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

func (bl *BotClient) sendGameAction(playerActionName string, actionData interface{}) {
	game := bl.room.game
	if game.status != GameStatusPlaying {
		log.Printf("BOT: Cannot send game action - wrong status of game = %s", game.status)
		return
	}
	var player *Player
	for _, gamePlayer := range game.players {
		if gamePlayer.client.Id() == bl.Id() {
			player = gamePlayer
			break
		}
	}
	if player == nil {
		return
	}
	playerAction := &PlayerAction{Name: playerActionName, Data: actionData, player: player}
	log.Printf("BOT: sending game action to game %+v", playerAction)
	bl.outgoingActions <- playerAction
}
func (bl *BotClient) sendingActionsToGame() {
	for {
		select {
		case playerAction := <-bl.outgoingActions:
			bl.room.game.playerActions <- playerAction
		}
	}
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
