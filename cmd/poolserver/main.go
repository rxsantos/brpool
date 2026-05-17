package main

import (
	// "encoding/json"
	"fmt"
	"os"

	"brpool/internal/pool/bitcoin"
	"brpool/internal/pool/jobs"
	"brpool/internal/pool/shares"
	"brpool/internal/pool/stratum"
	"brpool/internal/pool/types"

	"github.com/bytedance/sonic"
)

func main() {

	cfgFile, err := os.ReadFile("configs/config.json")
	if err != nil {
		panic(err)
	}

	var cfg types.Config

	if err := sonic.Unmarshal(cfgFile, &cfg); err != nil {
		panic(err)
	}

	rpc := bitcoin.NewRPC(cfg)

	tmpl, err := rpc.GetBlockTemplate()
	if err != nil {
		panic(err)
	}

	fmt.Println("template height:", tmpl.Height)

	job := jobs.CreateJob(tmpl)

	err = stratum.StartServer(
		cfg.StratumPort,
		job,
		func(params []interface{}) error {
			return shares.ValidateShare(params, job, rpc)
		},
	)

	if err != nil {
		panic(err)
	}
}