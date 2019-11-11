package nft

import (
	"encoding/json"
	"fmt"
	"math/big"
)

var (
	// BigZero is a math/big.Int with a value of 0.
	BigZero = big.NewInt(0)
	// BigNegOne is a math/big.In with a value of -1.
	BigNegOne = big.NewInt(-1)

	bigOne = big.NewInt(1)
)

// Contract is a DCRC1-compatible smart contract.
type Contract interface {
	BalanceOf(owner string) uint64
	OwnerOf(tokenID *big.Int) (string, bool)
	Mint(to string)
	Burn(tokenID *big.Int)
	Transfer(from, to string, tokenID *big.Int)
	TotalSupply() *big.Int
	TokenOfOwnerByIndex(owner string, idx uint64) (*big.Int, bool)
	MetadataFor(tokenID *big.Int) map[string]interface{}
}

// DefaultContract is a basic NFT smart contract implementation that is designed to work with
// the DragonChain platform.
type DefaultContract struct {
	Name            string                            `json:"name"`
	Symbol          string                            `json:"symbol"`
	TokenOwners     map[string]string                 `json:"tokenOwners"`
	OwnedTokens     map[string][]string               `json:"ownedTokens"`
	OwnedTokenIndex map[string]uint64                 `json:"ownedTokenIndex"`
	TotalTokens     string                            `json:"totalTokens"`
	TokenMetadata   map[string]map[string]interface{} `json:"tokenMetadata"`
}

// BalanceOf returns the current number of NFTs owned by owner.
func (c *DefaultContract) BalanceOf(owner string) uint64 {
	if c.OwnedTokens == nil {
		c.OwnedTokens = make(map[string][]string)
	}
	return uint64(len(c.OwnedTokens[owner]))
}

// OwnerOf returns the address of the current owner of a token.
// If the owner of the current token is unknown, the second boolean return
// argument will be false.
func (c *DefaultContract) OwnerOf(tokenID *big.Int) (string, bool) {
	if c.TokenOwners == nil {
		c.TokenOwners = make(map[string]string)
	}
	owner, ok := c.TokenOwners[tokenID.String()]
	return owner, ok
}

// Mint mints a new token and assigns it to the "to" address.
func (c *DefaultContract) Mint(to string) {
	total := c.TotalSupply()
	if total != BigNegOne {
		c.addToken(to, total.Add(total, bigOne).String())
	}
}

// Burn destroys a token and removes it from its owner.
func (c *DefaultContract) Burn(tokenID *big.Int) {
	if c.TokenOwners == nil {
		return
	}
	tid := tokenID.String()
	if owner, ok := c.TokenOwners[tid]; ok {
		c.removeToken(owner, tid)
	}
}

// Transfer transfers the token with the given id from the "from" address to the "to" address.
func (c *DefaultContract) Transfer(from, to string, tokenID *big.Int) {
	tid := tokenID.String()
	c.removeToken(from, tid)
	c.addToken(to, tid)
}

// TotalSupply returns the current known supply of the token. This supply is updated
// every time a new token is minted.
func (c *DefaultContract) TotalSupply() *big.Int {
	if totalSupply, ok := BigIntString(c.TotalTokens); ok {
		return totalSupply
	}
	return BigNegOne
}

// TokenOfOwnerByIndex returns the token id for a given index into int token owner's list of tokens.
// If either the owner is unknown, or the requested index is out of range, then the second boolean return
// argument will be false.
func (c *DefaultContract) TokenOfOwnerByIndex(owner string, idx uint64) (*big.Int, bool) {
	if c.OwnedTokens == nil {
		c.OwnedTokens = make(map[string][]string)
	}
	if tokens, ok := c.OwnedTokens[owner]; ok {
		if idx < uint64(len(tokens)) {
			bi, ok := BigIntString(tokens[idx])
			return bi, ok
		}
	}
	return nil, false
}

// MetadataFor returns the metadata associated with a token. If no metadata for the token exists,
// an new map will be created and assigned to the token. This guarantees that the caller will always
// have a valid map to work with.
func (c *DefaultContract) MetadataFor(tokenID *big.Int) map[string]interface{} {
	if meta, ok := c.TokenMetadata[tokenID.String()]; ok {
		return meta
	}
	meta := make(map[string]interface{})
	c.TokenMetadata[tokenID.String()] = meta
	return meta
}

func (c *DefaultContract) removeToken(from, tid string) {
	if c.OwnedTokenIndex == nil {
		c.OwnedTokenIndex = make(map[string]uint64)
	}
	if c.TokenOwners == nil {
		c.TokenOwners = make(map[string]string)
	}
	if c.OwnedTokens == nil {
		c.OwnedTokens = make(map[string][]string)
	}
	totalTokens, ok := BigIntString(c.TotalTokens)
	if !ok {
		return
	}
	if tokenIndex, ok := c.OwnedTokenIndex[tid]; ok {
		// remove token from "from" address
		delete(c.TokenOwners, tid)
		c.OwnedTokens[from] = append(c.OwnedTokens[from][:tokenIndex], c.OwnedTokens[from][tokenIndex+1:]...)
		delete(c.OwnedTokenIndex, tid)
		c.TotalTokens = totalTokens.Sub(totalTokens, bigOne).String()
	}
}

func (c *DefaultContract) addToken(to, tid string) {
	if c.TokenOwners == nil {
		c.TokenOwners = make(map[string]string)
	}
	if c.OwnedTokens == nil {
		c.OwnedTokens = make(map[string][]string)
	}
	if c.OwnedTokenIndex == nil {
		c.OwnedTokenIndex = make(map[string]uint64)
	}
	if totalTokens, ok := BigIntString(c.TotalTokens); ok {
		// add token to "to" address
		c.TokenOwners[tid] = to
		balance := c.BalanceOf(to)
		c.OwnedTokens[to] = append(c.OwnedTokens[to], tid)
		c.OwnedTokenIndex[tid] = balance
		c.TotalTokens = totalTokens.Add(totalTokens, bigOne).String()
	}
}

// BigIntString is a convenience function for creating a big.Int from string. The string is assumed to be
// a base 10 number. If the big.Int could not be created from the provided string, the second boolean return
// argument will be false.
func BigIntString(s string) (*big.Int, bool) {
	if s == "" {
		s = "0"
	}
	bi := &big.Int{}
	bi, ok := bi.SetString(s, 10)
	return bi, ok
}

// DefaultContractFactory creates a new DefaultContract from the heap.
type DefaultContractFactory struct{}

// CreateContract retunrs a new DefaultContract from heap. Heap must be the json
// encoded contract state. If not, an error is returned.
func (f *DefaultContractFactory) CreateContract(heap []byte) (Contract, error) {
	var contract DefaultContract
	if err := json.Unmarshal(heap, &contract); err != nil {
		return nil, fmt.Errorf("failed to JSON unmarshal heap: %s", err)
	}
	return &contract, nil
}
