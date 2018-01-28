package main

import (
	"log"
)

// MaxPlayersInRoom limits maximum number of players in room
const MaxPlayersInRoom = 6

// RoomMember represents connected to a room client.
type RoomMember struct {
	client     ClientSender
	wantToPlay bool
}

// Room represents place where some of members want to start a new game.
type Room struct {
	id      uint64
	owner   *RoomMember
	members map[*RoomMember]bool
	game    *Game
}

func newRoom(roomId uint64, owner *Client) *Room {
	members := make(map[*RoomMember]bool, 0)
	ownerInRoom := newRoomMember(owner)
	members[ownerInRoom] = true
	room := &Room{roomId, ownerInRoom, members, nil}
	owner.room = room

	roomJoinedEvent := RoomJoinedEvent{room.toRoomInfo()}
	owner.sendEvent(roomJoinedEvent)

	return room
}

func newRoomMember(client ClientSender) *RoomMember {
	return &RoomMember{client, true}
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

func (r *Room) onWantToPlayCommand(cc *ClientCommand) {
	r.changeMemberWantStatus(cc.client, true)
}

func (r *Room) onWantToSpectateCommand(cc *ClientCommand) {
	r.changeMemberWantStatus(cc.client, false)
}

func (r *Room) onClientCommand(cc *ClientCommand) {
	log.Println(cc.SubType)
	if cc.SubType == "want_to_play" {
		r.onWantToPlayCommand(cc)
	} else if cc.SubType == "want_to_spectate" {
		r.onWantToSpectateCommand(cc)
	}
}

func (rm *RoomMember) memberToRoomMemberInfo() *RoomMemberInfo {
	return &RoomMemberInfo{
		Id:         rm.client.Id(),
		Nickname:   rm.client.Nickname(),
		WantToPlay: rm.wantToPlay,
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
	}
	return roomInfo
}
