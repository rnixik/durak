package main

import (
	"encoding/json"
	"log"
	"sync/atomic"
)

// MaxPlayersInRoom limits maximum number of players in room
const MaxPlayersInRoom = 2

var lastGameId uint64

// RoomMember represents connected to a room client.
type RoomMember struct {
	client     ClientSender
	wantToPlay bool
	isPlayer   bool
}

// Room represents place where some of members want to start a new game.
type Room struct {
	id      uint64
	owner   *RoomMember
	members map[*RoomMember]bool
	game    *Game
	lobby   *Lobby
}

func newRoom(roomId uint64, owner *Client, lobby *Lobby) *Room {
	members := make(map[*RoomMember]bool, 0)
	ownerInRoom := newRoomMember(owner)
	ownerInRoom.isPlayer = true
	members[ownerInRoom] = true
	room := &Room{roomId, ownerInRoom, members, nil, lobby}
	owner.room = room

	roomJoinedEvent := RoomJoinedEvent{room.toRoomInfo()}
	owner.sendEvent(roomJoinedEvent)

	return room
}

func newRoomMember(client ClientSender) *RoomMember {
	return &RoomMember{client, true, false}
}

// Name returns name of the room by its owner.
func (r *Room) Name() string {
	return r.owner.client.Nickname()
}

// Id returns id of the room
func (r *Room) Id() uint64 {
	return r.id
}

func (r *Room) getRoomMember(client *Client) (*RoomMember, bool) {
	for c := range r.members {
		if c.client.Id() == client.Id() {
			return c, true
		}
	}
	return nil, false
}

func (r *Room) removeClient(client *Client) (changedOwner bool, roomBecameEmpty bool) {
	client.room = nil
	member, ok := r.getRoomMember(client)
	if !ok {
		return
	}
	delete(r.members, member)

	if r.game != nil {
		r.game.onClientRemoved(client)
	}

	roomUpdatedEvent := &RoomUpdatedEvent{r.toRoomInfo()}
	r.broadcastEvent(roomUpdatedEvent, nil)

	if len(r.members) == 0 {
		roomBecameEmpty = true
		return
	}
	if r.owner == member {
		for ic := range r.members {
			r.owner = ic
			changedOwner = true
			return
		}
	}
	return
}

func (r *Room) addClient(client *Client) {
	member := newRoomMember(client)
	r.members[member] = true
	client.room = r

	if len(r.getPlayers()) < 2 {
		member.isPlayer = true
	}

	roomUpdatedEvent := &RoomUpdatedEvent{r.toRoomInfo()}
	r.broadcastEvent(roomUpdatedEvent, member)

	roomJoinedEvent := RoomJoinedEvent{r.toRoomInfo()}
	client.sendEvent(roomJoinedEvent)
}

func (r *Room) broadcastEvent(event interface{}, exceptMember *RoomMember) {
	json, _ := eventToJSON(event)
	for m := range r.members {
		if exceptMember == nil || m.client.Id() != exceptMember.client.Id() {
			m.client.sendMessage(json)
		}
	}
}

func (r *Room) getPlayers() []*RoomMember {
	players := make([]*RoomMember, 0)
	for rm := range r.members {
		if rm.isPlayer {
			players = append(players, rm)
		}
	}
	return players
}

func (r *Room) getMembersWhoWantToPlay() []*RoomMember {
	membersWhoWantToPlay := make([]*RoomMember, 0)
	for rm := range r.members {
		if rm.wantToPlay {
			membersWhoWantToPlay = append(membersWhoWantToPlay, rm)
		}
	}
	return membersWhoWantToPlay
}

func (r *Room) hasSlotForPlayer() bool {
	membersWhoWantToPlayNum := 0
	for rm := range r.members {
		if rm.wantToPlay {
			membersWhoWantToPlayNum++
		}
	}
	return membersWhoWantToPlayNum+1 <= MaxPlayersInRoom
}

func (r *Room) changeMemberWantStatus(client *Client, wantToPlay bool) {
	member, ok := r.getRoomMember(client)
	if !ok {
		return
	}
	member.wantToPlay = wantToPlay
	memberInfo := member.memberToRoomMemberInfo()
	changeStatusEvent := &RoomMemberChangedStatusEvent{memberInfo}
	r.broadcastEvent(changeStatusEvent, nil)
}

func (r *Room) onWantToPlayCommand(client *Client) {
	if r.game != nil {
		errEvent := &ClientCommandError{"Can't change status: game has been started"}
		client.sendEvent(errEvent)
		return
	}
	r.changeMemberWantStatus(client, true)
}

