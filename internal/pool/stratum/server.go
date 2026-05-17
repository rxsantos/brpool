package stratum

import (
	"fmt"
	"net"

	"brpool/internal/pool/types"
)

func StartServer(
	port int,
	job *types.Job,
	onSubmit func([]interface{}) error,
) error {

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	fmt.Println("stratum listening on", port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}

		go HandleConnection(conn, job, onSubmit)
	}
}