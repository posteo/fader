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
	"log"
	"sync"
	"time"
)

// Memory implements a memory fader.
type Memory struct {
	expiresIn  time.Duration
	items      itemHeap
	itemsMutex sync.RWMutex
	itemStored chan struct{}
	closed     chan struct{}
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
		items:      itemHeap{},
		itemStored: make(chan struct{}),
		closed:     make(chan struct{}),
	}

	m.itemsMutex.Lock()
	heap.Init(&m.items)
	m.itemsMutex.Unlock()

	go m.expiryLoop()

	return m
}

// Put places an item with the provided key, time and value in the fader.
func (m *Memory) Put(key []byte, t time.Time, value []byte) error {
	m.itemsMutex.Lock()
	heap.Push(&m.items, &item{
		key:   key,
		time:  t,
		value: value,
	})
	m.itemsMutex.Unlock()

	m.itemStored <- struct{}{}

	return nil
}

// Get returns time and value for the provided key. If no such key exists, a value
// of nil is returned.
func (m *Memory) Get(key []byte) (time.Time, []byte) {
	m.itemsMutex.RLock()
	for _, item := range m.items {
		if bytes.Equal(item.key, key) {
			m.itemsMutex.RUnlock()
			return item.time, item.value
		}
	}
	m.itemsMutex.RUnlock()
	return time.Time{}, nil
}

// Earliest returns key, time and value of the earliest item in the fader.
func (m *Memory) Earliest() ([]byte, time.Time, []byte) {
	m.itemsMutex.RLock()
	if m.items.Len() > 0 {
		item := m.items[0]
		m.itemsMutex.RUnlock()
		return item.key, item.time, item.value
	}
	m.itemsMutex.RUnlock()
	return nil, time.Time{}, nil
}

// Select returns all times and values with the provided key.
func (m *Memory) Select(key []byte) ([]time.Time, [][]byte) {
	m.itemsMutex.RLock()
	times := []time.Time{}
	values := [][]byte{}
	for _, item := range m.items {
		if bytes.Equal(item.key, key) {
			times = append(times, item.time)
			values = append(values, item.value)
		}
	}
	m.itemsMutex.RUnlock()
	return times, values
}

// Size returns the number of items in the fader.
func (m *Memory) Size() int {
	m.itemsMutex.RLock()
	l := m.items.Len()
	m.itemsMutex.RUnlock()
	return l
}

// Close tears down the fader.
func (m *Memory) Close() error {
	m.closed <- struct{}{}
	return nil
}

func (m *Memory) removeEarliest() *item {
	m.itemsMutex.Lock()
	if m.items.Len() > 0 {
		i := heap.Pop(&m.items).(*item)
		m.itemsMutex.Unlock()
		return i
	}
	m.itemsMutex.Unlock()
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

	expiryDelay := time.NewTimer(durationTillNextExpiry)
	defer expiryDelay.Stop()

	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic: %v", r)
		}
	}()

expiryLoop:
	for {
		expiryDelay.Reset(durationTillNextExpiry)
		select {
		case <-m.itemStored:
			durationTillNextExpiry = m.findNextDurationTillNextExpiry()
		case <-expiryDelay.C:
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
