package jobs

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"brpool/internal/pool/types"
)

// RandomJobID generates a random 4-byte hexadecimal string to uniquely identify a mining job.
func RandomJobID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// reverseBytes performs an in-place byte-order reversal (used for Little-Endian/Big-Endian conversions in Bitcoin).
func reverseBytes(b []byte) {
	for i := 0; i < len(b)/2; i++ {
		j := len(b) - i - 1
		b[i], b[j] = b[j], b[i]
	}
}

// doubleSHA computes SHA-256(SHA-256(b)), which is the standard hashing algorithm used throughout Bitcoin.
func doubleSHA(b []byte) []byte {
	h1 := sha256.Sum256(b)
	h2 := sha256.Sum256(h1[:])
	return h2[:]
}

// BuildCoinbase constructs a raw Bitcoin coinbase transaction based on BIP 34 (Block Height in Coinbase)
// and appends the Stratum ExtraNonce fields for miner unique work distribution.
func BuildCoinbase(
	tmpl *types.BlockTemplate,
	extraNonce1 string,
	extraNonce2 string,
) ([]byte, error) {

	height := uint32(tmpl.Height)

	// CRITICAL FIX REQUIRED: Block height serialization must comply with BIP 34 (Script pushdata rules)
	heightBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(heightBytes, height)

	scriptSig := append(heightBytes,
		[]byte(extraNonce1+extraNonce2)...)

	coinbase := make([]byte, 0)

	// Version (4 bytes) - Transaction version 1
	coinbase = append(coinbase,
		0x01, 0x00, 0x00, 0x00)

	// Input Count (1 byte) - Coinbase always has exactly 1 input
	coinbase = append(coinbase, 0x01)

	// Outpoint TXID (32 bytes) - All zeros for coinbase inputs
	for i := 0; i < 32; i++ {
		coinbase = append(coinbase, 0x00)
	}

	// Outpoint Index (4 bytes) - All 0xFF for coinbase inputs
	coinbase = append(coinbase,
		0xff, 0xff, 0xff, 0xff)

	// ScriptSig Length (1 byte)
	coinbase = append(coinbase,
		byte(len(scriptSig)))

	// ScriptSig (Contains Block Height and ExtraNonces)
	coinbase = append(coinbase,
		scriptSig...)

	// Sequence Number (4 bytes) - Fixed to 0xFFFFFFFF
	coinbase = append(coinbase,
		0xff, 0xff, 0xff, 0xff)

	// Output Count (1 byte) - CRITICAL FIX REQUIRED: Hardcoded to 1 output, ignoring SegWit commitments
	coinbase = append(coinbase, 0x01)

	// Output Value (8 bytes) - Total block reward in satoshis (Little Endian)
	value := make([]byte, 8)
	binary.LittleEndian.PutUint64(value,
		tmpl.CoinbaseValue)

	coinbase = append(coinbase,
		value...)

	// Output PkScript (Placeholder) - CRITICAL FIX REQUIRED: Hardcoded dummy address, must use config.PoolAddress
	pkScript, err := hex.DecodeString(
		"00140000000000000000000000000000000000000000")
	if err != nil {
		return nil, err
	}

	coinbase = append(coinbase,
		byte(len(pkScript)))

	coinbase = append(coinbase,
		pkScript...)

	// Locktime (4 bytes) - Transaction locktime (0x00000000)
	coinbase = append(coinbase,
		0x00, 0x00, 0x00, 0x00)

	return coinbase, nil
}

// BuildMerkleRoot calculates the cryptographic Merkle Root of all transactions in the block template,
// starting with the freshly generated coinbase transaction hash.
func BuildMerkleRoot(
	coinbaseHash []byte,
	tmpl *types.BlockTemplate,
) ([]byte, error) {

	hashes := make([][]byte, 0)

	// The coinbase transaction hash is always the first element in the Merkle Tree
	hashes = append(hashes, coinbaseHash)

	// Append and byte-reverse all other transactions from the mempool/template
	for _, tx := range tmpl.Transactions {

		txHash, err := hex.DecodeString(tx.TxID)
		if err != nil {
			return nil, err
		}

		reverseBytes(txHash)

		hashes = append(hashes, txHash)
	}

	// Iteratively pair and double-hash until only one single root hash remains
	for len(hashes) > 1 {

		var newLevel [][]byte

		for i := 0; i < len(hashes); i += 2 {

			left := hashes[i]

			right := left // If odd number of hashes, duplicate the last one

			if i+1 < len(hashes) {
				right = hashes[i+1]
			}

			combined := append(left, right...)

			newHash := doubleSHA(combined)

			newLevel = append(newLevel,
				newHash)
		}

		hashes = newLevel
	}

	return hashes[0], nil
}

// CreateJob instantiates a basic Stratum job skeleton from a given Bitcoin block template.
// CRITICAL NOTE: This does not yet compute or slice the Coinbase halves or Merkle Branches for Stratum clients.
func CreateJob(
	tmpl *types.BlockTemplate,
) *types.Job {

	return &types.Job{
		JobID: RandomJobID(),

		// CRITICAL FIX REQUIRED: Previous block hash must be byte-reversed for Stratum presentation
		PrevHash: tmpl.PreviousBlockHash,

		// CRITICAL FIX REQUIRED: Version and NTime must be formatted as Little-Endian hex strings for Stratum
		Version: fmt.Sprintf("%08x",
			tmpl.Version),

		Bits: tmpl.Bits,

		NTime: fmt.Sprintf("%08x",
			tmpl.CurTime),

		CleanJobs: true,

		BlockTemplate: tmpl,
	}
}