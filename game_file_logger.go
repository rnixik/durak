package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

type GameFileLogger struct {
	dir     string
	buffers map[string]string
}

func NewGameFileLogger(dir string) *GameFileLogger {
	_ = os.MkdirAll(dir, 0777)
	return &GameFileLogger{
		dir:     dir,
		buffers: make(map[string]string, 0),
	}
}

func (l *GameFileLogger) LogGameBegins(game *Game) {
	l.buffers[game.id] = ""
	lines := fmt.Sprintf("ENTRY Game begins. ID=%s\n", game.id)
	lines += getCurrentStateAsLines(game)
	log.Println(lines)
	l.buffers[game.id] += lines
}

func (l *GameFileLogger) LogPlayerActionAttack(game *Game, data AttackActionData) {
	lines := fmt.Sprintf("ENTRY Attack. card=%s%s;\n", data.Card.Value, data.Card.Suit)
	lines += getCurrentStateAsLines(game)
	log.Println(lines)
	l.buffers[game.id] += lines
}

func (l *GameFileLogger) LogPlayerActionDefend(game *Game, data DefendActionData) {
	lines := fmt.Sprintf(
		"Defend. attackingCard=%s%s; defendingCard=%s%s\n",
		data.AttackingCard.Value,
		data.AttackingCard.Suit,
		data.DefendingCard.Value,
		data.DefendingCard.Suit,
	)
	lines += getCurrentStateAsLines(game)
	log.Println(lines)
	l.buffers[game.id] += lines
}

func (l *GameFileLogger) LogPlayerActionPickUp(game *Game) {
	lines := fmt.Sprintf("ENTRY PickUp.\n")
	lines += getCurrentStateAsLines(game)
	log.Println(lines)
	l.buffers[game.id] += lines
}

func (l *GameFileLogger) LogPlayerActionComplete(game *Game) {
	lines := fmt.Sprintf("ENTRY Complete.\n")
	lines += getCurrentStateAsLines(game)
	log.Println(lines)
	l.buffers[game.id] += lines
}

func (l *GameFileLogger) LogGameEnds(game *Game, hasLoser bool, loserIndex int) error {
	lines := fmt.Sprintf("ENTRY Game ends. hasLoser=%t;loserIndex=%d\n", hasLoser, loserIndex)
	lines += getCurrentStateAsLines(game)
	log.Println(lines)
	l.buffers[game.id] += lines

	return l.write(game)
}

func (l *GameFileLogger) write(game *Game) error {
	currentTime := time.Now()
	dir := l.dir + "/" + fmt.Sprintf("%d%02d", currentTime.Year(), currentTime.Month())
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		return err
	}

	f, err := os.Create(dir + "/" + game.id + ".log")
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.WriteString(l.buffers[game.id])
	delete(l.buffers, game.id)
	if err != nil {
		return err
	}

	return nil
}

func getCurrentStateAsLines(game *Game) string {
	line := "State:\n"
	line += fmt.Sprintf("players=%d %s", len(game.players), getPlayersCards(game.players))
	line += "\n"
	line += fmt.Sprintf("deck=%d:%s;", len(game.deck.cards), cardsToString(game.deck.cards))
	line += fmt.Sprintf("battleground=%d:%s;", len(game.battleground), cardsToString(game.battleground))
	line += fmt.Sprintf("attacker=%d;", game.attackerIndex)
	line += fmt.Sprintf("defender=%d;", game.defenderIndex)
	line += fmt.Sprintf("trump=%s%s;", game.trumpCard.Value, game.trumpCard.Suit)
	line += "\n"
	return line
}

func getPlayersCards(players []*Player) string {
	str := ""

	for index, player := range players {
		if index != 0 {
			str += " "
		}
		isBot := "human"
		if strings.HasPrefix(player.Name, "bot-") {
			isBot = "bot"
		}

		str += fmt.Sprintf("P=%d(%s):%s;", index, isBot, cardsToString(player.cards))
	}

	return str
}

func cardsToString(cards []*Card) string {
	str := ""

	for _, card := range cards {
		str += card.Value + card.Suit
	}

	return str
}
