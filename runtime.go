package nft

import (
	"fmt"
	"io/ioutil"
	"os"
)

// HeapFetcher fetches the contract Heap.
type HeapFetcher interface {
	Heap() ([]byte, error)
}

// RPCHandler handles RPCs from clients.
type RPCHandler interface {
	HandleRPC(input []byte, contract Contract) ([]byte, error)
}

// ContractFactory creates a new Contract from a heap.
type ContractFactory interface {
	CreateContract(heap []byte) (Contract, error)
}

// Runtime is used to run and NFT contract.
type Runtime struct {
	heapFetcher     HeapFetcher
	rpcHandler      RPCHandler
	contractFactory ContractFactory
}

// NewRuntime returns a new Runtime instance.
func NewRuntime(heapFetcher HeapFetcher, rpcHandler RPCHandler, contractFactory ContractFactory) *Runtime {
	return &Runtime{
		heapFetcher:     heapFetcher,
		rpcHandler:      rpcHandler,
		contractFactory: contractFactory,
	}
}

// Run fetches the contract heap, creates a new contract, and
// then uses that contract to handle the input RPC.
func (r *Runtime) Run() {
	heap, err := r.heapFetcher.Heap()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to fetch heap: %s", err)
		os.Exit(1)
	}
	contract, err := r.contractFactory.CreateContract(heap)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create contract: %s", err)
		os.Exit(1)
	}
	b, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read stdin: %s", err)
		os.Exit(1)
	}
	b, err = r.rpcHandler.HandleRPC(b, contract)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to handle RPC: %s", err)
		os.Exit(1)
	}
	fmt.Print(b)
}
