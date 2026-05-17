package stratum

import (
	"bufio"
	// "encoding/json"
	"fmt"
	"net"

	"brpool/internal/pool/types"

	"github.com/bytedance/sonic"
)

type Request struct {
	ID     interface{}   `json:"id"`
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
}

func Send(conn net.Conn, v interface{}) error {
	b, _ := sonic.Marshal(v)
	b = append(b, '\n')
	_, err := conn.Write(b)
	return err
}

func HandleConnection(
	conn net.Conn,
	job *types.Job,
	onSubmit func([]interface{}) error,
) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	extraNonce1 := "00000001"

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			return
		}

		var req Request

		if err := sonic.Unmarshal(line, &req); err != nil {
			continue
		}

		switch req.Method {

		case "mining.subscribe":

			Send(conn, map[string]interface{}{
				"id": req.ID,
				"result": []interface{}{
					[][]string{{"mining.notify", "1"}},
					extraNonce1,
					4,
				},
				"error": nil,
			})

			SendNotify(conn, job)

		case "mining.authorize":

			Send(conn, map[string]interface{}{
				"id":     req.ID,
				"result": true,
				"error":  nil,
			})

		case "mining.submit":

			err := onSubmit(req.Params)

			if err != nil {
				Send(conn, map[string]interface{}{
					"id":     req.ID,
					"result": false,
					"error":  err.Error(),
				})
			} else {
				Send(conn, map[string]interface{}{
					"id":     req.ID,
					"result": true,
					"error":  nil,
				})
			}
		}
	}
}

func SendNotify(conn net.Conn, job *types.Job) {
	msg := map[string]interface{}{
		"id": nil,
		"method": "mining.notify",
		"params": []interface{}{
			job.JobID,
			job.PrevHash,
			job.Coinbase1,
			job.Coinbase2,
			job.MerkleBranches,
			job.Version,
			job.Bits,
			job.NTime,
			job.CleanJobs,
		},
	}

	fmt.Println("sending notify")

	Send(conn, msg)
}