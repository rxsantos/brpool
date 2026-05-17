package bitcoin

import (
	"context"
	"fmt"

	"github.com/go-zeromq/zmq4"
)

// StartZMQBlockListener connects to the Bitcoin Core ZMQ notification interface
// and listens for the 'hashblock' event to detect new blocks instantly.
// When a block is found on the network, it triggers the provided 'onBlock' callback.
func StartZMQBlockListener(
	address string,
	onBlock func(),
) error {

	ctx := context.Background()

	// Initialize a ZeroMQ Subscriber (SUB) socket
	sub := zmq4.NewSub(ctx)

	// Dial and connect to the Bitcoin Core ZMQ publishing address (e.g., tcp://127.0.0.1:28332)
	err := sub.Dial(address)
	if err != nil {
		return err
	}

	// Subscribe exclusively to the "hashblock" topic to avoid receiving unrelated payloads
	err = sub.SetOption(zmq4.OptionSubscribe,
		"hashblock")
	if err != nil {
		return err
	}

	fmt.Println("ZMQ connected:", address)

	// Infinite event loop to block and wait for incoming ZeroMQ multipart messages
	for {

		msg, err := sub.Recv()
		if err != nil {
			// If a network read or frame parsing error occurs, skip and try again
			continue
		}

		// Bitcoin ZMQ 'hashblock' multipart messages must contain at least 2 frames:
		// Frame 0: Topic name ("hashblock")
		// Frame 1: The 32-byte raw binary block hash
		if len(msg.Frames) < 2 {
			continue
		}

		topic := string(msg.Frames[0])

		// Double-check the topic name to ensure it matches our subscription
		if topic != "hashblock" {
			continue
		}

		fmt.Println("new block detected")

		// Execute the callback function to notify the pool to fetch a new block template
		onBlock()
	}
}