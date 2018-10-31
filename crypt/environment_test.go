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

package crypt_test

import (
	"encoding/hex"
	"testing"
)

type environment struct {
	tb  testing.TB
	key []byte
}

func setUp(tb testing.TB) *environment {
	e := &environment{tb: tb}
	e.key, _ = hex.DecodeString("ab72c77b97cb5fe9a382d9fe81ffdbed")
	return e
}
