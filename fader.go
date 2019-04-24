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

// Package fader provides an interface to store and fetch items. The implementation
// is responsible for the removal after a expiry period.
//
// Example for a memory fader, that expires items after 2 seconds
//
//    memoryFader := fader.NewMemory(2*time.Second)
//    defer memoryFader.Close()
//
//    memoryFader.Put([]byte("key"), time.Now(), []byte("value"))
//    memoryFader.Size() // => 1
//
//    time.Sleep(3*time.Second)
//    memoryFader.Size() // => 0
//
// The multicast fader can be used to distribute `Put` operations via a multicast
// group. Other instances that listen to the same group, will perform that operation
// on their own, so that each instance end up with the same data.
//
//    multicastFaderOne := fader.NewMulticast(memoryFaderOne, "224.0.0.1:1888", key)
//    defer multicastFaderOne.Close()
//
//    multicastFaderTwo := fader.NewMulticast(memoryFaderTwo, "224.0.0.1:1888", key)
//    defer multicastFaderTwo.Close()
//
//    multicastFaderOne.Put([]byte("key"), time.Now(), []byte("value"))
//    multicastFaderOne.Size() // => 1
//
//    time.Sleep(10*time.Millisecond)
//
//    multicastFaderTwo.Size() // => 1
package fader

import "time"

// Fader defines the fader interface.
type Fader interface {
	Put([]byte, time.Time, []byte) error
	Get([]byte) (time.Time, []byte)
	Earliest() ([]byte, time.Time, []byte)
	Select([]byte) ([]time.Time, [][]byte)
	Size() int
	Close() error
}
