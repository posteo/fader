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

	. "github.com/posteo/fader"
)

func TestMulticastTransferBetweenTwoFaders(t *testing.T) {
	e := setUp(t)
	defer e.tearDown()

	now := time.Now()
	e.assertNoError(
		e.multicastFaderOne.Store(&item{KeyField: "test", TimeField: now}))
	time.Sleep(10 * time.Millisecond)

	e.assertEquals(1, e.multicastFaderOne.Size())
	e.assertEquals(1, e.multicastFaderTwo.Size())

	itemOne := e.multicastFaderOne.Earliest()
	e.assertEquals("test", itemOne.Key())
	e.assertEquals(now, itemOne.Time())

	itemTwo := e.multicastFaderTwo.Earliest()
	e.assertEquals("test", itemTwo.Key())
	e.assertEquals(now, itemTwo.Time())
}

func TestMulticastTransferOfMultipleStores(t *testing.T) {
	e := setUp(t)
	defer e.tearDown()

	now := time.Now()
	e.multicastFaderOne.Store(&item{KeyField: "one", TimeField: now})
	e.multicastFaderOne.Store(&item{KeyField: "two", TimeField: now})
	time.Sleep(10 * time.Millisecond)

	e.assertEquals(2, e.multicastFaderOne.Size())
	e.assertEquals(2, e.multicastFaderTwo.Size())

	itemOne := e.multicastFaderOne.Earliest()
	e.assertEquals("one", itemOne.Key())
	e.assertEquals(now, itemOne.Time())

	itemTwo := e.multicastFaderTwo.Earliest()
	e.assertEquals("one", itemTwo.Key())
	e.assertEquals(now, itemTwo.Time())
}

func TestMulticastTransferOfStoreAndExpire(t *testing.T) {
	e := setUp(t)
	defer e.tearDown()

	now := time.Now()
	e.multicastFaderOne.Store(&item{KeyField: "test", TimeField: now})
	time.Sleep(10 * time.Millisecond)

	e.assertEquals(1, e.multicastFaderOne.Size())
	e.assertEquals(1, e.multicastFaderTwo.Size())
	time.Sleep(100 * time.Millisecond)

	e.assertEquals(0, e.multicastFaderOne.Size())
	e.assertEquals(0, e.multicastFaderTwo.Size())
}

func TestIfTransmissionFailsOnAReplyAttack(t *testing.T) {
	e := setUp(t)
	defer e.tearDown()

	i := &item{KeyField: "test", TimeField: time.Now()}
	e.assertNoError(
		e.multicastFaderOne.Store(i))
	time.Sleep(10 * time.Millisecond)

	e.assertEquals(1, e.multicastFaderOne.Size())
	e.assertEquals(1, e.multicastFaderTwo.Size())

	// forge a reply attack
	memoryFader := NewMemory(e.expiresIn)
	e.assertNoError(memoryFader.Open())
	defer memoryFader.Close()

	multicastFader := NewMulticast(memoryFader, "224.0.0.1:1888", e.key, e.multicastFaderIDOne)
	e.assertNoError(multicastFader.Open())
	defer multicastFader.Close()

	e.assertNoError(
		multicastFader.Store(i))
	time.Sleep(10 * time.Millisecond)

	e.assertEquals(1, e.multicastFaderOne.Size())
	e.assertEquals(1, e.multicastFaderTwo.Size())
}
