package service

import (
	"sync"
)

type SSEEvent struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type SSEManager struct {
	clients    map[string]map[string]chan SSEEvent
	boardUsers map[string]map[string]bool
	mu         sync.RWMutex
}

func NewSSEManager() *SSEManager {
	return &SSEManager{
		clients:    make(map[string]map[string]chan SSEEvent),
		boardUsers: make(map[string]map[string]bool),
	}
}

func (m *SSEManager) Register(userID, clientID string, ch chan SSEEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.clients[userID] == nil {
		m.clients[userID] = make(map[string]chan SSEEvent)
	}
	m.clients[userID][clientID] = ch
}

func (m *SSEManager) Unregister(userID, clientID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if clients, ok := m.clients[userID]; ok {
		if _, ok := clients[clientID]; ok {
			delete(clients, clientID)
		}
		if len(clients) == 0 {
			delete(m.clients, userID)
		}
	}
}

func (m *SSEManager) SendToUser(userID string, event SSEEvent) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if clients, ok := m.clients[userID]; ok {
		for _, ch := range clients {
			select {
			case ch <- event:
			default:
			}
		}
	}
}

func (m *SSEManager) BroadcastToBoard(boardID string, event SSEEvent, excludeUserID string) {
	m.mu.RLock()
	users := m.boardUsers[boardID]
	m.mu.RUnlock()

	for userID := range users {
		if userID != excludeUserID {
			m.SendToUser(userID, event)
		}
	}
}

func (m *SSEManager) JoinBoard(userID, boardID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.boardUsers[boardID] == nil {
		m.boardUsers[boardID] = make(map[string]bool)
	}
	m.boardUsers[boardID][userID] = true
}

func (m *SSEManager) LeaveBoard(userID, boardID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if users, ok := m.boardUsers[boardID]; ok {
		delete(users, userID)
		if len(users) == 0 {
			delete(m.boardUsers, boardID)
		}
	}
}

func (m *SSEManager) GetBoardUserCount(boardID string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if users, ok := m.boardUsers[boardID]; ok {
		return len(users)
	}
	return 0
}

func (m *SSEManager) IsUserConnected(userID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	clients, ok := m.clients[userID]
	return ok && len(clients) > 0
}
