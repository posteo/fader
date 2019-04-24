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
	"time"
)

type item struct {
	KeyField  string
	TimeField time.Time
}

func (i *item) Key() string {
	return i.KeyField
}

func (i *item) Time() time.Time {
	return i.TimeField
}

func (i *item) MarshalBinary() ([]byte, error) {
	td, err := i.TimeField.MarshalBinary()
	if err != nil {
		return nil, err
	}
	return append(td, []byte(i.Key())...), nil
}

func (i *item) UnmarshalBinary(data []byte) error {
	if err := i.TimeField.UnmarshalBinary(data[:15]); err != nil {
		return err
	}
	i.KeyField = string(data[15:])
	return nil
}
