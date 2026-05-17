package jobs

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"brpool/internal/pool/types"
)

func RandomJobID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func CreateJob(tmpl *types.BlockTemplate) *types.Job {
	return &types.Job{
		JobID:         RandomJobID(),
		PrevHash:      tmpl.PreviousBlockHash,
		Coinbase1:     "01000000",
		Coinbase2:     "ffffffff",
		MerkleBranches: []string{},
		Version:       fmt.Sprintf("%08x", tmpl.Version),
		Bits:          tmpl.Bits,
		NTime:         fmt.Sprintf("%08x", tmpl.CurTime),
		CleanJobs:     true,
	}
}