package types

import "net"

type Config struct {
	RPCHost     string `json:"rpc_host"`
	RPCPort     int    `json:"rpc_port"`
	RPCUser     string `json:"rpc_user"`
	RPCPass     string `json:"rpc_pass"`
	PoolAddress string `json:"pool_address"`
	StratumPort int    `json:"stratum_port"`
}

type BlockTemplate struct {
	Version           int      `json:"version"`
	PreviousBlockHash string   `json:"previousblockhash"`
	Bits              string   `json:"bits"`
	CurTime           uint32   `json:"curtime"`
	Height            int      `json:"height"`
	CoinbaseValue     uint64   `json:"coinbasevalue"`
	DefaultWitnessCommitment string `json:"default_witness_commitment"`

	Transactions []TxData `json:"transactions"`
}

type TxData struct {
	Data    string `json:"data"`
	TxID    string `json:"txid"`
	Hash    string `json:"hash"`
	Weight  int    `json:"weight"`
	Fee     int64  `json:"fee"`
	Depends []int  `json:"depends"`
}

type Job struct {
	JobID         string
	PrevHash      string
	Coinbase1     string
	Coinbase2     string
	MerkleBranches []string
	Version       string
	Bits          string
	NTime         string
	CleanJobs     bool

	BlockTemplate *BlockTemplate
}

type Miner struct {
	Conn        net.Conn
	ExtraNonce1 string
}