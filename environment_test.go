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

	"github.com/posteo/fader"
)

type environment struct {
	expiresIn           time.Duration
	memoryFaderOne      *fader.Memory
	memoryFaderTwo      *fader.Memory
	key                 []byte
	multicastFaderIDOne []byte
	multicastFaderIDTwo []byte
	multicastFaderOne   *fader.Multicast
	multicastFaderTwo   *fader.Multicast
	tearDown            func()
}

func setUp(tb testing.TB) *environment {
	expiresIn, _ := time.ParseDuration("50ms")

	memoryFaderOne := fader.NewMemory(expiresIn)
	memoryFaderTwo := fader.NewMemory(expiresIn)

	key, err := hex.DecodeString("ab72c77b97cb5fe9a382d9fe81ffdbed")
	require.NoError(tb, err)

	multicastFaderIDOne := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	multicastFaderIDTwo := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 1}

	multicastFaderOne, err := fader.NewMulticast(memoryFaderOne, "224.0.0.1:2000", key, multicastFaderIDOne, nil)
	require.NoError(tb, err)

	multicastFaderTwo, err := fader.NewMulticast(memoryFaderTwo, "224.0.0.1:2000", key, multicastFaderIDTwo, nil)
	require.NoError(tb, err)

	return &environment{
		expiresIn:           expiresIn,
		memoryFaderOne:      memoryFaderOne,
		memoryFaderTwo:      memoryFaderTwo,
		key:                 key,
		multicastFaderIDOne: multicastFaderIDOne,
		multicastFaderIDTwo: multicastFaderIDTwo,
		multicastFaderOne:   multicastFaderOne,
		multicastFaderTwo:   multicastFaderTwo,
		tearDown: func() {
			require.NoError(tb, memoryFaderOne.Close())
			require.NoError(tb, memoryFaderTwo.Close())
			require.NoError(tb, multicastFaderOne.Close())
			require.NoError(tb, multicastFaderTwo.Close())
		},
	}
}
