package main

import (
	"testing"
)

func TestNewRoom(t *testing.T) {
	client := &Client{nickname: "test_nickname"}
	room := newRoom(1, client)
	got := room.owner.client
	expected := client
	if got != expected {
		t.Errorf("TestNewRoom expected: %v, got: %v", expected, got)
	}
	membersNum := len(room.members)
	if membersNum != 1 {
		t.Errorf("TestNewRoom expected members: 1, got: %v", membersNum)
	}
}

func TestName(t *testing.T) {
	client := &Client{nickname: "test_nickname"}
	room := newRoom(1, client)
	got := room.Name()
	expected := "test_nickname"
	if got != expected {
		t.Errorf("TestName expected: %v, got: %v", expected, got)
	}
}

func TestId(t *testing.T) {
	client := &Client{id: 123}
	room := newRoom(12345, client)
	got := room.Id()
	expected := uint64(12345)
	if got != expected {
		t.Errorf("TestId expected: %v, got: %v", expected, got)
	}
}

func TestToRoomInList(t *testing.T) {
	client := &Client{id: 123, nickname: "test_nickname"}
	room := newRoom(1, client)
	roomInList := room.toRoomInList()
	got := getNameOfStruct(roomInList)
	expected := "RoomInList"
	if got != expected {
		t.Errorf("TestToRoomInList expected: %v, got: %v", expected, got)
	}
}

func TestToRoomInfo(t *testing.T) {
	client := &Client{id: 123, nickname: "test_nickname"}
	room := newRoom(1, client)
	roomInfo := room.toRoomInfo()
	got := getNameOfStruct(roomInfo)
	expected := "RoomInfo"
	if got != expected {
		t.Errorf("TestToRoomInfo expected: %v, got: %v", expected, got)
	}
}

func TestAddClient(t *testing.T) {
	client := &Client{id: 123, nickname: "test_nickname"}
	room := newRoom(1, client)
	client2 := &Client{id: 456, nickname: "test_nickname2"}
	room.addClient(client2)
	got := len(room.members)
	expected := 2
	if got != expected {
		t.Errorf("TestAddClient expected: %v, got: %v", expected, got)
	}
}

func TestRemoveLastClient(t *testing.T) {
	client := &Client{id: 123, nickname: "test_nickname"}
	room := newRoom(1, client)
	_, roomBecameEmpty := room.removeClient(client)
	if !roomBecameEmpty {
		t.Errorf("TestRemoveLastClient expected that room became empty")
	}
}

func TestRemoveOwnerClient(t *testing.T) {
	client1 := &Client{id: 123, nickname: "test_nickname"}
	room := newRoom(1, client1)
	client2 := &Client{id: 456, nickname: "test_nickname2"}
	room.addClient(client2)
	changedOwner, _ := room.removeClient(client1)
	if !changedOwner {
		t.Errorf("TestRemoveOwnerClient expected that room changed owner")
	}
	got := room.owner.client
	expected := client2
	if got != expected {
		t.Errorf("TestRemoveOwnerClient expected: %v, got: %v", expected, got)
	}
}

func TestRemoveRegularClient(t *testing.T) {
	client1 := &Client{id: 123, nickname: "test_nickname"}
	room := newRoom(1, client1)
	client2 := &Client{id: 456, nickname: "test_nickname2"}
	room.addClient(client2)
	changedOwner, roomBecameEmpty := room.removeClient(client2)
	if changedOwner {
		t.Errorf("TestRemoveRegularClient expected that room did not change owner")
	}
	if roomBecameEmpty {
		t.Errorf("TestRemoveRegularClient expected that room did not become empty")
	}
	got := room.owner.client
	expected := client1
	if got != expected {
		t.Errorf("TestRemoveRegularClient expected: %v, got: %v", expected, got)
	}
}