func (r *Room) onWantToSpectateCommand(client *Client) {
	if r.game != nil {
		errEvent := &ClientCommandError{"Can't change status: game has been started"}
		client.sendEvent(errEvent)
		return
	}
	r.changeMemberWantStatus(client, false)
	r.setPlayerStatus(client.Id(), false)
}

func (r *Room) onSetPlayerStatusCommand(c *Client, memberId uint64, playerStatus bool) {
	if r.game != nil {
		errEvent := &ClientCommandError{"Can't change status: game has been started"}
		c.sendEvent(errEvent)
		return
	}
	r.setPlayerStatus(memberId, playerStatus)
}

func (r *Room) setPlayerStatus(memberId uint64, playerStatus bool) {
	var foundMember *RoomMember
	for rm := range r.members {
		if rm.client.Id() == memberId {
			rm.isPlayer = playerStatus
			foundMember = rm
			break
		}
	}

	if foundMember == nil {
		return
	}

	memberInfo := foundMember.memberToRoomMemberInfo()
	roomMemberChangedPlayerStatusEvent := &RoomMemberChangedPlayerStatusEvent{memberInfo}
	r.broadcastEvent(roomMemberChangedPlayerStatusEvent, nil)
}

func (r *Room) onStartGameCommand(c *Client) {
	pls := r.getPlayers()
	if len(pls) < 2 {
		errEvent := &ClientCommandError{"Need 1 more player"}
		c.sendEvent(errEvent)
		return
	}
	if len(pls) > MaxPlayersInRoom {
		errEvent := &ClientCommandError{"Number of players exceeded limit"}
		c.sendEvent(errEvent)
		return
	}
	if r.game != nil {
		errEvent := &ClientCommandError{"Game has been already started"}
		c.sendEvent(errEvent)
		return
	}

	players := make([]*Player, 0)
	for rm := range r.members {
		player := newPlayer(rm.client, rm.isPlayer)
		players = append(players, player)
	}

	atomic.AddUint64(&lastGameId, 1)
	lastGameIdSafe := atomic.LoadUint64(&lastGameId)

	r.game = newGame(lastGameIdSafe, r, players[0], players)
	go r.game.begin()

	roomUpdatedEvent := &RoomUpdatedEvent{r.toRoomInfo()}
	r.broadcastEvent(roomUpdatedEvent, nil)

	r.lobby.sendRoomUpdate(r)
}

func (r *Room) onClientCommand(cc *ClientCommand) {
	log.Println(cc.SubType)
	if cc.SubType == "want_to_play" {
		r.onWantToPlayCommand(cc.client)
	} else if cc.SubType == "want_to_spectate" {
		r.onWantToSpectateCommand(cc.client)
	} else if cc.SubType == "set_player_status" {
		var statusData RoomSetPlayerStatusCommandData
		if err := json.Unmarshal(cc.Data, &statusData); err != nil {
			return
		}
		r.onSetPlayerStatusCommand(cc.client, statusData.MemberId, statusData.Status)
	} else if cc.SubType == "start_game" {
		r.onStartGameCommand(cc.client)
	}
}

func (rm *RoomMember) memberToRoomMemberInfo() *RoomMemberInfo {
	return &RoomMemberInfo{
		Id:         rm.client.Id(),
		Nickname:   rm.client.Nickname(),
		WantToPlay: rm.wantToPlay,
		IsPlayer:   rm.isPlayer,
	}
}

func (r *Room) toRoomInList() *RoomInList {
	gameStatus := ""
	if r.game != nil {
		gameStatus = r.game.status
	}
	roomInList := &RoomInList{
		Id:         r.Id(),
		OwnerId:    r.owner.client.Id(),
		Name:       r.Name(),
		GameStatus: gameStatus,
		MembersNum: len(r.members),
	}
	return roomInList
}

func (r *Room) toRoomInfo() *RoomInfo {
	gameStatus := ""
	if r.game != nil {
		gameStatus = r.game.status
	}

	membersInfo := make([]*RoomMemberInfo, 0)
	for member := range r.members {
		memberInfo := member.memberToRoomMemberInfo()
		membersInfo = append(membersInfo, memberInfo)
	}

	roomInfo := &RoomInfo{
		Id:         r.Id(),
		OwnerId:    r.owner.client.Id(),
		Name:       r.Name(),
		GameStatus: gameStatus,
		Members:    membersInfo,
		MaxPlayers: MaxPlayersInRoom,
	}
	return roomInfo
}
