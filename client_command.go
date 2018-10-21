package main

import "encoding/json"

const (
	ClientCommandTypeLobby              = "lobby"
	ClientCommandLobbySubTypeJoin       = "join"
	ClientCommandLobbySubTypeCreateRoom = "create_room"
	ClientCommandLobbySubTypeJoinRoom   = "join_room"

	ClientCommandTypeGame          = "game"
	ClientCommandGameSubTypeAttack = "attack"

	ClientCommandTypeRoom = "room"
)

// ClientCommand is a command message from connected client.
type ClientCommand struct {
	Type    string          `json:"type"`
	SubType string          `json:"sub_type"`
	Data    json.RawMessage `json:"data"`
	client  *Client
}
