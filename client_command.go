package main

import "encoding/json"

const (
	ClientCommandTypeLobby              = "lobby"
	ClientCommandLobbySubTypeJoin       = "join"
	ClientCommandLobbySubTypeCreateRoom = "createRoom"
	ClientCommandLobbySubTypeJoinRoom   = "joinRoom"

	ClientCommandTypeGame            = "game"
	ClientCommandGameSubTypeAttack   = "attack"
	ClientCommandGameSubTypeDefend   = "defend"
	ClientCommandGameSubTypePickUp   = "pickUp"
	ClientCommandGameSubTypeComplete = "complete"

	ClientCommandTypeRoom                   = "room"
	ClientCommandRoomSubTypeWantToPlay      = "wantToPlay"
	ClientCommandRoomSubTypeWantToSpectate  = "wantToSpectate"
	ClientCommandRoomSubTypeSetPlayerStatus = "setPlayerStatus"
	ClientCommandRoomSubTypeStartGame       = "startGame"
	ClientCommandRoomSubTypeDeleteGame      = "deleteGame"
	ClientCommandRoomSubTypeAddBot          = "addBot"
	ClientCommandRoomSubTypeRemoveBots      = "removeBots"
)

// ClientCommand is a command message from connected client.
type ClientCommand struct {
	Type    string          `json:"type"`
	SubType string          `json:"subType"`
	Data    json.RawMessage `json:"data"`
	client  *Client
}
