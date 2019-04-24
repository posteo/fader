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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/posteo/fader"
)

var (
	multicastFaderIDOne = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	multicastFaderIDTwo = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 1}
	multicastKey        = []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
)

func setUpFader(tb testing.TB, id []byte) *fader.Multicast {
	fader, err := fader.NewMulticast(fader.NewMemory(50*time.Millisecond), "224.0.0.1:2000", multicastKey, id, nil)
	require.NoError(tb, err)
	return fader
}

func TestMulticastTransferBetweenTwoFaders(t *testing.T) {
	faderOne := setUpFader(t, multicastFaderIDOne)
	faderTwo := setUpFader(t, multicastFaderIDTwo)

	now := time.Now()
	require.NoError(t, faderOne.Put([]byte("test"), now, []byte("value")))
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 1, faderOne.Size())
	assert.Equal(t, 1, faderTwo.Size())

	key, time, value := faderOne.Earliest()
	assert.Equal(t, "test", string(key))
	assert.Equal(t, now.Unix(), time.Unix())
	assert.Equal(t, "value", string(value))

	key, time, value = faderTwo.Earliest()
	assert.Equal(t, "test", string(key))
	assert.Equal(t, now.Unix(), time.Unix())
	assert.Equal(t, "value", string(value))
}

func TestMulticastTransferOfMultipleStores(t *testing.T) {
	faderOne := setUpFader(t, multicastFaderIDOne)
	faderTwo := setUpFader(t, multicastFaderIDTwo)

	now := time.Now()
	require.NoError(t, faderOne.Put([]byte("one"), now, []byte("value one")))
	require.NoError(t, faderOne.Put([]byte("two"), now, []byte("value two")))
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 2, faderOne.Size())
	assert.Equal(t, 2, faderTwo.Size())

	key, time, value := faderOne.Earliest()
	assert.Equal(t, "one", string(key))
	assert.Equal(t, now.Unix(), time.Unix())
	assert.Equal(t, "value one", string(value))

	key, time, value = faderTwo.Earliest()
	assert.Equal(t, "one", string(key))
	assert.Equal(t, now.Unix(), time.Unix())
	assert.Equal(t, "value one", string(value))
}

func TestMulticastTransferOfStoreAndExpire(t *testing.T) {
	faderOne := setUpFader(t, multicastFaderIDOne)
	faderTwo := setUpFader(t, multicastFaderIDTwo)

	now := time.Now()
	require.NoError(t, faderOne.Put([]byte("test"), now, []byte("value")))
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 1, faderOne.Size())
	assert.Equal(t, 1, faderTwo.Size())
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, 0, faderOne.Size())
	assert.Equal(t, 0, faderTwo.Size())
}

func TestMulticastIfTransmissionFailsOnAReplyAttack(t *testing.T) {
	faderOne := setUpFader(t, multicastFaderIDOne)
	faderTwo := setUpFader(t, multicastFaderIDTwo)

	now := time.Now()
	require.NoError(t, faderOne.Put([]byte("test"), now, []byte("value")))
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 1, faderOne.Size())
	assert.Equal(t, 1, faderTwo.Size())

	// forge a reply attack
	memoryFader := fader.NewMemory(50 * time.Millisecond)
	defer memoryFader.Close()

	multicastFader, err := fader.NewMulticast(memoryFader, "224.0.0.1:2000", multicastKey, multicastFaderIDOne, nil)
	require.NoError(t, err)
	defer multicastFader.Close()

	require.NoError(t, multicastFader.Put([]byte("test"), now, []byte("value")))
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 1, faderOne.Size())
	assert.Equal(t, 1, faderTwo.Size())
}
