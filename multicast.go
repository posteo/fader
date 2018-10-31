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
	"encoding/gob"
	"log"
	"net"
	"strings"

	"github.com/simia-tech/errx"

	"github.com/posteo/fader/crypt"
)

type multicast struct {
	parent              Fader
	address             string
	key                 []byte
	id                  []byte
	itemReceivedHandler func(Item) bool
	incomingConnection  *net.UDPConn
	outgoingConnection  *net.UDPConn
	transmitter         *multicastTransmitter
}

// NewMulticast creates a Fader instance that delegates all calls to a parent Fader instance.
// Additional to that, all store-operations are published to a multicast group
// which is specified by the given address. All packets will encrypted with
// AES-GCM using the given key. The length of the key's byte-slice, can be 16, 24
// or 32 and will define if AES-128, AES-192 or AES-256 is used.
// For testing purposes, a 10-byte long id can be set. If no id is nil, a random
// id will be generated.
// The argument can take a function that is called every time an item is received.
// If the function is nil or returns true, the received item will be stored in
// the parent fader. Otherwise, the item will be dismissed.
func NewMulticast(
	parent Fader,
	address string,
	key []byte,
	id []byte,
	itemReceivedHandler func(Item) bool,
) Fader {
	return &multicast{
		parent:              parent,
		address:             address,
		key:                 key,
		id:                  id,
		itemReceivedHandler: itemReceivedHandler,
	}
}

func (m *multicast) Open() error {
	if length := len(m.key); length != 16 && length != 24 && length != 32 {
		return errx.Errorf("key has length %d, but must have a length of 16, 24 or 32", length)
	}

	udpAddress, err := net.ResolveUDPAddr("udp", m.address)
	if err != nil {
		return errx.Annotatef(err, "resolve udp address [%s]", m.address)
	}

	m.incomingConnection, err = net.ListenMulticastUDP("udp", nil, udpAddress)
	if err != nil {
		return errx.Annotatef(err, "listen multicast udp")
	}

	m.outgoingConnection, err = net.DialUDP("udp", nil, udpAddress)
	if err != nil {
		return errx.Annotatef(err, "dial udp")
	}

	decrypter, err := crypt.NewDecrypter(m.incomingConnection, m.key)
	if err != nil {
		return errx.Annotatef(err, "new decrypter")
	}

	encrypter, err := crypt.NewEncrypter(m.outgoingConnection, m.key)
	if err != nil {
		return errx.Annotatef(err, "new encrypter")
	}

	m.transmitter = newMulticastTransmitter(encrypter, decrypter, m.id)

	go m.receiveLoop()

	return nil
}

func (m *multicast) Close() error {
	if err := m.incomingConnection.Close(); err != nil {
		return errx.Annotatef(err, "close incoming connection")
	}
	if err := m.outgoingConnection.Close(); err != nil {
		return errx.Annotatef(err, "close outgoing connection")
	}
	return nil
}

func (m *multicast) Store(item Item) error {
	if err := m.send(item); err != nil {
		return errx.Annotatef(err, "send item")
	}
	return m.parent.Store(item)
}

func (m *multicast) Earliest() Item {
	return m.parent.Earliest()
}

func (m *multicast) Select(key string) []Item {
	return m.parent.Select(key)
}

func (m *multicast) Detect(key string) Item {
	return m.parent.Detect(key)
}

func (m *multicast) Size() int {
	return m.parent.Size()
}

func (m *multicast) send(item Item) error {
	encoder := gob.NewEncoder(m.transmitter)

	if err := encoder.Encode(&item); err != nil {
		return errx.Annotatef(err, "encode item")
	}

	if err := m.transmitter.Flush(); err != nil {
		return errx.Annotatef(err, "flush")
	}

	return nil
}

func (m *multicast) receiveLoop() {
	var item Item
	for {
		decoder := gob.NewDecoder(m.transmitter)
		if err := decoder.Decode(&item); err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				return
			}
			log.Printf("error during message decoding: %v", err)
			continue
		}

		if m.itemReceivedHandler != nil && !m.itemReceivedHandler(item) {
			continue
		}

		if err := m.parent.Store(item); err != nil {
			log.Printf("error during message storing: %v", err)
			return
		}
	}
}
