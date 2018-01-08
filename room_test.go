package main

import (
	"testing"
)

func TestNewRoom(t *testing.T) {
	client := &Client{nickname: "test_nickname"}
	room := newRoom(client)
	got := room.owner
	expected := client
	if got != expected {
		t.Errorf("TestNewRoom expected: %v, got: %v", expected, got)
	}
}

func TestName(t *testing.T) {
	client := &Client{nickname: "test_nickname"}
	room := newRoom(client)
	got := room.Name()
	expected := "test_nickname"
	if got != expected {
		t.Errorf("TestName expected: %v, got: %v", expected, got)
	}
}

func TestId(t *testing.T) {
	client := &Client{id: 123}
	room := newRoom(client)
	got := room.Id()
	expected := uint64(123)
	if got != expected {
		t.Errorf("TestId expected: %v, got: %v", expected, got)
	}
}

func TestToRoomInList(t *testing.T) {
	client := &Client{id: 123, nickname: "test_nickname"}
	room := newRoom(client)
	roomInList := room.toRoomInList()
	got := getNameOfStruct(roomInList)
	expected := "RoomInList"
	if got != expected {
		t.Errorf("TestToRoomInList expected: %v, got: %v", expected, got)
	}
}
