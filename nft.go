package nft

import (
	"encoding/json"
	"io"
	"math/big"
)

var (
	// BigZero is the big.Int with a value of 0.
	BigZero = big.NewInt(0)

	bigOne = big.NewInt(1)
)

// Context is the NFT context, holding information about who owns what tokens, etc...
type Context struct {
	TokenOwners     map[string]string                 `json:"tokenOwners"`
	OwnedTokens     map[string][]string               `json:"ownedTokens"`
	OwnedTokenIndex map[string]uint64                 `json:"ownedTokenIndex"`
	TotalTokens     string                            `json:"totalTokens"`
	TokenMetadata   map[string]map[string]interface{} `json:"tokenMetadata"`
}

// Contract is an NFT smart contract implementation that is designed to work with
// the DragonChain platform.
type Contract struct {
	name    string
	symbol  string
	context *Context
}

// Opt is a function that can be used provide optional configuration to a Contract.
type Opt func(*Contract)

// WithContext returns an Opt that sets the Contract's initial Context to ctx.
func WithContext(ctx *Context) Opt {
	return func(c *Contract) {
		c.context = ctx
	}
}

// NewContract returns a new contract instance. The provided name and symbol are used to
// distinguish between different NFT contracts on the same node. A variable number of Opts
// can be provided to configure the initial state of the Contract.
func NewContract(name, symbol string, opts ...Opt) *Contract {
	contract := &Contract{
		name:    name,
		symbol:  symbol,
		context: &Context{},
	}
	for _, opt := range opts {
		opt(contract)
	}
	return contract
}

// BalanceOf returns the current number of NFTs owned by owner.
func (c *Contract) BalanceOf(owner string) uint64 {
	return uint64(len(c.context.OwnedTokens[owner]))
}

// OwnerOf returns the address of the current owner of a token.
// If the owner of the current token is unknown, the second boolean return
// argument will be false.
func (c *Contract) OwnerOf(tokenID *big.Int) (string, bool) {
	owner, ok := c.context.TokenOwners[tokenID.String()]
	return owner, ok
}

// Transfer transfers the token with the given id from the "from" address to the "to" address.
func (c *Contract) Transfer(from, to string, tokenID *big.Int) {
	tokenIndex, ok := c.context.OwnedTokenIndex[tokenID.String()]
	if !ok {
		return
	}
	totalTokens, ok := BigIntString(c.context.TotalTokens)
	if !ok {
		return
	}
	tid := tokenID.String()
	// remove token from "from" address
	delete(c.context.TokenOwners, tid)
	c.context.OwnedTokens[from] = append(c.context.OwnedTokens[from][:tokenIndex], c.context.OwnedTokens[from][tokenIndex+1:]...)
	delete(c.context.OwnedTokenIndex, tid)

	// add token to "to" address
	c.context.TokenOwners[tid] = to
	balance := c.BalanceOf(to)
	c.context.OwnedTokens[to] = append(c.context.OwnedTokens[to], tid)
	c.context.OwnedTokenIndex[tid] = balance
	c.context.TotalTokens = totalTokens.Add(totalTokens, bigOne).String()
}

// TotalSupply returns the current known supply of the token. This supply is updated
// every time a new token is minted.
func (c *Contract) TotalSupply() *big.Int {
	if totalSupply, ok := BigIntString(c.context.TotalTokens); ok {
		return totalSupply
	}
	return BigZero
}

// TokenOfOwnerByIndex returns the token id for a given index into int token owner's list of tokens.
// If either the owner is unknown, or the requested index is out of range, then the second boolean return
// argument will be false.
func (c *Contract) TokenOfOwnerByIndex(owner string, idx uint64) (*big.Int, bool) {
	if tokens, ok := c.context.OwnedTokens[owner]; ok {
		if idx < uint64(len(tokens)) {
			bi, ok := BigIntString(tokens[idx])
			return bi, ok
		}
	}
	return nil, false
}

// Name returns the name of the NFT contract.
func (c *Contract) Name() string {
	return c.name
}

// Symbol returns the Symbol of hte NFT contract.
func (c *Contract) Symbol() string {
	return c.symbol
}

// MetadataFor returns the metadata associated with a token. If no metadata for the token exists,
// an new map will be created and assigned to the token. This guarantees that the caller will always
// have a valid map to work with.
func (c *Contract) MetadataFor(tokenID *big.Int) map[string]interface{} {
	if meta, ok := c.context.TokenMetadata[tokenID.String()]; ok {
		return meta
	}
	meta := make(map[string]interface{})
	c.context.TokenMetadata[tokenID.String()] = meta
	return meta
}

// EncodeTo encodes the current state of the Contract (its Context) and writes it to a Writer. An error
// is returned if the state could not be written.
func (c *Contract) EncodeTo(w io.Writer) error {
	enc := json.NewEncoder(w)
	return enc.Encode(c.context)
}

// BigIntString is a convenience function for creating a big.Int from string. The string is assumed to be
// a base 10 number. If the big.Int could not be created from the provided string, the second boolean return
// argument will be false.
func BigIntString(s string) (*big.Int, bool) {
	bi := &big.Int{}
	bi, ok := bi.SetString(s, 10)
	return bi, ok
}
