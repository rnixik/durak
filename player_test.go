package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type MockedClient struct {
	mock.Mock
	Client
	nickname string
}

func (m *MockedClient) sendEvent(event interface{}) {
	m.Called(event)
}

func (m *MockedClient) sendMessage(message []byte) {
	m.Called(message)
}

type TestPlayerEvent struct {
}

func TestSendEvent(t *testing.T) {
	mockedClient := &MockedClient{}
	player := newPlayer(mockedClient, false)
	event := &TestPlayerEvent{}

	mockedClient.On("sendEvent", event).Return()

	player.sendEvent(event)

	mockedClient.AssertCalled(t, "sendEvent", event)
}

func TestSendMessage(t *testing.T) {
	mockedClient := &MockedClient{}
	player := newPlayer(mockedClient, false)
	message := []byte("test")

	mockedClient.On("sendMessage", message).Return()

	player.sendMessage(message)

	mockedClient.AssertCalled(t, "sendMessage", message)
}

func TestNewPlayer(t *testing.T) {
	client := &Client{nickname: "test_nickname"}
	player := newPlayer(client, true)

	assert := assert.New(t)
	assert.Equal("test_nickname", player.Name)
	assert.Equal(true, player.IsActive)
	assert.Equal(client, player.client)
}
