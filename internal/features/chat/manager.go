package chat

import (
	"encoding/json"
	"log"
	"sync"
)

var ChatManager *Manager

type EventList map[int64]ClientList

type Manager struct {
	EventList   EventList
	messageChan chan Message
	sync.RWMutex
}

type Message struct {
	EventId int64
	Payload json.RawMessage
	UserIds []int64
}

func NewManager() *Manager {
	return &Manager{
		EventList:   make(EventList),
		messageChan: make(chan Message, 1000),
	}
}

func (m *Manager) Unregister(client *Client) {
	defer m.Unlock()
	m.Lock()
	if event, ok := m.EventList[client.EventId]; ok {
		if client, ok := event[client.ClientId]; ok {
			delete(m.EventList[client.EventId], client.ClientId)
			close(client.SendMsgChan)
		}
	}
}

func (m *Manager) Register(client *Client) {
	defer m.Unlock()
	m.Lock()
	if client == nil {
		log.Println("nil client")
		return
	}
	if _, ok := m.EventList[client.EventId]; !ok {
		m.EventList[client.EventId] = make(ClientList)
	}
	m.EventList[client.EventId][client.ClientId] = client
}

func (m *Manager) Run() {
	for {
		select {
		case message := <-m.messageChan:
			for _, client := range m.EventList[message.EventId] {
				client := client
				go func() {
					client.SendMsgChan <- message
				}()
			}
		}
	}
}
