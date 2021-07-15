package telegram

import (
	"context"
	"errors"
	"sync"
)

type inmemDataStorage struct {
	rw   *sync.RWMutex
	data map[string]*UserData
}

type UserData struct {
	InviteNotificationSent bool
	Credentials            UserCredentials
}

func (m *inmemDataStorage) GetData(ctx context.Context, id string) (*UserData, error) {
	m.rw.RLock()
	userState := m.get(id)
	m.rw.RUnlock()

	if userState == nil {
		return nil, ErrDataMissing
	}

	return userState, nil
}

func (m *inmemDataStorage) SetData(ctx context.Context, id string, state *UserData) error {
	m.rw.Lock()
	m.data[id] = state
	m.rw.Unlock()

	return nil
}

func (m *inmemDataStorage) set(id string, state *UserData) {
	m.data[id] = state
}

func (m *inmemDataStorage) get(id string) *UserData {
	if uState, ok := m.data[id]; ok {
		return uState
	}

	return nil
}

func newInmemStorage() *inmemDataStorage {
	return &inmemDataStorage{
		rw:   new(sync.RWMutex),
		data: make(map[string]*UserData),
	}
}

func InitialUserData() *UserData {
	return &UserData{
		Credentials: UserCredentials{},
	}
}

var (
	ErrDataMissing = errors.New("no state for sender exist")
)
