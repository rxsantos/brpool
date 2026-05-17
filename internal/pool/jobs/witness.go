package jobs

import (
	"crypto/sha256"
	"encoding/hex"

	"brpool/internal/pool/types"
)

// DoubleSHA computes SHA-256(SHA-256(b)), which is the standard cryptographic hashing algorithm used in Bitcoin consensus.
func DoubleSHA(b []byte) []byte {

	h1 := sha256.Sum256(b)
	h2 := sha256.Sum256(h1[:])

	return h2[:]
}

// ReverseBytes performs an in-place byte-order reversal (used for Little-Endian/Big-Endian conversions in Bitcoin).
func ReverseBytes(b []byte) {

	for i := 0; i < len(b)/2; i++ {

		j := len(b) - i - 1

		b[i], b[j] = b[j], b[i]
	}
}

// BuildWitnessMerkleRoot calculates the SegWit Witness Merkle Root (BIP 141).
// Instead of using legacy TxIDs, this tree structures data using Witness Transaction IDs (wtxid),
// guaranteeing cryptographic verification of signatures (witness data) inside the block.
func BuildWitnessMerkleRoot(
	coinbaseWTxID []byte,
	// coinbaseWTxID := make([]byte, 32)
	tmpl *types.BlockTemplate,
) ([]byte, error) {

	hashes := make([][]byte, 0)

	// BIP 141 rule: The coinbase wtxid is defined as 32-bytes of all zeros (0x00...00)
	hashes = append(hashes,
		coinbaseWTxID)

	// Iterate through all mempool transactions to extract their wtxid mapping
	for _, tx := range tmpl.Transactions {

		hashHex := tx.Hash

		// If a transaction has no SegWit data, its legacy TxID is used as its wtxid representation
		if hashHex == "" {
			hashHex = tx.TxID
		}

		hash, err := hex.DecodeString(
			hashHex)
		if err != nil {
			return nil, err
		}

		// Reverse internal RPC hex representation bytes to binary consensus structure order
		ReverseBytes(hash)

		hashes = append(hashes,
			hash)
	}

	// Collapse the binary hash tree iteratively upwards to discover the final single root
	for len(hashes) > 1 {

		var next [][]byte

		for i := 0; i < len(hashes); i += 2 {

			left := hashes[i]

			right := left // Duplicate left node if odd counts occur

			if i+1 < len(hashes) {
				right = hashes[i+1]
			}
			combined := make([]byte, 0, 64)

			combined = append(combined, left...)
			combined = append(combined, right...)
			// newHash := DoubleSHA(
			// 	append(left, right...))
			newHash := DoubleSHA(combined)

			next = append(next,
				newHash)
		}

		hashes = next
	}

	// Return the completed Witness Merkle Root hash
	return hashes[0], nil
}

// BuildWitnessCommitment hashes the combination of the Witness Merkle Root 
// and the 32-byte witness reserved value to produce the exact structural commitment array 
// that gets injected inside the coinbase transaction vout.
func BuildWitnessCommitment(
	witnessMerkleRoot []byte,
) []byte {

	// BIP 141 requires a fixed 32-byte witness reserved value (all zeros)
	witnessReserved := make([]byte, 32)

	// Combine the 32-byte Witness Merkle Root with the 32-byte Witness Reserved Value
	// data := append(
	// 	witnessMerkleRoot,
	// 	witnessReserved...,
	// )
	data := make([]byte, 0, 64)

	data = append(data, witnessMerkleRoot...)
	data = append(data, witnessReserved...)

	// Perform the final double SHA-256 over the combined 64-byte sequence
	return DoubleSHA(data)
}