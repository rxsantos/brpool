package jobs

import (
	"fmt"

	"github.com/btcsuite/btcd/btcutil"
)

// AddressToScript converts a human-readable Bitcoin address string into its corresponding 
// binary scriptPubKey byte array. This script is required to lock the block reward (coinbase vout[0]) 
// to the pool's destination wallet.
func AddressToScript(address string) ([]byte, error) {

	// Decode the alphanumeric address string using btcd's utility library.
	// The second parameter is 'nil' to automatically detect the network type (Mainnet, Testnet, or Regtest).
	addr, err := btcutil.DecodeAddress(
		address,
		nil,
	)
	if err != nil {
		return nil, err
	}

	// Evaluate the internal concrete implementation type of the decoded address
	switch a := addr.(type) {

	// Native SegWit Pay-to-Witness-PubKey-Hash (P2WPKH) addresses (Bech32 starting with 'bc1q')
	case *btcutil.AddressWitnessPubKeyHash:

		// Extract the 20-byte public key hash (RIPEMD-160 of the SHA-256 of the public key)
		hash := a.Hash160()

		script := make([]byte, 0)

		// BIP 141 P2WPKH scriptPubKey structure: OP_0 (0x00) followed by a 20-byte push operator (0x14)
		script = append(script,
			0x00,
			0x14,
		)

		// Append the actual 20 bytes of the public key hash payload
		script = append(script,
			hash[:]...,
		)

		return script, nil
	}

	// Fallback error if the miner or config provides an unsupported script standard (e.g., Legacy P2PKH, P2SH, or Taproot)
	return nil, fmt.Errorf(
		"unsupported address type")
}