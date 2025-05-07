package stores

import (
	"context"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"
)

type ConnectionStorage struct {
	conns       sync.Map
	mu          sync.RWMutex
	connections map[string]map[string]struct{}
}

func NewConnectionStorage() *ConnectionStorage {
	return &ConnectionStorage{
		conns:       sync.Map{},
		connections: make(map[string]map[string]struct{}),
	}
}

type ConnectionData struct {
	ClientID          string
	ConnectionID      string
	Ctx               context.Context
	Conn              *websocket.Conn
	SubscribedChannel string
	IsAuthenticated   bool
	CreatedAt         time.Time
}

func (s *ConnectionStorage) Add(ctx context.Context, id, connId string, conn *websocket.Conn, isAuth bool) {
	newConn := ConnectionData{
		ClientID:          id,
		ConnectionID:      connId,
		Ctx:               ctx,
		Conn:              conn,
		IsAuthenticated:   isAuth,
		SubscribedChannel: "",
		CreatedAt:         time.Now(),
	}

	if !s.IsExists(id) {
		data := []ConnectionData{newConn}
		s.conns.Store(id, data)
	} else {
		data, _ := s.Get(id)
		newData := append(data, newConn)
		s.conns.Store(id, newData)
	}

}

func (s *ConnectionStorage) AddChannel(id string, connId, channel string) {
	data, _ := s.Get(id)
	for i, connData := range data {
		if connData.ConnectionID == connId && connData.SubscribedChannel == "" || connData.SubscribedChannel != channel {
			data[i].SubscribedChannel = channel
			break
		}
	}
	s.conns.Store(id, data)
}

func (s *ConnectionStorage) GetByChannel(channel string) ([]ConnectionData, error) {
	allConns := s.GetAll()

	var filteredConns []ConnectionData
	for _, connData := range allConns {
		if connData.SubscribedChannel == channel {
			filteredConns = append(filteredConns, connData)
		}
	}
	return filteredConns, nil
}

func (s *ConnectionStorage) Remove(id string) {
	s.conns.Delete(id)
}

// GetUserForConnection returns the user ID associated with a connection ID
func (s *ConnectionStorage) GetUserForConnection(connectionID string) string {
	for userID, connections := range s.connections {
		if _, exists := connections[connectionID]; exists {
			return userID
		}
	}

	return ""
}

func (s *ConnectionStorage) RemoveByConnID(id string, connId string) {
	data, _ := s.Get(id)
	for i, connData := range data {
		if connData.ConnectionID == connId {
			data = append(data[:i], data[i+1:]...)
			break
		}
	}
	s.conns.Store(id, data)
}

func (s *ConnectionStorage) GetAll() []ConnectionData {
	allConns := make([]ConnectionData, 0)
	s.conns.Range(func(key, value interface{}) bool {
		allConns = append(allConns, value.([]ConnectionData)...)
		return true
	})
	return allConns
}

func (s *ConnectionStorage) GetByConnID(id string, connId string) (*ConnectionData, bool) {
	data, ok := s.Get(id)
	if !ok {
		return nil, false
	}
	for _, connData := range data {
		if connData.ConnectionID == connId {
			return &connData, true
		}
	}
	return nil, false
}

func (s *ConnectionStorage) Get(id string) ([]ConnectionData, bool) {
	v, ok := s.conns.Load(id)
	if !ok {
		return nil, false
	}
	return v.([]ConnectionData), true
}

func (s *ConnectionStorage) IsExists(id string) bool {
	_, ok := s.conns.Load(id)
	return ok
}

func GenerateConnectionID() string {
	return uuid.New().String()
}
