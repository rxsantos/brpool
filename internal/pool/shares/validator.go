package shares

import (
	"bytes"
	"crypto/sha256"
	// "encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"

	"brpool/internal/pool/bitcoin"
	"brpool/internal/pool/jobs"
	"brpool/internal/pool/types"
)

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

// CompactToTarget converts the Bitcoin 'bits' compact format (uint32 representation) 
// into a full 256-bit big.Int target value for difficulty comparison.
func CompactToTarget(bits string) (*big.Int, error) {

	b, err := hex.DecodeString(bits)
	if err != nil {
		return nil, err
	}

	exponent := b[0]
	mantissa := b[1:]

	target := new(big.Int).SetBytes(mantissa)

	// Shift target left by 8 * (exponent - 3) bits as per Bitcoin target computation rules
	shift := 8 * (uint(exponent) - 3)

	target.Lsh(target, shift)

	return target, nil
}

// ValidateShare processes the 'mining.submit' parameters sent by a Stratum client.
// It reconstructs the block header submitted by the miner, calculates the double SHA-256 hash,
// validates it against the share/block difficulty target, and submits valid blocks to the Bitcoin node.
func ValidateShare(
	params []interface{},
	job *types.Job,
	rpc *bitcoin.RPCClient,
) error {

	// Stratum mining.submit usually sends: [worker_name, job_id, extranonce2, ntime, nonce]
	if len(params) < 5 {
		return fmt.Errorf("invalid submit")
	}

	extraNonce2 := params[2].(string)
	ntime := params[3].(string)
	nonce := params[4].(string)

	// CRITICAL FIX REQUIRED: Hardcoded extraNonce1. This must be dynamically tracked per connected Miner session.
	extraNonce1 := "00000001"

	// Rebuild the customized coinbase transaction using the miner's submitted extraNonce2
	coinbase, err := jobs.BuildCoinbase(
		job.BlockTemplate,
		extraNonce1,
		extraNonce2,
	)
	if err != nil {
		return err
	}

	coinbaseHash := doubleSHA(coinbase)

	// Recalculate the Merkle Root with the new coinbase hash injected as the first leaf
	merkleRoot, err := jobs.BuildMerkleRoot(
		coinbaseHash,
		job.BlockTemplate,
	)
	if err != nil {
		return err
	}

	// CRITICAL FIX REQUIRED: In Stratum, the reconstructed elements must be handled with precise endianness mapping
	reverseBytes(merkleRoot)

	prevHash, err := hex.DecodeString(
		job.PrevHash)
	if err != nil {
		return err
	}

	reverseBytes(prevHash)

	version, _ := hex.DecodeString(
		job.Version)

	reverseBytes(version)

	timeBytes, _ := hex.DecodeString(
		ntime)

	reverseBytes(timeBytes)

	bitsBytes, _ := hex.DecodeString(
		job.Bits)

	reverseBytes(bitsBytes)

	nonceBytes, _ := hex.DecodeString(
		nonce)

	reverseBytes(nonceBytes)

	// Concatenate all 6 fields to reconstruct the standard 80-byte Bitcoin block header
	header := bytes.Join([][]byte{
		version,
		prevHash,
		merkleRoot,
		timeBytes,
		bitsBytes,
		nonceBytes,
	}, []byte{})

	// Hash the reconstructed 80-byte block header
	hash := doubleSHA(header)

	hashForCompare := make([]byte, len(hash))
	copy(hashForCompare, hash)

	// Reverse the hash to Big-Endian format for numerical big.Int comparison
	reverseBytes(hashForCompare)

	hashInt := new(big.Int).SetBytes(
		hashForCompare)

	// Fetch the block target difficulty required by the Bitcoin Network
	target, err := CompactToTarget(
		job.Bits)
	if err != nil {
		return err
	}

	fmt.Println("share hash:",
		hex.EncodeToString(hashForCompare))

	// If the hash is greater than the network target, check if it fits the pool's lower share target
	if hashInt.Cmp(target) > 0 {

		// CRITICAL FIX REQUIRED: This pool currently rejects valid low-difficulty shares because it lacks a share target evaluation.
		fmt.Println("low difficulty share")

		return nil
	}

	// SUCCESS: The generated hash met the network target. A valid Bitcoin block has been found by a home miner!
	fmt.Println("BLOCK FOUND!")

	var txs []string

	for _, tx := range job.BlockTemplate.Transactions {
		txs = append(txs, tx.Data)
	}

	block := make([]byte, 0)

	// Build raw block payload: [80-byte header] + [tx_count varint] + [coinbase tx] + [mempool txs]
	block = append(block, header...)

	// CRITICAL FIX REQUIRED: CompactSize/VarInt formatting required for tx count instead of a raw single byte
	txCount := make([]byte, 1)
	txCount[0] = byte(len(txs) + 1)

	block = append(block, txCount...)

	block = append(block, coinbase...)

	for _, txHex := range txs {

		txBytes, err := hex.DecodeString(txHex)
		if err != nil {
			return err
		}

		block = append(block,
			txBytes...)
	}

	blockHex := hex.EncodeToString(block)

	fmt.Println("submitblock...")

	// Broadcast the newly assembled block to the local Bitcoin Core full node
	err = rpc.SubmitBlock(blockHex)
	if err != nil {
		return err
	}

	fmt.Println("BLOCK SUBMITTED!")

	return nil
}