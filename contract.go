package nft

import (
	"math/big"
)

var (
	// BigZero is a math/big.Int with a value of 0.
	BigZero = big.NewInt(0)

	bigOne = big.NewInt(1)
)

// Contract is an NFT smart contract implementation that is designed to work with
// the DragonChain platform.
type Contract struct {
	Name            string                            `json:"name"`
	Symbol          string                            `json:"symbol"`
	TokenOwners     map[string]string                 `json:"tokenOwners"`
	OwnedTokens     map[string][]string               `json:"ownedTokens"`
	OwnedTokenIndex map[string]uint64                 `json:"ownedTokenIndex"`
	TotalTokens     string                            `json:"totalTokens"`
	TokenMetadata   map[string]map[string]interface{} `json:"tokenMetadata"`
}

// BalanceOf returns the current number of NFTs owned by owner.
func (c *Contract) BalanceOf(owner string) uint64 {
	if c.OwnedTokens == nil {
		return 0
	}
	return uint64(len(c.OwnedTokens[owner]))
}

// OwnerOf returns the address of the current owner of a token.
// If the owner of the current token is unknown, the second boolean return
// argument will be false.
func (c *Contract) OwnerOf(tokenID *big.Int) (string, bool) {
	if c.TokenOwners == nil {
		return "", false
	}
	owner, ok := c.TokenOwners[tokenID.String()]
	return owner, ok
}

// Transfer transfers the token with the given id from the "from" address to the "to" address.
func (c *Contract) Transfer(from, to string, tokenID *big.Int) {
	if c.OwnedTokens == nil || c.OwnedTokenIndex == nil || c.TokenOwners == nil {
		return
	}
	tokenIndex, ok := c.OwnedTokenIndex[tokenID.String()]
	if !ok {
		return
	}
	totalTokens, ok := BigIntString(c.TotalTokens)
	if !ok {
		return
	}
	tid := tokenID.String()
	// remove token from "from" address
	delete(c.TokenOwners, tid)
	c.OwnedTokens[from] = append(c.OwnedTokens[from][:tokenIndex], c.OwnedTokens[from][tokenIndex+1:]...)
	delete(c.OwnedTokenIndex, tid)

	// add token to "to" address
	c.TokenOwners[tid] = to
	balance := c.BalanceOf(to)
	c.OwnedTokens[to] = append(c.OwnedTokens[to], tid)
	c.OwnedTokenIndex[tid] = balance
	c.TotalTokens = totalTokens.Add(totalTokens, bigOne).String()
}

// TotalSupply returns the current known supply of the token. This supply is updated
// every time a new token is minted.
func (c *Contract) TotalSupply() *big.Int {
	if totalSupply, ok := BigIntString(c.TotalTokens); ok {
		return totalSupply
	}
	return BigZero
}

// TokenOfOwnerByIndex returns the token id for a given index into int token owner's list of tokens.
// If either the owner is unknown, or the requested index is out of range, then the second boolean return
// argument will be false.
func (c *Contract) TokenOfOwnerByIndex(owner string, idx uint64) (*big.Int, bool) {
	if c.OwnedTokens == nil {
		return nil, false
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
func (c *Contract) MetadataFor(tokenID *big.Int) map[string]interface{} {
	if meta, ok := c.TokenMetadata[tokenID.String()]; ok {
		return meta
	}
	meta := make(map[string]interface{})
	c.TokenMetadata[tokenID.String()] = meta
	return meta
}

// BigIntString is a convenience function for creating a big.Int from string. The string is assumed to be
// a base 10 number. If the big.Int could not be created from the provided string, the second boolean return
// argument will be false.
func BigIntString(s string) (*big.Int, bool) {
	bi := &big.Int{}
	bi, ok := bi.SetString(s, 10)
	return bi, ok
}
