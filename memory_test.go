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

func TestMemoryPutAndGet(t *testing.T) {
	fader := fader.NewMemory(50 * time.Millisecond)

	key, time, value := []byte("key"), time.Now(), []byte("value")
	require.NoError(t, fader.Put(key, time, value))

	require.Equal(t, 1, fader.Size())

	ti, v := fader.Get(key)
	assert.Equal(t, time, ti)
	assert.Equal(t, value, v)
}

func TestMemorySortingByTime(t *testing.T) {
	fader := fader.NewMemory(50 * time.Millisecond)

	now := time.Now()

	require.NoError(t, fader.Put([]byte("one"), now.Add(time.Second), []byte("value one")))
	require.NoError(t, fader.Put([]byte("two"), now, []byte("value two")))

	key, time, value := fader.Earliest()
	assert.Equal(t, "two", string(key))
	assert.Equal(t, now, time)
	assert.Equal(t, "value two", string(value))
}

func TestMemorySelect(t *testing.T) {
	fader := fader.NewMemory(50 * time.Millisecond)

	now := time.Now()

	require.NoError(t, fader.Put([]byte("one"), now, []byte("value one")))
	require.NoError(t, fader.Put([]byte("two"), now, []byte("value two")))

	values := fader.Select([]byte("one"))
	assert.Equal(t, 1, len(values))
	assert.Equal(t, "value one", string(values[0]))
}

func TestMemoryExpiry(t *testing.T) {
	fader := fader.NewMemory(50 * time.Millisecond)

	require.NoError(t, fader.Put([]byte("one"), time.Now(), []byte("value one")))

	time.Sleep(100 * time.Millisecond)

	_, v := fader.Get([]byte("one"))
	assert.Nil(t, v)
}

func TestMemoryExpiryOfTwoItem(t *testing.T) {
	fader := fader.NewMemory(50 * time.Millisecond)

	now := time.Now()

	require.NoError(t, fader.Put([]byte("one"), now, []byte("value one")))
	require.NoError(t, fader.Put([]byte("two"), now.Add(20*time.Millisecond), []byte("value two")))

	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 2, fader.Size())
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 1, fader.Size())
	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, 0, fader.Size())
}

func TestMemoryExpiryOfTwoItemsThatHasBeenAddedInReverseOrder(t *testing.T) {
	fader := fader.NewMemory(50 * time.Millisecond)

	now := time.Now()

	require.NoError(t, fader.Put([]byte("two"), now.Add(20*time.Millisecond), []byte("value two")))
	time.Sleep(5 * time.Millisecond)
	require.NoError(t, fader.Put([]byte("one"), now, []byte("value one")))
	time.Sleep(5 * time.Millisecond)

	assert.Equal(t, 2, fader.Size())
	key, _, _ := fader.Earliest()
	assert.Equal(t, "one", string(key))
	time.Sleep(50 * time.Millisecond)

	assert.Equal(t, 1, fader.Size())
	key, _, _ = fader.Earliest()
	assert.Equal(t, "two", string(key))
	time.Sleep(20 * time.Millisecond)

	assert.Equal(t, 0, fader.Size())
	key, _, _ = fader.Earliest()
	assert.Nil(t, key)
}
