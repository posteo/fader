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
	"crypto/rand"
	"fmt"
	"log"
	"math/big"

	"github.com/posteo/fader/crypt"
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

func newMulticastTransmitter(writer crypt.Writer, reader crypt.Reader, id []byte) *multicastTransmitter {
	if id == nil || len(id) != 10 {
		id = randomBytes(idSize)
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
		log.Printf("send an udp multicast packet of size %d, should not exceed %d",
			t.writeBuffer.Len(), maximalWriteBufferSize)
	}

	buffer := append(t.id, t.writeBuffer.Bytes()...)
	if _, err := t.writer.Write(t.nonce, buffer); err != nil {
		t.increaseNonce()
		return fmt.Errorf("write: %w", err)
	}
	t.writeBuffer.Reset()

	t.increaseNonce()
	return nil
}

func (t *multicastTransmitter) Read(payload []byte) (int, error) {
	buffer := make([]byte, idSize+len(payload))
	packet := []byte{}

	nonce := big.NewInt(0)
	for {
		n, err := t.reader.Read(nonce, buffer)
		if err != nil {
			return 0, fmt.Errorf("read: %w", err)
		}
		packet = buffer[:n]

		id := packet[:idSize]
		if bytes.Equal(t.id, id) {
			continue
		}

		if !t.validNonce(id, nonce) {
			continue
		}

		break
	}

	return copy(payload, packet[idSize:]), nil
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

func randomBytes(count int) []byte {
	result := make([]byte, count)
	if _, err := rand.Read(result); err != nil {
		panic(err)
	}
	return result
}
