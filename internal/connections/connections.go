package connections

import (
	"sync"

	"github.com/coder/websocket"
	"github.com/google/uuid"
)

type ConnectionStorage struct {
	conns sync.Map
}

func NewConnectionStorage() *ConnectionStorage {
	return &ConnectionStorage{
		conns: sync.Map{},
	}
}

type ConnectionData struct {
	ConnectionID      string
	Conn              *websocket.Conn
	SubscribedChannel string
	IsAuthenticated   bool
}

func (s *ConnectionStorage) Add(id, connId string, conn *websocket.Conn, isAuth bool) {
	if !s.IsExists(id) {
		data := []ConnectionData{
			{
				ConnectionID:    connId,
				Conn:            conn,
				IsAuthenticated: isAuth,
			},
		}
		s.conns.Store(id, data)
	} else {
		data, _ := s.Get(id)
		newData := append(data, ConnectionData{
			ConnectionID:    connId,
			Conn:            conn,
			IsAuthenticated: isAuth,
		})
		s.conns.Store(id, newData)
	}
}

func (s *ConnectionStorage) AddChannel(id string, channel string) {
	data, _ := s.Get(id)
	for i, connData := range data {
		if connData.SubscribedChannel == "" {
			data[i].SubscribedChannel = channel
			break
		}
	}
	s.conns.Store(id, data)
}

func (s *ConnectionStorage) Remove(id string) {
	s.conns.Delete(id)
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

func (s *ConnectionStorage) GetAll() map[string][]ConnectionData {
	allConns := make(map[string][]ConnectionData)
	s.conns.Range(func(key, value interface{}) bool {
		allConns[key.(string)] = value.([]ConnectionData)
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
