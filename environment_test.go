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
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	. "github.com/posteo/fader"
)

type environment struct {
	tb                  testing.TB
	expiresIn           time.Duration
	memoryFaderOne      Fader
	memoryFaderTwo      Fader
	key                 []byte
	multicastFaderIDOne []byte
	multicastFaderIDTwo []byte
	multicastFaderOne   Fader
	multicastFaderTwo   Fader
	tearDown            func()
}

func setUp(tb testing.TB) *environment {
	e := &environment{tb: tb}

	e.expiresIn, _ = time.ParseDuration("50ms")

	e.memoryFaderOne = NewMemory(e.expiresIn)
	require.NoError(tb, e.memoryFaderOne.Open())

	e.memoryFaderTwo = NewMemory(e.expiresIn)
	require.NoError(tb, e.memoryFaderTwo.Open())

	e.key, _ = hex.DecodeString("ab72c77b97cb5fe9a382d9fe81ffdbed")

	e.multicastFaderIDOne = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	e.multicastFaderIDTwo = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 1}

	e.multicastFaderOne = NewMulticast(e.memoryFaderOne, "224.0.0.1:2000", e.key, e.multicastFaderIDOne, nil)
	require.NoError(tb, e.multicastFaderOne.Open())

	e.multicastFaderTwo = NewMulticast(e.memoryFaderTwo, "224.0.0.1:2000", e.key, e.multicastFaderIDTwo, nil)
	require.NoError(tb, e.multicastFaderTwo.Open())

	e.tearDown = func() {
		require.NoError(tb, e.memoryFaderOne.Close())
		require.NoError(tb, e.memoryFaderTwo.Close())
		require.NoError(tb, e.multicastFaderOne.Close())
		require.NoError(tb, e.multicastFaderTwo.Close())
	}

	return e
}
