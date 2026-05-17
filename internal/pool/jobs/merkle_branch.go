package jobs

import (
	"encoding/hex"

	"brpool/internal/pool/types"
)

// BuildMerkleBranches extracts the necessary transaction hashes from the block template
// to construct the Merkle steps (branches) required by the Stratum protocol.
// These branches allow micro-miners to build their own local block header hashes efficiently.
func BuildMerkleBranches(
	tmpl *types.BlockTemplate,
) ([]string, error) {

	var branches []string

	// CRITICAL STRATUM BUG: This loop merely lists all transaction IDs without building the tree hierarchy.
	for _, tx := range tmpl.Transactions {

		txid, err := hex.DecodeString(
			tx.TxID)
		if err != nil {
			return nil, err
		}

		// Reverse internal RPC bytes to match consensus ordering
		ReverseBytes(txid)

		// Append the reversed hex string to the branch slice
		branches = append(branches,
			hex.EncodeToString(txid))
	}

	return branches, nil
}