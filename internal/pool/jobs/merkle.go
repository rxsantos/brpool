package jobs

import (
	"crypto/sha256"
	"encoding/hex"

	"brpool/internal/pool/types"
)

// reverseBytes performs an in-place byte-order reversal (used for Little-Endian/Big-Endian conversions in Bitcoin).
func reverseBytes(b []byte) {

	for i := 0; i < len(b)/2; i++ {

		j := len(b) - i - 1

		b[i], b[j] = b[j], b[i]
	}
}

// doubleSHA computes SHA-256(SHA-256(b)), which is the standard cryptographic hashing algorithm used in Bitcoin consensus.
func doubleSHA(b []byte) []byte {

	h1 := sha256.Sum256(b)
	h2 := sha256.Sum256(h1[:])

	return h2[:]
}

// BuildMerkleRoot calculates the mathematical Merkle Root of all transactions inside a block candidate.
// It explicitly injects the freshly generated coinbase hash as the first leaf (index 0) of the tree.
func BuildMerkleRoot(
	coinbaseHash []byte,
	tmpl *types.BlockTemplate,
) ([]byte, error) {

	hashes := make([][]byte, 0)

	// The coinbase transaction hash is natively the first element in the Bitcoin Merkle Tree structure
	hashes = append(hashes,
		coinbaseHash)

	// Decode and byte-reverse all pending mempool transaction IDs from the block template
	for _, tx := range tmpl.Transactions {

		txid, err := hex.DecodeString(
			tx.TxID)
		if err != nil {
			return nil, err
		}

		// Bitcoin transaction IDs inside full nodes are represented in RPC byte-reversed string formats
		reverseBytes(txid)

		hashes = append(hashes,
			txid)
	}

	// Iteratively pair adjacent hashes and double-hash them to collapse the tree upwards
	var merkleSteps []string
	for len(hashes) > 1 {
		if len(hashes) > 1 {
				merkleSteps = append(merkleSteps, hex.EncodeToString(hashes[1]))
			}
		var next [][]byte

		for i := 0; i < len(hashes); i += 2 {

			left := hashes[i]

			right := left // BIP rule: If an odd number of elements exists, duplicate the last transaction leaf

			if i+1 < len(hashes) {
				right = hashes[i+1]
			}

			// Concatenate left + right nodes and perform SHA-256D
			combined := make([]byte, 0, 64)

			combined = append(combined, left...)
			combined = append(combined, right...)
			// hash := doubleSHA(
			// 	append(left, right...))
			hash := doubleSHA(combined)

			next = append(next,
				hash)
		}

		// Move up one level in the Merkle Tree hierarchy
		hashes = next
	}

	// Return the final single remaining hash: the Merkle Root
	return hashes[0], nil
}