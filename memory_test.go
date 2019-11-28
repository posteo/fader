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
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/posteo/fader"
)

func TestMemory(t *testing.T) {
	t.Run("PutAndGet", func(t *testing.T) {
		fader := fader.NewMemory(50 * time.Millisecond)

		key, time, value := []byte("key"), time.Now(), []byte("value")
		require.NoError(t, fader.Put(key, time, value))

		require.Equal(t, 1, fader.Size())

		ti, v := fader.Get(key)
		assert.Equal(t, time, ti)
		assert.Equal(t, value, v)
	})

	t.Run("SortingByTime", func(t *testing.T) {
		fader := fader.NewMemory(50 * time.Millisecond)

		now := time.Now()

		require.NoError(t, fader.Put([]byte("one"), now.Add(time.Second), []byte("value one")))
		require.NoError(t, fader.Put([]byte("two"), now, []byte("value two")))

		key, time, value := fader.Earliest()
		assert.Equal(t, "two", string(key))
		assert.Equal(t, now, time)
		assert.Equal(t, "value two", string(value))
	})

	t.Run("Select", func(t *testing.T) {
		fader := fader.NewMemory(50 * time.Millisecond)

		now := time.Now()

		require.NoError(t, fader.Put([]byte("one"), now, []byte("value one")))
		require.NoError(t, fader.Put([]byte("two"), now, []byte("value two")))

		times, values := fader.Select([]byte("one"))
		assert.Equal(t, 1, len(times))
		assert.Equal(t, now.Unix(), times[0].Unix())
		assert.Equal(t, 1, len(values))
		assert.Equal(t, "value one", string(values[0]))
	})

	t.Run("Expiry", func(t *testing.T) {
		fader := fader.NewMemory(50 * time.Millisecond)

		require.NoError(t, fader.Put([]byte("one"), time.Now(), []byte("value one")))

		time.Sleep(100 * time.Millisecond)

		_, v := fader.Get([]byte("one"))
		assert.Nil(t, v)
	})

	t.Run("ExpiryOfTwoItem", func(t *testing.T) {
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
	})

	t.Run("ExpiryOfTwoItemsThatHasBeenAddedInReverseOrder", func(t *testing.T) {
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
	})

	t.Run("ConcurrentPut", func(t *testing.T) {
		fader := fader.NewMemory(time.Second)

		wg := sync.WaitGroup{}
		for index := 0; index < 100; index++ {
			wg.Add(1)
			go func(key []byte) {
				for i := 0; i < 50; i++ {
					require.NoError(t, fader.Put(key, time.Now(), []byte("value")))
				}
				wg.Done()
			}([]byte(strconv.Itoa(index)))
		}
		wg.Wait()

		assert.Equal(t, 100*50, fader.Size())
	})

	t.Run("ConcurrentPutAndGet", func(t *testing.T) {
		fader := fader.NewMemory(time.Second)

		wg := sync.WaitGroup{}
		for index := 0; index < 30; index++ {
			key := []byte(strconv.Itoa(index))

			if index%3 == 0 {
				wg.Add(1)
				go func(key []byte) {
					for i := 0; i < 50; i++ {
						require.NoError(t, fader.Put(key, time.Now(), []byte("value")))
					}
					wg.Done()
				}(key)
			}

			wg.Add(1)
			go func(key []byte) {
				for i := 0; i < 50; i++ {
					fader.Get(key)
				}
				wg.Done()
			}(key)
		}
		wg.Wait()

		assert.Equal(t, 10*50, fader.Size())
	})
}

func BenchmarkMemoryPut(b *testing.B) {
	b.Run("Put", func(b *testing.B) {
		fader := fader.NewMemory(50 * time.Millisecond)

		key, time, value := []byte("key"), time.Now(), []byte("value")

		b.ReportAllocs()
		b.ResetTimer()
		for index := 0; index < b.N; index++ {
			if err := fader.Put(key, time, value); err != nil {
				b.Fatalf("put: %v", err)
			}
		}
	})
}
