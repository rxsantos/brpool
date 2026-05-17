package jobs

import "encoding/binary"

// VarInt (Variable Length Integer), also known as CompactSize in Bitcoin Core,
// serializes a uint64 integer into a dynamic byte array following Bitcoin's consensus serialization rules.
// This is heavily used for encoding transaction counts and script sizes in raw block structures.
func VarInt(n uint64) []byte {

	// If the value is less than 253 (0xFD), it fits into a single, direct byte.
	if n < 0xfd {
		return []byte{byte(n)}
	}

	// If the value fits in 2 bytes (uint16), prefix it with 0xFD followed by the Little-Endian bytes.
	if n <= 0xffff {

		b := make([]byte, 3)

		b[0] = 0xfd

		binary.LittleEndian.PutUint16(
			b[1:], uint16(n))

		return b
	}

	// If the value fits in 4 bytes (uint32), prefix it with 0xFE followed by the Little-Endian bytes.
	if n <= 0xffffffff {

		b := make([]byte, 5)

		b[0] = 0xfe

		binary.LittleEndian.PutUint32(
			b[1:], uint32(n))

		return b
	}

	// For any larger value up to uint64, prefix it with 0xFF followed by the Little-Endian bytes.
	b := make([]byte, 9)

	b[0] = 0xff

	binary.LittleEndian.PutUint64(
		b[1:], n)

	return b
}