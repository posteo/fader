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
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/posteo/fader/crypt"
)

// Multicast implements a multicast faader.
type Multicast struct {
	parent              Fader
	address             string
	key                 []byte
	id                  []byte
	itemReceivedHandler ReceivedHandler
	incomingConnection  *net.UDPConn
	outgoingConnection  *net.UDPConn
	transmitter         *multicastTransmitter
}

// ReceivedHandler defines a handler for received items.
type ReceivedHandler func([]byte, time.Time, []byte) bool

// ErrInvalidKeyLength is returns if a key with an invalid length is provided. Valid lengths
// are 16, 24 and 32.
var ErrInvalidKeyLength = errors.New("invalid key length")

// NewMulticast creates a Fader instance that delegates all calls to a parent Fader instance.
// Additional to that, all store-operations are published to a Multicast group
// which is specified by the given address. All packets will encrypted with
// AES-GCM using the given key. The length of the key's byte-slice, can be 16, 24
// or 32 and will define if AES-128, AES-192 or AES-256 is used.
// A 10-byte long id can be set to avoid collisions. If no id is nil, a random
// id will be generated.
// The argument can take a function that is called every time an item is received.
// If the function is nil or returns true, the received item will be stored in
// the parent fader. Otherwise, the item will be dismissed.
func NewMulticast(
	parent Fader,
	address string,
	key []byte,
	id []byte,
	itemReceivedHandler ReceivedHandler,
) (*Multicast, error) {
	m := &Multicast{
		parent:              parent,
		address:             address,
		key:                 key,
		id:                  id,
		itemReceivedHandler: itemReceivedHandler,
	}

	if length := len(m.key); length != 16 && length != 24 && length != 32 {
		return nil, fmt.Errorf("key length %d: %w", length, ErrInvalidKeyLength)
	}

	udpAddress, err := net.ResolveUDPAddr("udp", m.address)
	if err != nil {
		return nil, fmt.Errorf("resolve udp address [%s]: %w", m.address, err)
	}

	m.incomingConnection, err = net.ListenMulticastUDP("udp", nil, udpAddress)
	if err != nil {
		return nil, fmt.Errorf("listen Multicast udp: %w", err)
	}

	m.outgoingConnection, err = net.DialUDP("udp", nil, udpAddress)
	if err != nil {
		return nil, fmt.Errorf("dial udp: %w", err)
	}

	decrypter, err := crypt.NewDecrypter(m.incomingConnection, m.key)
	if err != nil {
		return nil, fmt.Errorf("new decrypter: %w", err)
	}

	encrypter, err := crypt.NewEncrypter(m.outgoingConnection, m.key)
	if err != nil {
		return nil, fmt.Errorf("new encrypter: %w", err)
	}

	m.transmitter = newMulticastTransmitter(encrypter, decrypter, m.id)

	go m.receiveLoop()

	return m, nil
}

// Put places an item with the provided key, time and value in the fader.
func (m *Multicast) Put(key []byte, time time.Time, value []byte) error {
	if err := m.send(key, time, value); err != nil {
		return fmt.Errorf("send item: %w", err)
	}
	return m.parent.Put(key, time, value)
}

// Get returns time and value for the provided key. If no such key exists, a value
// of nil is returned.
func (m *Multicast) Get(key []byte) (time.Time, []byte) {
	return m.parent.Get(key)
}

// Earliest returns key, time and value of the earliest item in the fader.
func (m *Multicast) Earliest() ([]byte, time.Time, []byte) {
	return m.parent.Earliest()
}

// Select returns all times and values with the provided key.
func (m *Multicast) Select(key []byte) ([]time.Time, [][]byte) {
	return m.parent.Select(key)
}

// Size returns the number of items in the fader.
func (m *Multicast) Size() int {
	return m.parent.Size()
}

// Close tears down the fader.
func (m *Multicast) Close() error {
	if err := m.incomingConnection.Close(); err != nil {
		return fmt.Errorf("close incoming connection: %w", err)
	}
	if err := m.outgoingConnection.Close(); err != nil {
		return fmt.Errorf("close outgoing connection: %w", err)
	}
	return nil
}

func (m *Multicast) send(key []byte, time time.Time, value []byte) error {
	mp := multicastPacket{key, time, value}

	packet, err := mp.MarshalBinary()
	if err != nil {
		return fmt.Errorf("marshal packet: %w", err)
	}

	if _, err := m.transmitter.Write(packet); err != nil {
		return fmt.Errorf("write packet: %w", err)
	}

	if err := m.transmitter.Flush(); err != nil {
		return fmt.Errorf("flush: %w", err)
	}

	return nil
}

func (m *Multicast) receiveLoop() {
	buffer := [2048]byte{}
	mp := &multicastPacket{}
	for {
		n, err := m.transmitter.Read(buffer[:])
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				return
			}
			log.Printf("read error: %v", err)
			continue
		}
		packet := buffer[:n]

		if err := mp.UnmarshalBinary(packet); err != nil {
			log.Printf("unmarshal packet error: %v", err)
			continue
		}

		if m.itemReceivedHandler != nil && !m.itemReceivedHandler(mp.key, mp.time, mp.value) {
			continue
		}

		if err := m.parent.Put(mp.key, mp.time, mp.value); err != nil {
			log.Printf("put into parent fader: %v", err)
			return
		}
	}
}
