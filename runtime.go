// +build !test

package nft

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// RPCHandlerFunc is a convenience type that allows for using a function in place
// of an RPCHandler.
type RPCHandlerFunc func(input []byte, contract Contract) (interface{}, error)

// HandleRPC exists to satisfy the RPCHandler interface. It is a strait pass-through to
// the underlying function.
func (f RPCHandlerFunc) HandleRPC(input []byte, contract Contract) (interface{}, error) {
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
	// The returned object will be json serialized and written to stdout. As such, will be
	// stored on the heap, as per the usual DragonChain smart contract heap semantics.
	//
	// An optional error can be returned to signify that the handling of the RPC failed.
	// In this case, nothing will be written to the heap, and the error will be logged to stderr.
	HandleRPC(input []byte, contract Contract) (interface{}, error)
}

// ContractFactory creates a new Contract from some input.
type ContractFactory interface {
	CreateContract(name, symbol string) (Contract, error)
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
	name, symbol := os.Getenv("CONTRACT_NAME"), os.Getenv("CONTRACT_SYMBOL")
	if name == "" {
		fmt.Fprintln(os.Stderr, "no name provided for contract")
		os.Exit(1)
	}
	if symbol == "" {
		fmt.Fprintln(os.Stderr, "no symbol provided for contract")
		os.Exit(1)
	}
	contract, err := r.contractFactory.CreateContract(name, symbol)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create contract: %s\n", err)
		os.Exit(1)
	}
	b, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read stdin: %s\n", err)
		os.Exit(1)
	}
	obj, err := r.rpcHandler.HandleRPC(b, contract)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to handle RPC: %s\n", err)
		os.Exit(1)
	}
	if err = json.NewEncoder(os.Stdout).Encode(obj); err != nil {
		fmt.Fprintf(os.Stderr, "failed to JSON encode heap output: %s\n", err)
		os.Exit(1)
	}
}
