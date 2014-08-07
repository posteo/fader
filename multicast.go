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
	"expvar"
	"net"
	"strings"

	"github.com/juju/errgo"
	"github.com/simia-tech/gol"

	"github.com/posteo/fader/crypt"
)

type multicast struct {
	parent             Fader
	address            string
	key                []byte
	ids                [][]byte
	incomingConnection *net.UDPConn
	outgoingConnection *net.UDPConn
	transmitter        *multicastTransmitter
}

var (
	DefaultKey = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	metricSentItems     = expvar.NewInt("event:fader.multicast.sent")
	metricReceivedItems = expvar.NewInt("event:fader.multicast.received")
)

// Creates a Fader instance that delegates all calls to a parent Fader instance.
// Additional to that, all store-operations are published to a multicast group
// which is specified by the given address. All packets will encrypted with
// AES-GCM using the given key. The length of the key's byte-slice, can be 16, 24
// or 32 and will define if AES-128, AES-192 or AES-256 is used.
func NewMulticast(parent Fader, address string, key []byte, ids ...[]byte) Fader {
	return &multicast{
		parent:  parent,
		address: address,
		key:     key,
		ids:     ids,
	}
}

func (m *multicast) Open() error {
	udpAddress, err := net.ResolveUDPAddr("udp", m.address)
	if err != nil {
		return errgo.Mask(err)
	}

	m.incomingConnection, err = net.ListenMulticastUDP("udp", nil, udpAddress)
	if err != nil {
		return errgo.Mask(err)
	}

	m.outgoingConnection, err = net.DialUDP("udp", nil, udpAddress)
	if err != nil {
		return errgo.Mask(err)
	}

	decrypter, err := crypt.NewDecrypter(m.incomingConnection, m.key)
	if err != nil {
		return errgo.Mask(err)
	}

	encrypter, err := crypt.NewEncrypter(m.outgoingConnection, m.key)
	if err != nil {
		return errgo.Mask(err)
	}

	m.transmitter = newMulticastTransmitter(encrypter, decrypter, m.ids...)

	go m.receiveLoop()

	gol.Info("joined multicast group at %s with id %x", m.address, m.transmitter.id)

	return nil
}

func (m *multicast) Close() error {
	if err := m.incomingConnection.Close(); err != nil {
		return errgo.Mask(err)
	}
	if err := m.outgoingConnection.Close(); err != nil {
		return errgo.Mask(err)
	}
	return nil
}

func (m *multicast) Store(item Item) error {
	if err := m.send(item); err != nil {
		return errgo.Mask(err)
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
		return errgo.Mask(err)
	}

	if err := m.transmitter.Flush(); err != nil {
		return errgo.Mask(err)
	}

	metricSentItems.Add(1)

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
			gol.Handle(errgo.Mask(err))
			continue
		}

		metricReceivedItems.Add(1)

		if err := m.parent.Store(item); err != nil {
			gol.Handle(errgo.Mask(err))
			return
		}
	}
}
