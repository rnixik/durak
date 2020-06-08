package main

import "encoding/json"

const (
	// ClientCommandTypeLobby namespace for commands about a lobby
	ClientCommandTypeLobby = "lobby"
	// ClientCommandLobbySubTypeJoin join the lobby
	ClientCommandLobbySubTypeJoin = "join"
	// ClientCommandLobbySubTypeCreateRoom create a room
	ClientCommandLobbySubTypeCreateRoom = "createRoom"
	// ClientCommandLobbySubTypeJoinRoom join the room
	ClientCommandLobbySubTypeJoinRoom = "joinRoom"

	// ClientCommandTypeGame namespace for commands about a game
	ClientCommandTypeGame = "game"
	// ClientCommandGameSubTypeAttack attack command in game
	ClientCommandGameSubTypeAttack = "attack"
	// ClientCommandGameSubTypeDefend defend command in game
	ClientCommandGameSubTypeDefend = "defend"
	// ClientCommandGameSubTypePickUp pick up cards from desk in game
	ClientCommandGameSubTypePickUp = "pickUp"
	// ClientCommandGameSubTypeComplete complete round in game
	ClientCommandGameSubTypeComplete = "complete"

	// ClientCommandTypeRoom namespace for commands in room
	ClientCommandTypeRoom = "room"
	// ClientCommandRoomSubTypeWantToPlay command to show intention to play the game in room
	ClientCommandRoomSubTypeWantToPlay = "wantToPlay"
	// ClientCommandRoomSubTypeWantToSpectate command to show intention to spectate game only in room
	ClientCommandRoomSubTypeWantToSpectate = "wantToSpectate"
	// ClientCommandRoomSubTypeSetPlayerStatus command to set the status of a player in a room by room owner
	ClientCommandRoomSubTypeSetPlayerStatus = "setPlayerStatus"
	// ClientCommandRoomSubTypeStartGame command to start the game in the room
	ClientCommandRoomSubTypeStartGame = "startGame"
	// ClientCommandRoomSubTypeDeleteGame command to delete the game in the room
	ClientCommandRoomSubTypeDeleteGame = "deleteGame"
	// ClientCommandRoomSubTypeAddBot command to add a bot to the game
	ClientCommandRoomSubTypeAddBot = "addBot"
	// ClientCommandRoomSubTypeRemoveBots command to remove all bots from the game
	ClientCommandRoomSubTypeRemoveBots = "removeBots"
)

// ClientCommand is a command message from connected client.
type ClientCommand struct {
	Type    string          `json:"type"`
	SubType string          `json:"subType"`
	Data    json.RawMessage `json:"data"`
	client  *Client
}
