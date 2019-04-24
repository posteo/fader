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

package fader

import (
	"encoding/binary"
	"time"
)

type multicastPacket struct {
	key   []byte
	time  time.Time
	value []byte
}

func (mp *multicastPacket) MarshalBinary() ([]byte, error) {
	buffer := make([]byte, 2+len(mp.key)+15+2+len(mp.value))

	index := 0
	binary.BigEndian.PutUint16(buffer[index:index+2], uint16(len(mp.key)))
	index += 2

	index += copy(buffer[index:index+len(mp.key)], mp.key)

	time, _ := mp.time.MarshalBinary()
	index += copy(buffer[index:index+15], time)

	binary.BigEndian.PutUint16(buffer[index:index+2], uint16(len(mp.value)))
	index += 2
	copy(buffer[index:], mp.value)

	return buffer, nil
}

func (mp *multicastPacket) UnmarshalBinary(buffer []byte) error {
	index := 0

	keySize := int(binary.BigEndian.Uint16(buffer[index : index+2]))
	index += 2

	mp.key = make([]byte, keySize)
	index += copy(mp.key, buffer[index:index+keySize])

	if err := mp.time.UnmarshalBinary(buffer[index : index+15]); err != nil {
		return err
	}
	index += 15

	valueSize := int(binary.BigEndian.Uint16(buffer[index : index+2]))
	index += 2

	mp.value = make([]byte, valueSize)
	copy(mp.value, buffer[index:index+valueSize])

	return nil
}
