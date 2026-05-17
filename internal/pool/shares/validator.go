package shares

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"brpool/internal/pool/bitcoin"
	"brpool/internal/pool/types"
)

func doubleSHA(data []byte) []byte {
	h1 := sha256.Sum256(data)
	h2 := sha256.Sum256(h1[:])
	return h2[:]
}

func ValidateShare(
	params []interface{},
	job *types.Job,
	rpc *bitcoin.RPCClient,
) error {

	if len(params) < 5 {
		return fmt.Errorf("invalid submit")
	}

	extraNonce2 := params[2].(string)
	ntime := params[3].(string)
	nonce := params[4].(string)

	fmt.Println("share received")
	fmt.Println("extranonce2:", extraNonce2)
	fmt.Println("ntime:", ntime)
	fmt.Println("nonce:", nonce)

	headerHex := job.Version +
		job.PrevHash +
		job.Bits +
		ntime +
		nonce

	header, err := hex.DecodeString(headerHex)
	if err != nil {
		return err
	}

	hash := doubleSHA(header)

	fmt.Println("hash:", hex.EncodeToString(hash))

	return nil
}