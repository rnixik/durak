package main

import (
	"encoding/json"
	"log"
	"sync/atomic"
)

// MaxPlayersInRoom limits maximum number of players in room
const MaxPlayersInRoom = 6

// RoomMember represents connected to a room client.
type RoomMember struct {
	client     ClientSender
	wantToPlay bool
	isPlayer   bool
	isBot      bool
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
	ownerInRoom := newRoomMember(owner, false)
	ownerInRoom.isPlayer = true
	members[ownerInRoom] = true
	room := &Room{roomId, ownerInRoom, members, nil, lobby}
	owner.room = room

	return room
}

func newRoomMember(client ClientSender, isBot bool) *RoomMember {
	return &RoomMember{client, true, false, isBot}
}

// Name returns name of the room by its owner.
func (r *Room) Name() string {
	return r.owner.client.Nickname()
}

// Id returns id of the room
func (r *Room) Id() uint64 {
	return r.id
}

func (r *Room) getRoomMember(client ClientSender) (*RoomMember, bool) {
	for c := range r.members {
		if c.client.Id() == client.Id() {
			return c, true
		}
	}
	return nil, false
}

func (r *Room) removeClient(client ClientSender) (changedOwner bool, roomBecameEmpty bool) {
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

	nonBotsMembersNumber := 0
	for ic := range r.members {
		if !ic.isBot {
			nonBotsMembersNumber += 1
		}
	}

	if nonBotsMembersNumber == 0 {
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
	member := newRoomMember(client, false)
	r.members[member] = true
	client.room = r

	if len(r.getPlayers()) < 2 {
		member.isPlayer = true
	}

	roomUpdatedEvent := &RoomUpdatedEvent{r.toRoomInfo()}
	r.broadcastEvent(roomUpdatedEvent, member.client.(*Client))

	roomJoinedEvent := RoomJoinedEvent{r.toRoomInfo()}
	client.sendEvent(roomJoinedEvent)

	if r.game != nil && r.game.status == GameStatusPlaying {
		player := newPlayer(client, false)
		r.game.players = append(r.game.players, player)
		r.game.onLatePlayerJoin(player)
	}
}

func (r *Room) addBot(bot *BotClient) {
	member := newRoomMember(bot, true)
	r.members[member] = true
	member.isPlayer = true

	roomUpdatedEvent := &RoomUpdatedEvent{r.toRoomInfo()}
	r.broadcastEvent(roomUpdatedEvent, nil)

	roomJoinedEvent := RoomJoinedEvent{r.toRoomInfo()}
	bot.sendEvent(roomJoinedEvent)
}

func (r *Room) broadcastEvent(event interface{}, exceptClient *Client) {
	eventJson, _ := eventToJSON(event)
	for m := range r.members {
		if exceptClient == nil || m.client.Id() != exceptClient.Id() {
			m.client.sendMessage(eventJson)
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
		errEvent := &ClientCommandError{errorCantChangeStatusGameHasBeenStarted}
		client.sendEvent(errEvent)
		return
	}
	r.changeMemberWantStatus(client, true)
}

func (r *Room) onWantToSpectateCommand(client *Client) {
	if r.game != nil {
		errEvent := &ClientCommandError{errorCantChangeStatusGameHasBeenStarted}
		client.sendEvent(errEvent)
		return
	}
	r.changeMemberWantStatus(client, false)
	r.setPlayerStatus(client.Id(), false)
}

func (r *Room) onSetPlayerStatusCommand(c *Client, memberId uint64, playerStatus bool) {
	if r.owner.client.Id() != c.Id() {
		errEvent := &ClientCommandError{errorYouShouldBeOwner}
		c.sendEvent(errEvent)
		return
	}
	if r.game != nil {
		errEvent := &ClientCommandError{errorCantChangeStatusGameHasBeenStarted}
		c.sendEvent(errEvent)
		return
	}

	if playerStatus && !r.hasSlotForPlayer() {
		errEvent := &ClientCommandError{errorNumberOfPlayersExceededLimit}
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
		errEvent := &ClientCommandError{errorNeedOneMorePlayer}
		c.sendEvent(errEvent)
		return
	}
	if len(pls) > MaxPlayersInRoom {
		errEvent := &ClientCommandError{errorNumberOfPlayersExceededLimit}
		c.sendEvent(errEvent)
		return
	}
	if r.game != nil {
		errEvent := &ClientCommandError{errorGameHasBeenAlreadyStarted}
		c.sendEvent(errEvent)
		return
	}

	players := make([]*Player, 0)
	for rm := range r.members {
		if rm.isPlayer {
			player := newPlayer(rm.client, rm.isPlayer)
			players = append(players, player)
		}
	}

	r.game = newGame(r, players, r.lobby.gameLogger)
	go r.game.begin()

	roomUpdatedEvent := &RoomUpdatedEvent{r.toRoomInfo()}
	r.broadcastEvent(roomUpdatedEvent, nil)

	r.lobby.sendRoomUpdate(r)
}

func (r *Room) onDeleteGameCommand(c *Client) {
	if r.owner.client.Id() != c.Id() {
		errEvent := &ClientCommandError{errorYouShouldBeOwner}
		c.sendEvent(errEvent)
		return
	}
	if r.game == nil {
		errEvent := &ClientCommandError{errorGameAlreadyDeleted}
		c.sendEvent(errEvent)
		return
	}

	r.game = nil

	roomUpdatedEvent := &RoomUpdatedEvent{r.toRoomInfo()}
	r.broadcastEvent(roomUpdatedEvent, nil)

	r.lobby.sendRoomUpdate(r)
}

func (r *Room) onAddBotCommand(c *Client) {
	if r.owner.client.Id() != c.Id() {
		errEvent := &ClientCommandError{errorYouShouldBeOwner}
		c.sendEvent(errEvent)
		return
	}
	if r.game != nil {
		errEvent := &ClientCommandError{errorGameHasBeenAlreadyStarted}
		c.sendEvent(errEvent)
		return
	}
	if !r.hasSlotForPlayer() {
		errEvent := &ClientCommandError{errorNumberOfPlayersExceededLimit}
		c.sendEvent(errEvent)
		return
	}

	atomic.AddUint64(&lastClientId, 1)
	lastBotIdSafe := atomic.LoadUint64(&lastClientId)
	bot := newBotClient(lastBotIdSafe, r)
	r.addBot(bot)
}

func (r *Room) onRemoveBotsCommand(c *Client) {
	if r.owner.client.Id() != c.Id() {
		errEvent := &ClientCommandError{errorYouShouldBeOwner}
		c.sendEvent(errEvent)
		return
	}
	if r.game != nil {
		errEvent := &ClientCommandError{errorGameHasBeenAlreadyStarted}
		c.sendEvent(errEvent)
		return
	}

	for rm := range r.members {
		if rm.isBot {
			r.removeClient(rm.client)
		}
	}
}

func (r *Room) onClientCommand(cc *ClientCommand) {
	log.Println(cc.SubType)
	switch cc.SubType {
	case ClientCommandRoomSubTypeWantToPlay:
		r.onWantToPlayCommand(cc.client)
	case ClientCommandRoomSubTypeWantToSpectate:
		r.onWantToSpectateCommand(cc.client)
	case ClientCommandRoomSubTypeSetPlayerStatus:
		var statusData RoomSetPlayerStatusCommandData
		if err := json.Unmarshal(cc.Data, &statusData); err != nil {
			return
		}
		r.onSetPlayerStatusCommand(cc.client, statusData.MemberId, statusData.Status)
	case ClientCommandRoomSubTypeStartGame:
		r.onStartGameCommand(cc.client)
	case ClientCommandRoomSubTypeDeleteGame:
		r.onDeleteGameCommand(cc.client)
	case ClientCommandRoomSubTypeAddBot:
		r.onAddBotCommand(cc.client)
	case ClientCommandRoomSubTypeRemoveBots:
		r.onRemoveBotsCommand(cc.client)
	}
}

func (r *Room) onGameStarted() {
	roomUpdatedEvent := &RoomUpdatedEvent{r.toRoomInfo()}
	r.broadcastEvent(roomUpdatedEvent, nil)
	r.lobby.sendRoomUpdate(r)
}

func (r *Room) onGameEnded() {
	roomUpdatedEvent := &RoomUpdatedEvent{r.toRoomInfo()}
	r.broadcastEvent(roomUpdatedEvent, nil)
	r.lobby.sendRoomUpdate(r)
}

func (rm *RoomMember) memberToRoomMemberInfo() *RoomMemberInfo {
	return &RoomMemberInfo{
		Id:         rm.client.Id(),
		Nickname:   rm.client.Nickname(),
		WantToPlay: rm.wantToPlay,
		IsPlayer:   rm.isPlayer,
		IsBot:      rm.isBot,
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
