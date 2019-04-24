[![Documentation](https://godoc.org/github.com/posteo/fader?status.svg)](http://godoc.org/github.com/posteo/fader) [![Go Report Card](https://goreportcard.com/badge/posteo/fader)](https://goreportcard.com/report/posteo/fader) [![Build Status](https://travis-ci.org/posteo/fader.svg?branch=master)](https://travis-ci.org/posteo/fader) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/posteo/fader/blob/master/LICENSE)

# Fader

In-Memory storage that distributes items via UDP multicast.

## Interface

```go
type Fader interface {
    Put([]byte, time.Time, []byte) error
    Get([]byte) (time.Time, []byte)
    Earliest() ([]byte, time.Time, []byte)
    Select([]byte) [][]byte
    Size() int
}
```

A remove function is missing, because every stored item is supposed to expire after a defined
period of time. See the memory fader implementation for details.

## Memory Fader

An implementation of the Fader interface, that stores all items in memory using `container/heap`. The
constructor function takes a `time.Duration` that specifies the period after which a stored item is removed.

```go
memoryFader := fader.NewMemory(1*time.Second)
defer memoryFader.Close()

memoryFader.Put([]byte("key"), time.Now(), []byte("value"))
memoryFader.Size() // => 1

time.Sleep(2*time.Second)
memoryFader.Size() // => 0
```

## Multicast Fader

Another implementation of the Fader interface. It does not store item directly, but delegates them to a given
parent instance. Additionally, every `Store` operation is converted into an UDP packet which is sent to the
given multicast group. The packet is encrypted using the given key.

```go
multicastFaderOne := fader.NewMulticast(memoryFaderOne, "224.0.0.1:1888", key)
defer multicastFaderOne.Close()

multicastFaderTwo := fader.NewMulticast(memoryFaderTwo, "224.0.0.1:1888", key)
defer multicastFaderTwo.Close()

multicastFaderOne.Put([]byte("key"), time.Now(), []byte("value"))
multicastFaderOne.Size() // => 1

time.Sleep(10*time.Millisecond)

multicastFaderTwo.Size() // => 1
```

## Contribution

Any contribution is welcome! Feel free to open an issue or do a pull request.
