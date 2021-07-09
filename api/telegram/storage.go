package telegram

import (
	"context"
	"errors"
	"sync"
)

type inmemDataStorage struct {
	rw   *sync.RWMutex
	data map[int]*UserData
}

type UserData struct {
	QRRetries   int
	Credentials UserCredentials
}

func (m *inmemDataStorage) GetData(ctx context.Context, id int) (*UserData, error) {
	m.rw.RLock()
	userState := m.get(id)
	m.rw.RUnlock()

	if userState == nil {
		return nil, ErrDataMissing
	}

	return userState, nil
}

func (m *inmemDataStorage) SetData(ctx context.Context, id int, state *UserData) error {
	m.rw.Lock()
	m.data[id] = state
	m.rw.Unlock()

	return nil
}

func (m *inmemDataStorage) set(id int, state *UserData) {
	m.data[id] = state
}

func (m *inmemDataStorage) get(id int) *UserData {
	if uState, ok := m.data[id]; ok {
		return uState
	}

	return nil
}

func newInmemStorage() *inmemDataStorage {
	return &inmemDataStorage{
		rw:   new(sync.RWMutex),
		data: make(map[int]*UserData),
	}
}

func InitialUserData() *UserData {
	return &UserData{
		QRRetries:   maxQRGenerations,
		Credentials: UserCredentials{},
	}
}

var (
	ErrDataMissing = errors.New("no state for user exist")
)
