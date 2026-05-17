package types

import "net"

// Config holds the application configuration loaded from the configuration file (e.g., config.json).
// It contains credentials for the Bitcoin node RPC and localized server parameters.
type Config struct {
	RPCHost     string `json:"rpc_host"`     // Bitcoin Core RPC host address
	RPCPort     int    `json:"rpc_port"`     // Bitcoin Core RPC port number
	RPCUser     string `json:"rpc_user"`     // Username for Bitcoin Core RPC authentication
	RPCPass     string `json:"rpc_pass"`     // Password for Bitcoin Core RPC authentication
	PoolAddress string `json:"pool_address"` // On-chain Bitcoin address where block rewards (coinbase) are sent
	StratumPort int    `json:"stratum_port"` // Local port for the Stratum server to accept miner connections
	ZMQBlock 	string `json:"zmq_block"`	 // ZMQ notification interface and listens for the 'hashblock' event to detect new blocks instantly	
}

// TxData represents raw transaction data received from the block template.
// These are the transactions waiting in the mempool that the pool will pack into the next block.
type TxData struct {
	Data    string `json:"data"`    // Serialized hex-encoded transaction data
	TxID    string `json:"txid"`    // Transaction ID (txid)
	Hash    string `json:"hash"`    // Witness transaction hash (wtxid)
	Weight  int    `json:"weight"`  // Transaction weight (segwit standard measure)
	Fee     int64  `json:"fee"`     // Transaction fee in satoshis
	Depends []int  `json:"depends"` // 1-based indexes of transactions in the template that this transaction depends on
}

// BlockTemplate maps the JSON response structure of the 'getblocktemplate' RPC call from Bitcoin Core (BIP 22).
// It represents the raw structure of a new block candidate waiting to be mined.
type BlockTemplate struct {
	Version                  int      `json:"version"`                    // Block version number
	PreviousBlockHash        string   `json:"previousblockhash"`          // Hex string of the previous block's hash
	Bits                     string   `json:"bits"`                       // Encoded target difficulty for the block
	CurTime                  uint32   `json:"curtime"`                    // Current timestamp in seconds since epoch
	Height                   int      `json:"height"`                     // The block height in the blockchain
	CoinbaseValue            uint64   `json:"coinbasevalue"`              // Total block reward (subsidy + transaction fees) in satoshis
	DefaultWitnessCommitment string   `json:"default_witness_commitment"` // SegWit commitment hash inserted in the coinbase transaction

	PoolAddress string

	Transactions []TxData `json:"transactions"` // List of transactions included in the template
}

// Job represents a unique mining job generated from a BlockTemplate.
// This data is serialized and split according to the Stratum protocol to be sent to the miners.
type Job struct {
	JobID string // Unique identifier for the mining job

	PrevHash string // Hash of the previous block, reversed and formatted for Stratum miners

	Coinbase1 string // Part 1 of the coinbase transaction (prefix) before the ExtraNonce fields
	Coinbase2 string // Part 2 of the coinbase transaction (suffix) after the ExtraNonce fields

	MerkleBranches []string // Merkle tree steps used by micro-miners to build the block header hash

	Version string // Block version formatted as a hex string for the Stratum protocol
	Bits    string // Difficulty target formatted as a hex string
	NTime   string // Block timestamp formatted as a hex string

	CleanJobs bool // Tells the miner whether to drop current working shares and switch to this job instantly

	BlockTemplate *BlockTemplate // Reference to the original full block template used to build this job
}

// Miner represents an active connection from a local mining device (e.g., NerdMiner, NerdAxe).
// It manages the network socket and tracking parameters required for the Stratum session.
type Miner struct {
	Conn        net.Conn // TCP network connection to the miner
	ExtraNonce1 string   // Unique hex string assigned by the pool to prevent hash collisions between miners
}

type Session struct {
	Miner       *Miner
	ExtraNonce1 string
}