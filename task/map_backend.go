package task

import (
	"sync"

	"github.com/baetyl/baetyl-go/v2/errors"
)

var (
	ErrResultNotFound = errors.New("failed to find result")
)

type mapBackend struct {
	mapLock sync.RWMutex
	cache   map[string]*ResultMessage
}

func NewMapBackend() TaskBackend {
	return &mapBackend{
		cache: map[string]*ResultMessage{},
	}
}

func (m *mapBackend) GetResult(taskId string) (*ResultMessage, error) {
	m.mapLock.Lock()
	defer m.mapLock.Unlock()
	result, ok := m.cache[taskId]
	if !ok {
		return nil, ErrResultNotFound
	}
	delete(m.cache, taskId)
	return result, nil
}

func (m *mapBackend) SetResult(taskID string, result *ResultMessage) error {
	m.mapLock.Lock()
	defer m.mapLock.Unlock()
	m.cache[taskID] = result
	return nil
}
