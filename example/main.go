package main

import (
	"github.com/summerplaygames/nft"
)

var (
	dcID       string
	apiKey     string
	apiKeyID   string
	baseAPIURL string
	contractID string
)

func main() {
	contractFactory := &nft.DefaultContractFactory{}
	rt := nft.NewRuntime(handleRPC(), contractFactory)
	rt.Run()
}

func handleRPC() nft.RPCHandlerFunc {
	return func(rpc []byte, contract nft.Contract) (interface{}, error) {
		return contract, nil
	}
}
