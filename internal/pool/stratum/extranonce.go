package stratum

import (
	"encoding/binary"
	"encoding/hex"
	"sync/atomic"
)

// extraNonceCounter is a thread-safe global concurrent counter using sync/atomic.
// It guarantees that every calling goroutine receives a unique sequential integer,
// preventing duplicate nonce state allocation across concurrent miner sessions.
var extraNonceCounter atomic.Uint32

// NextExtraNonce1 increments the global counter safely and returns a unique 4-byte 
// hexadecimal string representation formatted in Big-Endian.
// This is used as the unique dynamic ExtraNonce1 handshake token for Stratum clients.
func NextExtraNonce1() string {

	// Increment and fetch the new unique value atomically
	n := extraNonceCounter.Add(1)

	b := make([]byte, 4)

	// Serialize the uint32 counter into a 4-byte raw binary array using Big-Endian order
	binary.BigEndian.PutUint32(
		b,
		n,
	)

	// Return the encoded text representation to be wrapped into the 'mining.subscribe' JSON response
	return hex.EncodeToString(b)
}