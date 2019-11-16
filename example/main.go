package main

import (
	"fmt"
	"net/http"
	"os"
	
	"github.com/dragonchain/dragonchain-sdk-go"
	"github.com/summerplaygames/nft"
	dcheap "github.com/summerplaygames/nft/dragonchain"
	
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
	rt := runtime(dcClient)
	rt.Run()
}

func handleRPC(rpc []byte, contract nft.Contract) {
	// interpret RPC and handle with contract.
}

func runtime(dcClient *dragonchain.Client) *nft.Runtime {
	heapFetcher := &dcheap.HeapFetcher{
		Client: dcClient,
	}
	contractFactory := &nft.DefaultContractFactory{}
	handler := &rpcHandler{}
	return nft.NewRuntime(heapFetcher, handler, contractFactory)
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
