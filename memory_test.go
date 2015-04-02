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

package fader_test

import (
	"testing"
	"time"
)

func TestStorage(t *testing.T) {
	e := setUp(t)

	item := &item{KeyField: "test", TimeField: time.Now()}
	e.memoryFaderOne.Store(item)

	e.assertEquals(1, e.memoryFaderOne.Size())
	e.assertEquals(item, e.memoryFaderOne.Detect(item.Key()))
}

func TestSortingByTime(t *testing.T) {
	e := setUp(t)

	now := time.Now()
	duration, _ := time.ParseDuration("1s")

	itemOne := &item{KeyField: "one", TimeField: now.Add(duration)}
	itemTwo := &item{KeyField: "two", TimeField: now}

	e.memoryFaderOne.Store(itemOne)
	e.memoryFaderOne.Store(itemTwo)

	e.assertEquals(itemTwo, e.memoryFaderOne.Earliest())
}

func TestSelect(t *testing.T) {
	e := setUp(t)

	itemOne := &item{KeyField: "one", TimeField: time.Now()}
	itemTwo := &item{KeyField: "two", TimeField: time.Now()}

	e.memoryFaderOne.Store(itemOne)
	e.memoryFaderOne.Store(itemTwo)

	items := e.memoryFaderOne.Select("one")
	e.assertEquals(1, len(items))
	e.assertEquals(itemOne, items[0])
}

func TestClear(t *testing.T) {
	e := setUp(t)

	e.memoryFaderOne.Store(&item{KeyField: "one", TimeField: time.Now()})
	e.assertEquals(1, e.memoryFaderOne.Size())

	e.memoryFaderOne.Clear()
	e.assertEquals(0, e.memoryFaderOne.Size())

	e.memoryFaderOne.Store(&item{KeyField: "one", TimeField: time.Now()})
	e.assertEquals(1, e.memoryFaderOne.Size())
}

func TestExpiry(t *testing.T) {
	e := setUp(t)

	item := &item{KeyField: "one", TimeField: time.Now()}
	e.memoryFaderOne.Store(item)

	time.Sleep(100 * time.Millisecond)

	e.assertEquals(nil, e.memoryFaderOne.Detect(item.Key()))
}

func TestExpiryOfTwoItem(t *testing.T) {
	e := setUp(t)

	now := time.Now()
	duration, _ := time.ParseDuration("20ms")

	itemOne := &item{KeyField: "one", TimeField: now}
	itemTwo := &item{KeyField: "two", TimeField: now.Add(duration)}

	e.memoryFaderOne.Store(itemOne)
	e.memoryFaderOne.Store(itemTwo)
	time.Sleep(10 * time.Millisecond)

	e.assertEquals(2, e.memoryFaderOne.Size())
	time.Sleep(50 * time.Millisecond)
	e.assertEquals(1, e.memoryFaderOne.Size())
	time.Sleep(20 * time.Millisecond)
	e.assertEquals(0, e.memoryFaderOne.Size())
}

func TestExpiryOfTwoItemsThatHasBeenAddedInReverseOrder(t *testing.T) {
	e := setUp(t)

	now := time.Now()
	duration, _ := time.ParseDuration("20ms")

	itemOne := &item{KeyField: "one", TimeField: now}
	itemTwo := &item{KeyField: "two", TimeField: now.Add(duration)}

	e.memoryFaderOne.Store(itemTwo)
	time.Sleep(5 * time.Millisecond)
	e.memoryFaderOne.Store(itemOne)
	time.Sleep(5 * time.Millisecond)

	e.assertEquals(2, e.memoryFaderOne.Size())
	e.assertEquals(itemOne, e.memoryFaderOne.Earliest())
	time.Sleep(50 * time.Millisecond)

	e.assertEquals(1, e.memoryFaderOne.Size())
	e.assertEquals(itemTwo, e.memoryFaderOne.Earliest())
	time.Sleep(20 * time.Millisecond)

	e.assertEquals(0, e.memoryFaderOne.Size())
	e.assertEquals(nil, e.memoryFaderOne.Earliest())
}
