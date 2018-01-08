package main

import "testing"

type TestEvent struct {
	Field1 string `json:"field1"`
	Field2 int    `json:"field2"`
}

func TestEventToJSON(t *testing.T) {
	event := &TestEvent{"test", 123}
	got, err := eventToJSON(event)
	if err != nil {
		t.Fatalf("TestEventToJSON got error: %s", err)
	}
	expected := "{\"name\":\"TestEvent\",\"data\":{\"field1\":\"test\",\"field2\":123}}"
	if string(got) != expected {
		t.Errorf("TestEventToJSON expected: %s, got: %s", expected, got)
	}
}

func TestGetNameOfStructPointer(t *testing.T) {
	event := &TestEvent{}
	got := getNameOfStruct(event)
	expected := "TestEvent"
	if got != expected {
		t.Errorf("TestGetNameOfStructPointer expected: %s, got: %s", expected, got)
	}
}

func TestGetNameOfStructValue(t *testing.T) {
	event := TestEvent{}
	got := getNameOfStruct(event)
	expected := "TestEvent"
	if got != expected {
		t.Errorf("TestGetNameOfStructPointer expected: %s, got: %s", expected, got)
	}
}
