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
	ClientCommandGameSubTypePickUp   = "pick_up"
	ClientCommandGameSubTypeComplete = "complete"

	ClientCommandTypeRoom                   = "room"
	ClientCommandRoomSubTypeWantToPlay      = "wantToPlay"
	ClientCommandRoomSubTypeWantToSpectate  = "wantToSpectate"
	ClientCommandRoomSubTypeSetPlayerStatus = "setPlayerStatus"
	ClientCommandRoomSubTypeStartGame       = "startGame"
)

// ClientCommand is a command message from connected client.
type ClientCommand struct {
	Type    string          `json:"type"`
	SubType string          `json:"sub_type"`
	Data    json.RawMessage `json:"data"`
	client  *Client
}
