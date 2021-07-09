package telegram

import (
	"math/rand"
	"sync"
	"time"
)

const (
	maxResponses     = 5
	noAnswerDuration = time.Second * 30
)

type chatterLimit struct {
	responsesLeft int
	noAnswerTill  time.Time
}

func (l *chatterLimit) canAnswer() bool {
	if l.noAnswerTill.IsZero() {
		return true
	}

	can := time.Now().UnixNano() >= l.noAnswerTill.UnixNano()
	if can {
		l.noAnswerTill = time.Time{}
		l.responsesLeft = maxResponses
		return true
	}

	return false
}

func (l *chatterLimit) update() {
	l.responsesLeft--
	if l.responsesLeft <= 0 {
		l.responsesLeft = 0
		l.noAnswerTill = time.Now().Add(noAnswerDuration)
	}
}

type chatter struct {
	rw               *sync.RWMutex
	defaultResponses []string
	lastResponses    []string
	limits           map[int]*chatterLimit
}

func (c *chatter) answerFor(id int) string {
	c.rw.Lock()
	defer c.rw.Unlock()
	limit, ok := c.limits[id]
	if !ok {
		limit = &chatterLimit{responsesLeft: maxResponses}
		c.limits[id] = limit
	}

	if !limit.canAnswer() {
		return ""
	}

	defer limit.update()
	if limit.responsesLeft <= len(c.lastResponses) {
		return c.lastResponses[limit.responsesLeft-1]
	}

	return c.defaultResponses[rand.Intn(len(c.defaultResponses)-1)]
}

func newChatter(defaults, last []string) *chatter {
	return &chatter{
		rw:               new(sync.RWMutex),
		defaultResponses: defaults,
		lastResponses:    last,
		limits:           make(map[int]*chatterLimit),
	}
}
