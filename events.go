package main

import (
	"encoding/json"
	"reflect"
)

const (
	errorNeedOneMorePlayer                  = "need_one_more_player"
	errorNumberOfPlayersExceededLimit       = "number_of_players_exceeded_limit"
	errorGameHasBeenAlreadyStarted          = "game_has_been_already_started"
	errorYouCanCreateOneRoomOnly            = "you_can_create_one_room_only"
	errorRoomDoesNotExist                   = "room_does_not_exist"
	errorCantChangeStatusGameHasBeenStarted = "cant_change_status_game_has_been_started"
)

// JSONEvent represents a message to clients with some event.
type JSONEvent struct {
	Name string      `json:"name"`
	Data interface{} `json:"data"`
}

// ClientCommandError contains info about error on client's command.
type ClientCommandError struct {
	Message string `json:"message"`
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
