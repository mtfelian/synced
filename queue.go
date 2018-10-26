package synced

import (
	"errors"
	"fmt"
	"sync"
)

// mode constants
const (
	modeNormal = iota
	modeDrop   // drops elements if push overflows
)

// errors
var (
	ErrQueueIsEmpty    = errors.New("queue is empty")
	ErrQueueOverflowed = errors.New("queue overflowed")
	ErrFailedToDrop    = func(err error) error { return fmt.Errorf("failed to drop element: %v", err) }
)

// Queue that is thread-safe
type Queue struct {
	queue  []interface{}
	maxLen int
	mode   int
	sync.Mutex
}

// NewQueue returns a new synced queue
func NewQueue() Queue { return Queue{queue: []interface{}{}, mode: modeNormal} }

// NewLimitedQueue returns a new synced limited queue
func NewLimitedQueue(max int) Queue {
	return Queue{queue: make([]interface{}, 0, max), maxLen: max, mode: modeNormal}
}

// NewDroppingQueue returns a new synced dropping queue
func NewDroppingQueue(max int) Queue {
	return Queue{queue: make([]interface{}, 0, max), maxLen: max, mode: modeDrop}
}

// Push pushed an object to a queue
func (q *Queue) Push(object interface{}) error {
	q.Lock()
	defer q.Unlock()

	if q.maxLen == 0 || q.len() < q.maxLen {
		q.queue = append(q.queue, object)
		return nil
	}

	// q.maxLen > 0 && q.len() == q.maxLen

	switch q.mode {
	case modeNormal:
		return ErrQueueOverflowed
	case modeDrop:
		if _, err := q.pop(); err != nil {
			return ErrFailedToDrop(err)
		}
		q.queue = append(q.queue, object)
		return nil
	}
	return nil
}

func (q *Queue) len() int { return len(q.queue) }

// Len returns a queue current length
func (q *Queue) Len() int {
	q.Lock()
	defer q.Unlock()
	return q.len()
}

func (q *Queue) pop() (interface{}, error) {
	if q.len() == 0 {
		return nil, ErrQueueIsEmpty
	}
	popped := q.queue[0]
	q.queue = q.queue[1:]
	return popped, nil
}

// Pop returns an object from a queue
func (q *Queue) Pop() (interface{}, error) {
	q.Lock()
	defer q.Unlock()
	return q.pop()
}

// Get element but don't pop it
func (q *Queue) Get() (interface{}, error) {
	q.Lock()
	defer q.Unlock()
	if q.len() == 0 {
		return nil, ErrQueueIsEmpty
	}
	return q.queue[0], nil
}
