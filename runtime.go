package nft

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/dragonchain/dragonchain-sdk-go"
)

// HeapKeyTokens is used as the key for which to store token data on the DragonChain heap.
const HeapKeyTokens = "tokens"

// RPCHandlerFunc is a convenience type that allows for using a function in place
// of an RPCHandler.
type RPCHandlerFunc func(input []byte, contract Contract) ([]byte, error)

// HandleRPC exists to satisfy the RPCHandler interface. It is a strait pass-through to
// the underlying function.
func (f RPCHandlerFunc) HandleRPC(input []byte, contract Contract) ([]byte, error) {
	return f(input, contract)
}

// RPCHandler handles RPCs from clients.
type RPCHandler interface {
	// HandleRPC handles the provided input by performing actions on a Contract.
	// The concrete type of the provided contract can be discerned via interface type
	// assertions. For example:
	//   if concrete, ok := contract.(ConcreteType); ok {
	//	     // Do something...
	//   }
	// The returned byte array will be written to stdout, and as such, will be stored on the heap,
	// as per the usual DragonChain smart contract heap semantics.
	//
	// An optional error can be returned to signify that the handling of the RPC failed.
	// In this case, nothing will be written to the heap, and the error will be logged to stderr.
	HandleRPC(input []byte, contract Contract) ([]byte, error)
}

// ContractFactory creates a new Contract from some input.
type ContractFactory interface {
	CreateContract() (Contract, error)
}

// Runtime is used to run and NFT contract.
type Runtime struct {
	rpcHandler      RPCHandler
	contractFactory ContractFactory
}

// NewRuntime returns a new Runtime instance from the provided RPCHandler and ContractFactory.
func NewRuntime(rpcHandler RPCHandler, contractFactory ContractFactory) *Runtime {
	return &Runtime{
		rpcHandler:      rpcHandler,
		contractFactory: contractFactory,
	}
}

// Run fetches the contract heap, creates a new contract, and
// then uses that contract to handle the input RPC.
func (r *Runtime) Run() {
	contract, err := r.contractFactory.CreateContract()
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

// Load returns the smart contract heap. An error is returned if the API request fails
// or returns a non-2xx status code.
func loadHeap(client *dragonchain.Client) ([]byte, error) {
	resp, err := client.GetSmartContractObject(HeapKeyTokens, "")
	if err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, fmt.Errorf("recieved invalid status code %d from dragonchain api", resp.Status)
	}
	return resp.Response.([]byte), nil
}
