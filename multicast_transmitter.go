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
	"bytes"
	"expvar"
	"math/big"

	"github.com/juju/errgo"
	"github.com/simia-tech/gol"

	"github.com/posteo/fader/crypt"
	"code.posteo.de/shared/util"
)

const (
	idSize                 = 10
	maximalWriteBufferSize = 512
)

type multicastTransmitter struct {
	writer        crypt.Writer
	reader        crypt.Reader
	writeBuffer   *bytes.Buffer
	id            []byte
	nonce         *big.Int
	foreignNonces map[string]*big.Int
}

var (
	metricDroppedPackets = expvar.NewInt("event:fader.multicast.dropped")
)

func newMulticastTransmitter(writer crypt.Writer, reader crypt.Reader, ids ...[]byte) *multicastTransmitter {
	id := util.RandomBytes(idSize)
	if len(ids) > 0 {
		copy(id, ids[0])
	}
	return &multicastTransmitter{
		writer:        writer,
		reader:        reader,
		writeBuffer:   &bytes.Buffer{},
		id:            id,
		nonce:         big.NewInt(0),
		foreignNonces: make(map[string]*big.Int),
	}
}

func (t *multicastTransmitter) Write(payload []byte) (int, error) {
	return t.writeBuffer.Write(payload)
}

func (t *multicastTransmitter) Flush() error {
	if t.writeBuffer.Len() > maximalWriteBufferSize {
		gol.Warning("send an udp multicast packet of size %d, should not exceed %d",
			t.writeBuffer.Len(), maximalWriteBufferSize)
	}

	if _, err := t.writer.Write(t.nonce, append(t.id, t.writeBuffer.Bytes()...)); err != nil {
		t.increaseNonce()
		return errgo.Mask(err)
	}
	t.writeBuffer.Reset()

	t.increaseNonce()
	return nil
}

func (t *multicastTransmitter) Read(payload []byte) (int, error) {
	buffer := make([]byte, idSize+len(payload))

	nonce := big.NewInt(0)
	for {
		if _, err := t.reader.Read(nonce, buffer); err != nil {
			return 0, errgo.Mask(err)
		}

		id := buffer[:idSize]
		if bytes.Equal(t.id, id) {
			continue
		}

		if !t.validNonce(id, nonce) {
			metricDroppedPackets.Add(1)
			continue
		}

		break
	}

	copy(payload, buffer[idSize:])
	return len(payload), nil
}

func (t *multicastTransmitter) increaseNonce() {
	t.nonce = t.nonce.Add(t.nonce, big.NewInt(1))
}

func (t *multicastTransmitter) validNonce(id []byte, nonce *big.Int) bool {
	foreignNonce, found := t.foreignNonces[string(id)]
	if found {
		switch nonce.Cmp(foreignNonce) {
		case -1, 0:
			return false
		case 1:
			foreignNonce.Set(nonce)
			return true
		}
	} else {
		t.foreignNonces[string(id)] = nonce
	}
	return true
}
