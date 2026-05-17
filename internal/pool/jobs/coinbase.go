package jobs

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"

	"brpool/internal/pool/types"
)

// BuildCoinbaseTx constructs a fully valid, SegWit-compliant raw Bitcoin coinbase transaction.
// This version dynamically accepts the raw 32-byte witness commitment hash to build the mandatory
// OP_RETURN ScriptPubKey required for block verification under BIP 141 rules.
func BuildCoinbaseTx(
	tmpl *types.BlockTemplate,
	extraNonce1 string,
	extraNonce2 string,
	witnessCommitment []byte, // 32-byte raw commitment hash calculated from the witness tree
) ([]byte, error) {

	heightBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(
		heightBytes,
		uint32(tmpl.Height),
	)

	// Decode Stratum ExtraNonces from hex strings into raw binary slices
	en1, _ := hex.DecodeString(extraNonce1)
	en2, _ := hex.DecodeString(extraNonce2)

	scriptSig := make([]byte, 0)

	// BIP 34: Prefix the block height bytes with their length push-data byte
	scriptSig = append(scriptSig,
		byte(len(heightBytes)))

	scriptSig = append(scriptSig,
		heightBytes...)

	// Append entropy distribution arrays allocated for the session session
	scriptSig = append(scriptSig,
		en1...)

	scriptSig = append(scriptSig,
		en2...)

	var tx bytes.Buffer

	// Version (4 bytes) - Transaction version 1
	tx.Write([]byte{
		0x01, 0x00, 0x00, 0x00,
	})

	// SegWit Marker & Flag (2 bytes) - Marker: 0x00, Flag: 0x01 (BIP 141 serialization)
	tx.Write([]byte{
		0x00, 0x01,
	})

	// Input Count VarInt - Exactly 1 input for coinbase transactions
	tx.Write(VarInt(1))

	// Outpoint TXID (32 bytes) - All zeros for coinbase inputs
	tx.Write(make([]byte, 32))

	// Outpoint Index (4 bytes) - Fixed to 0xFFFFFFFF
	tx.Write([]byte{
		0xff, 0xff, 0xff, 0xff,
	})

	// Input ScriptSig Size (VarInt)
	tx.Write(VarInt(
		uint64(len(scriptSig)),
	))

	// Input ScriptSig payload (Contains Block Height + ExtraNonces)
	tx.Write(scriptSig)

	// Sequence Number (4 bytes) - Fixed to 0xFFFFFFFF
	tx.Write([]byte{
		0xff, 0xff, 0xff, 0xff,
	})

	// Output Count VarInt - Exactly 2 outputs (Output 0: Pool Payout, Output 1: SegWit Commitment)
	tx.Write(VarInt(2))

	// Output 0: Reward Value (8 bytes) - Total block subsidy + transaction fees in satoshis (Little Endian)
	value := make([]byte, 8)
	binary.LittleEndian.PutUint64(
		value,
		tmpl.CoinbaseValue,
	)
	tx.Write(value)

	// Output 0: Payout PkScript Placeholder (CRITICAL: Currently a dummy 0014... address script)
	// payoutScript, _ := hex.DecodeString(
	// 	"00140000000000000000000000000000000000000000",
	// )
	payoutScript, err := AddressToScript(
	tmpl.PoolAddress,
	)
	if err != nil {
		return nil, err
	}

	tx.Write(VarInt(
		uint64(len(payoutScript)),
	))
	tx.Write(payoutScript)

	// Output 1: Commitment Value (8 bytes) - Set to 0 satoshis as required by BIP 141
	tx.Write(make([]byte, 8))

	// Output 1: Assembling the SegWit Commitment ScriptPubKey dynamically
	commitmentScript := make([]byte, 0)
	commitmentScript = append(
		commitmentScript,
		0x6a, // OP_RETURN
		0x24, // OP_PUSHBYTES_36 (0x24 = 36 bytes payload following)
		0xaa, // SegWit Magic Bytes header prefix (4 bytes: 0xaa21a9ed)
		0x21,
		0xa9,
		0xed,
	)
	// Append the 32-byte witness commitment hash computed from the witness tree
	commitmentScript = append(
		commitmentScript,
		witnessCommitment...,
	)

	// Output 1: Commitment ScriptPubKey Size (VarInt)
	tx.Write(VarInt(
		uint64(len(commitmentScript)),
	))
	tx.Write(commitmentScript)

	// Witness Stack Array - Contains exactly 1 structural stack item
	tx.Write(VarInt(1))

	// Witness Element Size - Exactly 32 bytes for the witness reserved value
	tx.Write(VarInt(32))

	// Witness Element Payload - The 32-byte zero allocation array for coinbase transactions
	tx.Write(make([]byte, 32))

	// Locktime (4 bytes) - Fixed to 0x00000000
	tx.Write([]byte{
		0x00, 0x00, 0x00, 0x00,
	})

	return tx.Bytes(), nil
}

// CoinbaseTxID calculates the legacy (non-witness) Transaction ID (txid) of the coinbase.
// It strips the SegWit serialization markers before hashing, as required by BIP 141 protocol rules.
func CoinbaseTxID(tx []byte) []byte {

	base := StripWitness(tx)

	return DoubleSHA(base)
}

// CoinbaseWTxID calculates the modern Witness Transaction ID (wtxid) of the coinbase.
// It performs a double SHA-256 over the complete SegWit-serialized transaction array.
func CoinbaseWTxID(tx []byte) []byte {

	return DoubleSHA(tx)
}

// StripWitness extracts the traditional transaction body by stripping out 
// the SegWit Marker (byte 4) and Flag (byte 5) from a serialized raw transaction array.
func StripWitness(tx []byte) []byte {

	if len(tx) < 6 {
		return tx
	}

	// Verify if the transaction layout matches SegWit dynamic formatting (0x0001 marker)
	if tx[4] != 0x00 ||
		tx[5] != 0x01 {
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