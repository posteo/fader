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
	"bytes"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/posteo/fader/crypt"
)

func TestDecryption(t *testing.T) {
	input, _ := hex.DecodeString("001800000000000000000000000048d484579c9da1845613bcb0b13154268384ffba962cd4d7")
	inputBuffer := bytes.NewBuffer(input)
	decrypter, err := crypt.NewDecrypter(inputBuffer, key)
	require.NoError(t, err)

	nonce := big.NewInt(0)
	plainText := make([]byte, 8)
	n, err := decrypter.Read(nonce, plainText)
	require.NoError(t, err)
	assert.Equal(t, 8, n)
	assert.Equal(t, []byte{1, 2, 3, 4, 5, 6, 7, 8}, plainText)
}

func TestCorrectNonceReading(t *testing.T) {
	input, _ := hex.DecodeString("001800000000000000000021e88e57ca9ec99d535f2c5915a084191e59c343125c26142b7fff")
	inputReader := bytes.NewReader(input)
	decrypter, err := crypt.NewDecrypter(inputReader, key)
	require.NoError(t, err)

	nonce := big.NewInt(0)
	plainText := make([]byte, 8)
	n, err := decrypter.Read(nonce, plainText)
	require.NoError(t, err)
	assert.Equal(t, 8, n)
	assert.Equal(t, big.NewInt(2222222), nonce)
	assert.Equal(t, []byte{1, 2, 3, 4, 5, 6, 7, 8}, plainText)
}
