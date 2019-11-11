package main

import (
	"fmt"
	"github.com/dragonchain/dragonchain-sdk-go"
	"net/http"
	"nft"
	dcheap "nft/dragonchain"
	"os"
)

var (
	dcID       string
	apiKey     string
	apiKeyID   string
	baseAPIURL string
	contractID string
)

func main() {
	dcClient, err := dragonClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create dragonchain client: %s\n", err)
		os.Exit(1)
	}
	rt := &nft.Runtime{
		HeapFetcher: &dcheap.HeapFetcher{
			Client: dcClient,
		},
		ContractFactory: &nft.DefaultContractFactory{},
		RPCHandler:      &rpcHandler{},
	}
	rt.Run()
}

func handleRPC(rpc []byte, contract nft.Contract) {
	// interpret RPC and handle with contract.
}

func dragonClient() (*dragonchain.Client, error) {
	httpClient := &http.Client{}
	creds, err := dragonchain.NewCredentials(dcID, apiKey, apiKeyID, dragonchain.HashSHA256)
	if err != nil {
		return nil, err
	}
	client := dragonchain.NewClient(creds, baseAPIURL, httpClient)
	return client, nil
}

type rpcHandler struct {
}

func (h *rpcHandler) HandleRPC(rpc []byte, contract nft.Contract) ([]byte, error) {
	return []byte{}, nil
}
