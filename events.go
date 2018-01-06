package main

import (
	"encoding/json"
	"reflect"
)

type JsonEvent struct {
	Name string      `json:"name"`
	Data interface{} `json:"data"`
}

type GameInList struct {
	Name string `json:"name"`
	Id   uint64 `json:"id"`
}

type ClientJoinedEvent struct {
	Id         uint64        `json:"id"`
	Nickname   string        `json:"nickname"`
	ClientsNum int           `json:"clients_num"`
	Games      []*GameInList `json:"games"`
}

type ClientBroadCastJoinedEvent struct {
	Id         uint64 `json:"id"`
	Nickname   string `json:"nickname"`
	ClientsNum int    `json:"clients_num"`
}

type PlayersEvent struct {
	YourPlayerIndex int       `json:"your_player_index"`
	Players         []*Player `json:"players"`
}

type DealEvent struct {
	YourHand                      []*Card `json:"your_hand"`
	HandsSizes                    []int   `json:"hands_sizes"`
	PileSize                      int     `json:"pile_size"`
	TrumpSuit                     string  `json:"trump_suit"`
	TrumpCard                     *Card   `json:"trump_card"`
	TrumpCardIsInPile             bool    `json:"trump_card_is_in_pile"`
	TrumpCardIsOwnedByPlayerIndex int     `json:"trump_card_is_owned_by_player_index"`
}

type FirstAttackerEvent struct {
	ReasonCard    *Card `json:"reason_card"`
	AttackerIndex int   `json:"attacker_index"`
}

func eventToJson(e interface{}) ([]byte, error) {
	name := getNameOfStruct(e)
	jsonEvent := JsonEvent{Name: name, Data: e}
	return json.Marshal(jsonEvent)
}

func getNameOfStruct(s interface{}) string {
	if t := reflect.TypeOf(s); t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	} else {
		return t.Name()
	}
}
