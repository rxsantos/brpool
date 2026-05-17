package main

import (
	"fmt"
	"os"
	"sync/atomic"

	"github.com/bytedance/sonic"

	"brpool/internal/pool/bitcoin"
	"brpool/internal/pool/jobs"
	"brpool/internal/pool/shares"
	"brpool/internal/pool/stratum"
	"brpool/internal/pool/types"
)

// currentJob acts as the central, thread-safe global state holder for the active mining job.
// It uses sync/atomic.Pointer to guarantee lock-free, ultra-fast concurrent reads and writes 
// across all Stratum client routines and the ZMQ background listener.
var currentJob atomic.Pointer[types.Job]

// refreshJob pulls a new block template from the Bitcoin full node via RPC,
// transforms it into a Stratum-ready job structure, updates the global pointer atomic state,
// and broadcasts the new work payload to all connected micro-miners instantly.
func refreshJob(
	rpc *bitcoin.RPCClient,
) error {

	// Fetch a clean execution block template from Bitcoin Core (getblocktemplate RPC call)
	tmpl, err := rpc.GetBlockTemplate()
	if err != nil {
		return err
	}
	var cfg types.Config
	
	tmpl.PoolAddress = cfg.PoolAddress
	// Build the mining job skeleton from the newly fetched template
	job := jobs.CreateJob(tmpl)

	// Update the global atomic pointer safely without blocking active miner read operations
	currentJob.Store(job)

	fmt.Println("template height:",
		tmpl.Height)

	// Push the updated job to all active home miners to trigger immediate work switching
	stratum.BroadcastJob(job)

	return nil
}

// main acts as the master orchestrator, handling bootstrap routines, configuration parsing,
// background event loops initialization, and bringing up the low-latency Stratum network topology.
func main() {

	// 1. Bootstrap: Load environmental variables or system parameters from the JSON file
	cfgFile, err := os.ReadFile(
		"configs/config.json")
	if err != nil {
		panic(err)
	}

	var cfg types.Config

	// Parse JSON variables using the high-performance 'sonic' encoder library
	err = sonic.Unmarshal(cfgFile, &cfg)
	if err != nil {
		panic(err)
	}

	// 2. Client Setup: Instantiate the asynchronous RPC network engine for Bitcoin Core communication
	rpc := bitcoin.NewRPC(cfg)

	// 3. Priming: Force an initial job refresh to ensure a job exists in memory before clients connect
	err = refreshJob(rpc)
	if err != nil {
		panic(err)
	}

	// 4. Reactive Pipeline: Spawn a dedicated goroutine to catch incoming blocks from the network via ZeroMQ
	go func() {

		err := bitcoin.StartZMQBlockListener(
			cfg.ZMQBlock,
			func() {
				// Reactive Callback: Trigger a hot-reload of the block template when ZMQ fires a 'hashblock' event
				err := refreshJob(rpc)
				if err != nil {
					fmt.Println(err)
				}
			},
		)

		// Panic guard inside the background routine if the infrastructure notification channel drops completely
		if err != nil {
			// panic(err)
			fmt.Println(err)
		}
	}()

	// 5. Network Lifecycle: Open the Stratum TCP pipeline and inject state access closures
	err = stratum.StartServer(

		cfg.StratumPort,

		// State Access Closure: Passes the latest lock-free reference of the active block target
		func() *types.Job {
			return currentJob.Load()
		},

		// Submission Handler Closure: Routes incoming miner shares straight to the cryptographic engine
		func(params []interface{}) error {

			job := currentJob.Load()
			if job == nil {
				return nil
			}

			return shares.ValidateShare(
				params,
				job,
				rpc,
			)
		},
	)

	if err != nil {
		panic(err)
	}
}