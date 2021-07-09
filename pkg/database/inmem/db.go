package inmem

import (
	"errors"
	"sync"
)

type Table struct {
	rw   *sync.RWMutex
	rows map[string]interface{}
}

type DB struct {
	rw     *sync.RWMutex
	tables map[string]*Table
}

func (db *DB) Table(name string) *Table {
	db.rw.Lock()
	defer db.rw.Unlock()
	if db.tables == nil {
		db.tables = make(map[string]*Table)
	}

	t, ok := db.tables[name]
	if !ok {
		t = newTable()
		db.tables[name] = t
	}

	return t
}

func newTable() *Table {
	return &Table{rw: new(sync.RWMutex), rows: make(map[string]interface{})}
}

func (t *Table) List() []interface{} {
	t.rw.RLock()
	defer t.rw.RUnlock()
	res := make([]interface{}, 0, len(t.rows))
	for _, r := range t.rows {
		res = append(res, r)
	}

	return res
}

func (t *Table) Delete(key string) bool {
	t.rw.Lock()
	defer t.rw.Unlock()
	delete(t.rows, key)

	return true
}

func (t *Table) Get(key string) (interface{}, error) {
	t.rw.RLock()
	defer t.rw.RUnlock()
	row, ok := t.rows[key]
	if !ok {
		return nil, ErrRowNotFound
	}

	return row, nil
}

func (t *Table) Write(key string, row interface{}) error {
	t.rw.Lock()
	t.rows[key] = row
	t.rw.Unlock()

	return nil
}

func New() *DB {
	return &DB{
		rw: new(sync.RWMutex),
	}
}

var (
	ErrRowNotFound = errors.New("row not found")
)
