package jobs

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"

	"brpool/internal/pool/types"
)

// BuildCoinbaseTx constructs a fully valid, SegWit-compliant raw Bitcoin coinbase transaction.
// It encodes the block height (BIP 34), embeds the Stratum extraNonces inside the scriptSig,
// maps the block reward output, and appends the mandatory SegWit commitment script.
func BuildCoinbaseTx(
	tmpl *types.BlockTemplate,
	extraNonce1 string,
	extraNonce2 string,
) ([]byte, error) {

	heightBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(
		heightBytes,
		uint32(tmpl.Height),
	)

	// Decode the Stratum ExtraNonces from hex strings to raw binary bytes
	en1, _ := hex.DecodeString(extraNonce1)
	en2, _ := hex.DecodeString(extraNonce2)

	scriptSig := make([]byte, 0)

	// BIP 34 rule: The scriptSig must begin with a push-data length byte for the block height
	scriptSig = append(scriptSig,
		byte(len(heightBytes)))

	scriptSig = append(scriptSig,
		heightBytes...)

	// Append entropy fields allocated to the pool (en1) and individual miner sessions (en2)
	scriptSig = append(scriptSig,
		en1...)

	scriptSig = append(scriptSig,
		en2...)

	// BIP 141 (SegWit): The coinbase transaction must define a 32-byte dummy witness reserved value (all zeros)
	witnessReserved := make([]byte, 32)

	var tx bytes.Buffer

	// Version (4 bytes) - Fixed to transaction version 1
	tx.Write([]byte{
		0x01, 0x00, 0x00, 0x00,
	})

	// SegWit Marker & Flag (2 bytes) - Marker: 0x00, Flag: 0x01 (Signals presence of witness serialized data)
	tx.Write([]byte{0x00, 0x01})

	// Input Count VarInt - Coinbase always contains exactly 1 execution input
	tx.Write(VarInt(1))

	// Outpoint TXID (32 bytes) - Hardcoded to all zeros for coinbase transactions
	prevHash := make([]byte, 32)
	tx.Write(prevHash)

	// Outpoint Index (4 bytes) - Hardcoded to 0xFFFFFFFF for coinbase inputs
	tx.Write([]byte{
		0xff, 0xff, 0xff, 0xff,
	})

	// Input ScriptSig Size (VarInt)
	tx.Write(VarInt(
		uint64(len(scriptSig)),
	))

	// Input ScriptSig payload (Contains Block Height + ExtraNonces)
	tx.Write(scriptSig)

	// Sequence Number (4 bytes) - Set to 0xFFFFFFFF
	tx.Write([]byte{
		0xff, 0xff, 0xff, 0xff,
	})

	// Output Count VarInt - Set to 2 outputs (Output 0: Pool Payout, Output 1: SegWit Commitment)
	tx.Write(VarInt(2))

	// Output 0: Reward Value (8 bytes) - Total block subsidy + fees in satoshis (Little Endian)
	value := make([]byte, 8)
	binary.LittleEndian.PutUint64(
		value,
		tmpl.CoinbaseValue,
	)
	tx.Write(value)

	// Output 0: Payout PkScript Placeholder (CRITICAL: Currently a dummy 0014... address script)
	payoutScript, _ := hex.DecodeString(
		"00140000000000000000000000000000000000000000",
	)

	tx.Write(VarInt(
		uint64(len(payoutScript)),
	))
	tx.Write(payoutScript)

	// Output 1: Commitment Value (8 bytes) - Set to 0 satoshis as required by BIP 141
	tx.Write(make([]byte, 8))

	// Output 1: Witness Commitment ScriptPubKey provided by Bitcoin Core template
	witnessScript, _ := hex.DecodeString(
		tmpl.DefaultWitnessCommitment,
	)

	tx.Write(VarInt(
		uint64(len(witnessScript)),
	))
	tx.Write(witnessScript)

	// Witness Stack Array - Contains exactly 1 structural stack item
	tx.Write(VarInt(1))

	// Witness Element Size - Exactly 32 bytes for the witness reserved value
	tx.Write(VarInt(32))

	// Witness Element Payload - The 32-byte zero allocation array
	tx.Write(witnessReserved)

	// Locktime (4 bytes) - Set to standard immediate execution sequence (0x00000000)
	tx.Write([]byte{
		0x00, 0x00, 0x00, 0x00,
	})

	return tx.Bytes(), nil
}

// CoinbaseTxID calculates the legacy (non-witness) Transaction ID (txid) of the coinbase.
// It strips the SegWit serialization markers before hashing, as required by BIP 141 protocol rules.
func CoinbaseTxID(tx []byte) []byte {

	// Strip marker and flag fields to preserve legacy hashing structures
	base := stripWitness(tx)

	// Compute standard SHA-256D over the non-witness structure
	h1 := sha256.Sum256(base)
	h2 := sha256.Sum256(h1[:])

	return h2[:]
}

// stripWitness extracts the traditional transaction body by stripping out 
// the SegWit Marker (byte 4) and Flag (byte 5) from a serialized raw transaction array.
func stripWitness(tx []byte) []byte {

	if len(tx) < 6 {
		return tx
	}

	// Verify if the transaction layout matches SegWit dynamic formatting (0x0001 marker)
	if tx[4] != 0x00 || tx[5] != 0x01 {
		return tx
	}

	out := make([]byte, 0)

	// Slice and bridge data around bytes 4 and 5
	out = append(out,
		tx[:4]...) // Keeps Version fields

	out = append(out,
		tx[6:]...) // Appends Inputs, Outputs and Locktime fields directly

	return out
}