package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// GameFileLogger is implementation of GameLogger which stores logs in files
type GameFileLogger struct {
	dir         string
	errCallback func(err error)
	bufferChans map[string]chan string
	stopChans   map[string]chan bool
}

// NewGameFileLogger constructor for GameFileLogger
func NewGameFileLogger(dir string, errCallback func(err error)) *GameFileLogger {
	_ = os.MkdirAll(dir, 0777)
	return &GameFileLogger{
		dir:         dir,
		errCallback: errCallback,
		bufferChans: make(map[string]chan string, 0),
		stopChans:   make(map[string]chan bool, 0),
	}
}

// LogGameBegins starts recording log and adds entry about beginning the game
func (l *GameFileLogger) LogGameBegins(game *Game) {
	l.bufferChans[game.id] = make(chan string)
	l.stopChans[game.id] = make(chan bool)
	go l.writeLoop(game.id)

	lines := fmt.Sprintf("ENTRY Game begins. ID=%s\n", game.id)
	lines += getCurrentStateAsLines(game)
	l.bufferChans[game.id] <- lines
}

// LogPlayerActionAttack adds entry about attack
func (l *GameFileLogger) LogPlayerActionAttack(game *Game, data AttackActionData) {
	lines := fmt.Sprintf("ENTRY Attack. card=%s%s;\n", data.Card.Value, data.Card.Suit)
	lines += getCurrentStateAsLines(game)
	l.bufferChans[game.id] <- lines
}

// LogPlayerActionDefend adds entry about defense
func (l *GameFileLogger) LogPlayerActionDefend(game *Game, data DefendActionData) {
	lines := fmt.Sprintf(
		"Defend. attackingCard=%s%s; defendingCard=%s%s\n",
		data.AttackingCard.Value,
		data.AttackingCard.Suit,
		data.DefendingCard.Value,
		data.DefendingCard.Suit,
	)
	lines += getCurrentStateAsLines(game)
	l.bufferChans[game.id] <- lines
}

// LogPlayerActionPickUp adds entry about pick up
func (l *GameFileLogger) LogPlayerActionPickUp(game *Game) {
	lines := fmt.Sprintf("ENTRY PickUp.\n")
	lines += getCurrentStateAsLines(game)
	l.bufferChans[game.id] <- lines
}

// LogPlayerActionComplete adds entry about completing a round
func (l *GameFileLogger) LogPlayerActionComplete(game *Game) {
	lines := fmt.Sprintf("ENTRY Complete.\n")
	lines += getCurrentStateAsLines(game)
	l.bufferChans[game.id] <- lines
}

// LogGameEnds adds entry about ending game and writes file with entries
func (l *GameFileLogger) LogGameEnds(game *Game, hasLoser bool, loserIndex int) {
	lines := fmt.Sprintf("ENTRY Game ends. hasLoser=%t;loserIndex=%d\n", hasLoser, loserIndex)
	lines += getCurrentStateAsLines(game)
	l.bufferChans[game.id] <- lines
	l.stopChans[game.id] <- true
}

func (l *GameFileLogger) write(gameId string, contents string) error {
	currentTime := time.Now()
	dir := l.dir + "/" + fmt.Sprintf("%d%02d", currentTime.Year(), currentTime.Month())
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		return err
	}

	f, err := os.Create(dir + "/" + gameId + ".log")
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.WriteString(contents)
	if err != nil {
		return err
	}

	return nil
}

func (l *GameFileLogger) writeLoop(gameId string) {
	contents := ""

	defer func() {
		close(l.stopChans[gameId])

		err := l.write(gameId, contents)
		if err != nil {
			l.errCallback(err)
		}
	}()

	for {
		select {
		case lines, ok := <-l.bufferChans[gameId]:
			if !ok {
				return
			}
			contents += lines
		case <-l.stopChans[gameId]:
			close(l.bufferChans[gameId])
		}
	}
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
