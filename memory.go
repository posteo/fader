// Copyright 2014 The fader authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fader

import (
	"bytes"
	"container/heap"
	"time"
)

type Memory struct {
	expiresIn  time.Duration
	items      *itemHeap
	itemStored chan bool
	closed     chan bool
}

var (
	veryLongDuration time.Duration
)

func init() {
	veryLongDuration, _ = time.ParseDuration("24h")
}

// NewMemory creates a Fader instance that stores all data in the Memory. The expiresIn
// parameter defines after which period a stored item will be removed.
func NewMemory(expiresIn time.Duration) *Memory {
	m := &Memory{
		expiresIn:  expiresIn,
		items:      &itemHeap{},
		itemStored: make(chan bool),
		closed:     make(chan bool),
	}
	heap.Init(m.items)
	go m.expiryLoop()
	return m
}

func (m *Memory) Put(key []byte, t time.Time, value []byte) error {
	heap.Push(m.items, &item{
		key:   key,
		time:  t,
		value: value,
	})
	m.itemStored <- true
	return nil
}

func (m *Memory) Get(key []byte) (time.Time, []byte) {
	for _, item := range *m.items {
		if bytes.Equal(item.key, key) {
			return item.time, item.value
		}
	}
	return time.Time{}, nil
}

func (m *Memory) Earliest() ([]byte, time.Time, []byte) {
	if m.Size() > 0 {
		item := (*m.items)[0]
		return item.key, item.time, item.value
	}
	return nil, time.Time{}, nil
}

func (m *Memory) Select(key []byte) [][]byte {
	var values [][]byte
	for _, item := range *m.items {
		if bytes.Equal(item.key, key) {
			values = append(values, item.value)
		}
	}
	return values
}

func (m *Memory) Size() int {
	return m.items.Len()
}

func (m *Memory) Close() error {
	m.closed <- true
	return nil
}

func (m *Memory) removeEarliest() *item {
	if m.Size() > 0 {
		return heap.Pop(m.items).(*item)
	}
	return nil
}

// This function should run in it's own goroutine. It runs in an infinite loop
// until something is send through the m.closed channel.
// If a new item is stored, the earliest item is fetched from the heap and the
// duration to it's expiry is calculated. Even if the earliest item hasn't
// changed, this calculation is needed, because the duration to it's expiry
// has changed.
// If the earliest item finally expires, it's removed from the heap and the
// duration to the next item expiry is calculated.
// If no items left, the function returns to it's initial state where it waits
// for an item to be stored.
func (m *Memory) expiryLoop() {
	durationTillNextExpiry := veryLongDuration

expiryLoop:
	for {
		select {
		case <-m.itemStored:
			durationTillNextExpiry = m.findNextDurationTillNextExpiry()
		case <-time.After(durationTillNextExpiry):
			m.removeEarliest()
			durationTillNextExpiry = m.findNextDurationTillNextExpiry()
		case <-m.closed:
			break expiryLoop
		}
	}
}

func (m *Memory) findNextDurationTillNextExpiry() time.Duration {
	result := veryLongDuration

	for k, t, _ := m.Earliest(); k != nil; k, t, _ = m.Earliest() {
		duration := t.Sub(time.Now().Add(-m.expiresIn))
		if duration > 0 {
			result = duration
			break
		}

		m.removeEarliest()
	}

	return result
}
