package main

import (
	"encoding/json"
	"reflect"
)

// JSONEvent represents a message to clients with some event.
type JSONEvent struct {
	Name string      `json:"name"`
	Data interface{} `json:"data"`
}

// RoomInList contains short info about room in the lobby.
type RoomInList struct {
	Id         uint64 `json:"id"`
	OwnerId    uint64 `json:"owner_id"`
	Name       string `json:"name"`
	GameStatus string `json:"game_status"`
}

// ClientInList contains short info about client in the lobby.
type ClientInList struct {
	Id       uint64 `json:"id"`
	Nickname string `json:"nickname"`
}

// ClientJoinedEvent contains info for the just connected client.
type ClientJoinedEvent struct {
	YourId       uint64          `json:"your_id"`
	YourNickname string          `json:"your_nickname"`
	Clients      []*ClientInList `json:"clients"`
	Rooms        []*RoomInList   `json:"rooms"`
}

// ClientLeftEvent contains id of client who left lobby.
type ClientLeftEvent struct {
	Id uint64 `json:"id"`
}

// ClientBroadCastJoinedEvent contains info for other clients when a new client was connected.
type ClientBroadCastJoinedEvent struct {
	Id       uint64 `json:"id"`
	Nickname string `json:"nickname"`
}

// ClientCommandError contains info about error on client's command.
type ClientCommandError struct {
	Message string `json:"message"`
}

// ClientCreatedRoomEvent contains info of created room.
type ClientCreatedRoomEvent struct {
	Room *RoomInList `json:"room"`
}

// PlayersEvent contains list of players which were connected to a game.
type PlayersEvent struct {
	YourPlayerIndex int       `json:"your_player_index"`
	Players         []*Player `json:"players"`
}

// DealEvent contains info about game after the deal. It includes list of cards for each player.
type DealEvent struct {
	YourHand                      []*Card `json:"your_hand"`
	HandsSizes                    []int   `json:"hands_sizes"`
	PileSize                      int     `json:"pile_size"`
	TrumpSuit                     string  `json:"trump_suit"`
	TrumpCard                     *Card   `json:"trump_card"`
	TrumpCardIsInPile             bool    `json:"trump_card_is_in_pile"`
	TrumpCardIsOwnedByPlayerIndex int     `json:"trump_card_is_owned_by_player_index"`
}

// FirstAttackerEvent contains info who is the first attacker and why.
type FirstAttackerEvent struct {
	ReasonCard    *Card `json:"reason_card"`
	AttackerIndex int   `json:"attacker_index"`
}

func eventToJSON(e interface{}) ([]byte, error) {
	name := getNameOfStruct(e)
	jsonEvent := JSONEvent{Name: name, Data: e}
	return json.Marshal(jsonEvent)
}

func getNameOfStruct(s interface{}) string {
	t := reflect.TypeOf(s)
	if t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	}
	return t.Name()
}
