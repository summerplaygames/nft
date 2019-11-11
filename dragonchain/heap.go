package dragonchain

import (
	"fmt"

	"github.com/dragonchain/dragonchain-sdk-go"
)

// Client is a client for interacting with the DragonChain API.
type Client interface {
	GetSmartContractObject(key, smartContractID string) (*dragonchain.Response, error)
}

// HeapFetcher fetches the smart contract heap using the DragonChain APIs.
type HeapFetcher struct {
	Client Client
}

// Heap returns the smart contract heap. An error is returned if the API request fails
// or returns a non-2xx status code.
func (h *HeapFetcher) Heap() ([]byte, error) {
	resp, err := h.Client.GetSmartContractObject("", "")
	if err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, fmt.Errorf("recieved invalid status code %d from dragonchain api", resp.Status)
	}
	return resp.Response.([]byte), nil
}
