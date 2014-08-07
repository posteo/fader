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

package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"io"
	"math/big"

	"github.com/juju/errgo"
)

type encrypter struct {
	parent io.Writer
	aesGCM cipher.AEAD
}

func NewEncrypter(parent io.Writer, key []byte) (Writer, error) {
	aes, err := aes.NewCipher(key)
	if err != nil {
		return nil, errgo.Mask(err)
	}

	aesGCM, err := cipher.NewGCM(aes)
	if err != nil {
		return nil, errgo.Mask(err)
	}

	return &encrypter{
		parent: parent,
		aesGCM: aesGCM,
	}, nil
}

func (e *encrypter) Write(nonce *big.Int, plainText []byte) (int, error) {
	nonceBytes := nonce.Bytes()
	nonceBytes = append(make([]byte, e.aesGCM.NonceSize()-len(nonceBytes)), nonceBytes...)

	cipherText := e.aesGCM.Seal(nil, nonceBytes, plainText, []byte{})

	length := uint16(len(cipherText))
	if err := binary.Write(e.parent, binary.BigEndian, length); err != nil {
		return 0, errgo.Mask(err)
	}

	if _, err := e.parent.Write(nonceBytes); err != nil {
		return 0, errgo.Mask(err)
	}

	if _, err := e.parent.Write(cipherText); err != nil {
		return 0, errgo.Mask(err)
	}

	return len(plainText), nil
}
