package telegram

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type ChatState string

type StateManager interface {
	GetState(ctx context.Context, id int) (ChatState, error)
	SetState(ctx context.Context, id int, state ChatState) error
}

var NoopHandler = &StateHandler{
	Handle: func(ctx *Context) (*Context, error) {
		return ctx, nil
	},
}

type StateHandler struct {
	Title   string
	OnEnter func(ctx *Context) error
	Handle  func(ctx *Context) (*Context, error)
	OnExit  func(ctx *Context) error
}

type Machine struct {
	StateManager
	transitions map[ChatState][]ChatState
	states      map[ChatState]*StateHandler
}

func (m *Machine) GetHandlerForState(state ChatState) *StateHandler {
	return m.states[state]
}

func (m *Machine) TransitFrom(id int, from, to ChatState) (*StateHandler, error) {
	if !m.isAllowedToTransit(from, to) {
		return nil, ErrTransitionNowAllowed
	}

	if err := m.SetState(context.Background(), id, to); err != nil {
		return nil, fmt.Errorf("failed to transit to new state %q: %w", to, err)
	}

	return m.states[to], nil
}

func (m *Machine) isAllowedToTransit(from, to ChatState) bool {
	transitions, ok := m.transitions[from]
	if !ok {
		return false
	}
	for _, t := range transitions {
		if t == to {
			return true
		}
	}

	return false
}

func NewMachine(manager StateManager) *Machine {
	return &Machine{
		StateManager: manager,
	}
}

func (m *Machine) AddTransitions(from ChatState, to ...ChatState) {
	if m.transitions == nil {
		m.transitions = make(map[ChatState][]ChatState)
	}
	if m.states == nil {
		m.states = make(map[ChatState]*StateHandler)
	}

	if len(m.transitions[from]) == 0 {
		m.transitions[from] = to
	} else {
		m.transitions[from] = append(m.transitions[from], to...)
	}
}

func (m *Machine) AddStateHandler(state ChatState, handler *StateHandler) error {
	if _, ok := m.transitions[state]; !ok {
		return ErrStateMissing
	}

	m.states[state] = handler

	return nil
}

type inmemState struct {
	rw   *sync.RWMutex
	data map[int]ChatState
}

func (m *inmemState) GetState(ctx context.Context, id int) (ChatState, error) {
	m.rw.RLock()
	userState := m.get(id)
	m.rw.RUnlock()

	if userState == "" {
		return "", ErrStateMissing
	}

	return userState, nil
}

func (m *inmemState) SetState(ctx context.Context, id int, state ChatState) error {
	m.rw.Lock()
	m.data[id] = state
	m.rw.Unlock()

	return nil
}

func (m *inmemState) set(id int, state ChatState) {
	m.data[id] = state
}

func (m *inmemState) get(id int) ChatState {
	if uState, ok := m.data[id]; ok {
		return uState
	}

	return ""
}

func newInmemState() *inmemState {
	return &inmemState{
		rw:   new(sync.RWMutex),
		data: make(map[int]ChatState),
	}
}

var (
	ErrTransitionNowAllowed = errors.New("transition is not allowed")
	ErrStateMissing         = errors.New("no state for user exist")
)
