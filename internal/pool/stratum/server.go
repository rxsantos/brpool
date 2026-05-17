package stratum

import (
	"fmt"
	"net"
	"sync"

	"brpool/internal/pool/types"
)

// miners is a thread-safe global map using sync.Map to track all active Stratum TCP connections.
// The key is the miner's remote address string, and the value is the net.Conn object.
var miners sync.Map

// BroadcastJob loops through all currently connected miners and pushes a new mining job
// concurrently. This is triggered instantly whenever a new block is detected on the Bitcoin network.
func BroadcastJob(job *types.Job) {

	miners.Range(func(_, value interface{}) bool {

		// Type assert the generic interface value back to a net.Conn network socket
		conn := value.(net.Conn)

		// Fire a separate goroutine per miner to prevent a slow network connection 
		// from blocking notifications to other miners (ultra-low latency distribution).
		go SendNotify(conn, job)

		return true // Continue iterating through the rest of the map
	})
}

// StartServer initializes the Stratum TCP server on the configured port.
// It listens for incoming mining device connections (e.g., NerdMiner, NerdAxe) and spawns handler routines.
func StartServer(
	port int,
	jobManager func() *types.Job, // Callback function to fetch the current active mining job
	onSubmit func([]interface{}) error, // Callback function triggered when a miner submits a share
) error {

	// Bind and listen to the specified local TCP port
	ln, err := net.Listen("tcp",
		fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	fmt.Println("stratum listening on", port)

	// Infinite loop to accept incoming client connections
	for {

		// Block until a new miner establishes a TCP handshake
		conn, err := ln.Accept()
		if err != nil {
			continue // If connection fails mid-handshake, ignore and wait for the next one
		}

		// Store the connection socket in the concurrent map using its Remote IP:Port as the unique key
		// miners.Store(conn.RemoteAddr().String(),
		// 	conn)
		clientAddr := conn.RemoteAddr().String()
		miners.Store(clientAddr, conn)

		// Spawn a dedicated goroutine to handle the entire session of this specific miner concurrently
		go func() {

			// Ensure the miner is removed from the active map when the connection drops or closes
			// defer miners.Delete(
			// 	conn.RemoteAddr().String())
			defer miners.Delete(clientAddr)
    		defer conn.Close()

			// Process the Stratum RPC protocol lifecycle for this connection
			HandleConnection(
				conn,
				jobManager(), // Pass a fresh copy of the current block job
				onSubmit,
			)
		}()
	}
}