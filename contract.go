package nft

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"os"

	"github.com/dragonchain/dragonchain-sdk-go"
)

var (
	// BigZero is a math/big.Int with a value of 0.
	BigZero = big.NewInt(0)

	bigOne = big.NewInt(1)

	// ErrNoExist is returned when a requested resource does not exist.
	ErrNoExist = errors.New("resource does not exist")
	// ErrAlreadyExists is returned when a resource already exists and it shouldn't.
	ErrAlreadyExists = errors.New("resource already exists")
	// ErrInvalidBigIntString is returned when a String cannot be converted to a big.Int
	ErrInvalidBigIntString = errors.New("big.Int invalid")
)

// Client is a client for interacting with the DragonChain API.
type Client interface {
	GetSmartContractObject(key, smartContractID string) (*dragonchain.Response, error)
}

// Contract is a DCRC1-compatible smart contract.
type Contract interface {
	Name() string
	Symbol() string
	BalanceOf(owner string) (uint64, error)
	OwnerOf(tokenID string) (string, error)
	Mint(to, tokenID string) error
	Burn(tokenID string) error
	Transfer(from, to, tokenID string) error
	TotalSupply() (*big.Int, error)
	TokensOwnedBy(owner string) ([]string, error)
}

// DefaultContract is a basic NFT smart contract implementation that is designed to work with
// the DragonChain platform.
type DefaultContract struct {
	TokenOwners     map[string]string   `json:"tokenOwners,omitempty"`
	OwnedTokens     map[string][]string `json:"ownedTokens,omitempty"`
	OwnedTokenIndex map[string]uint64   `json:"ownedTokenIndex,omitempty"`
	TotalTokens     string              `json:"totalTokens,omitempty"`

	ContractName   string `json:"name"`
	ContractSymbol string `json:"symbol"`

	client Client
}

// NewDefaultContract returns a DefaultContract that uses the provided DragonChain client.
func NewDefaultContract(name, symbol string, client Client) *DefaultContract {
	return &DefaultContract{
		ContractName:   name,
		ContractSymbol: symbol,
		client:         client,
	}
}

// Name returns the name of the Contract.
func (c *DefaultContract) Name() string {
	return c.ContractName
}

// Symbol returns the Contract's symbol.
func (c *DefaultContract) Symbol() string {
	return c.ContractSymbol
}

// BalanceOf returns the current number of NFTs owned by owner.
func (c *DefaultContract) BalanceOf(owner string) (uint64, error) {
	tokens, err := c.TokensOwnedBy(owner)
	return uint64(len(tokens)), err
}

// OwnerOf returns the address of the current owner of a token.
func (c *DefaultContract) OwnerOf(tokenID string) (string, error) {
	if c.TokenOwners == nil {
		if err := c.fetchTokenOwners(); err != nil {
			return "", err
		}
	}
	if owner, ok := c.TokenOwners[tokenID]; ok {
		return owner, nil
	}
	return "", ErrNoExist
}

// Mint mints a new token with the provided ID and assigns it to the "to" address.
func (c *DefaultContract) Mint(to, tokenID string) error {
	return c.addToken(to, tokenID)
}

// Burn destroys a token and removes it from its owner.
func (c *DefaultContract) Burn(tokenID string) error {
	owner, err := c.OwnerOf(tokenID)
	if err != nil {
		return err
	}
	return c.removeToken(owner, tokenID)
}

// Transfer transfers the token with the given id from the "from" address to the "to" address.
func (c *DefaultContract) Transfer(from, to, tokenID string) error {
	if c.TokenOwners == nil {
		if err := c.fetchTokenOwners(); err != nil {
			return err
		}
	}
	if c.OwnedTokens == nil {
		if err := c.fetchOwnedTokens(); err != nil {
			return err
		}
	}
	if c.OwnedTokenIndex == nil {
		if err := c.fetchOwnedTokenIndices(); err != nil {
			return err
		}
	}
	balance, err := c.BalanceOf(to)
	if err != nil && err != ErrNoExist {
		return err
	}
	// Make sure the token is actually owned by the from address.
	tokenIndex, ok := c.OwnedTokenIndex[tokenID]
	if !ok {
		return ErrNoExist
	}
	// Make sure the from address has tokens to begin with.
	if _, ok := c.OwnedTokens[from]; !ok {
		return ErrNoExist
	}
	// remove token from "from" address
	delete(c.TokenOwners, tokenID)
	c.OwnedTokens[from] = append(c.OwnedTokens[from][:tokenIndex], c.OwnedTokens[from][tokenIndex+1:]...)
	if len(c.OwnedTokens[from]) == 0 {
		delete(c.OwnedTokens, from)
	}
	delete(c.OwnedTokenIndex, tokenID)

	// add token to "to" address
	c.TokenOwners[tokenID] = to
	c.OwnedTokens[to] = append(c.OwnedTokens[to], tokenID)
	c.OwnedTokenIndex[tokenID] = balance
	return nil
}

// TotalSupply returns the current known supply of the token. This supply is updated
// every time a new token is minted.
func (c *DefaultContract) TotalSupply() (*big.Int, error) {
	if totalSupply, err := BigIntString(c.TotalTokens); err == nil {
		return totalSupply, nil
	}
	totalSupply, err := c.fetchTotalSupply()
	if err != nil {
		return BigZero, err
	}
	c.TotalTokens = totalSupply.String()
	return totalSupply, nil
}

// TokensOwnedBy returns the list of token ids owned by owner.
func (c *DefaultContract) TokensOwnedBy(owner string) ([]string, error) {
	if c.OwnedTokens == nil {
		if err := c.fetchOwnedTokens(); err != nil {
			return nil, err
		}
	}
	if tokens, ok := c.OwnedTokens[owner]; ok {
		return tokens, nil
	}
	return nil, ErrNoExist
}

// GetDragonObject fetches an object with the provided key from the DragonChain smart
// contract's heap. An error is returned if the object could not be fetched.
func (c *DefaultContract) GetDragonObject(key string) ([]byte, error) {
	resp, err := c.client.GetSmartContractObject(key, "")
	if err != nil {
		return nil, err
	}
	// TODO: Handle not found case.
	if !resp.OK {
		return nil, fmt.Errorf("bad status code %d received from DragonChain GetSmartContractObject API request", resp.Status)
	}
	return resp.Response.([]byte), nil
}

func (c *DefaultContract) removeToken(from, tid string) error {
	if c.TokenOwners == nil {
		if err := c.fetchTokenOwners(); err != nil {
			return err
		}
	}
	if c.OwnedTokens == nil {
		if err := c.fetchOwnedTokens(); err != nil {
			return err
		}
	}
	if c.OwnedTokenIndex == nil {
		if err := c.fetchOwnedTokenIndices(); err != nil {
			return err
		}
	}
	totalTokens, err := c.TotalSupply()
	if err != nil {
		return err
	}
	tokenIndex, ok := c.OwnedTokenIndex[tid]
	if !ok {
		return ErrNoExist
	}
	// remove token from "from" address
	delete(c.TokenOwners, tid)
	c.OwnedTokens[from] = append(c.OwnedTokens[from][:tokenIndex], c.OwnedTokens[from][tokenIndex+1:]...)
	if len(c.OwnedTokens[from]) == 0 {
		delete(c.OwnedTokens, from)
	}
	delete(c.OwnedTokenIndex, tid)
	c.TotalTokens = totalTokens.Sub(totalTokens, bigOne).String()
	return nil
}

func (c *DefaultContract) addToken(to, tid string) error {
	// If the token ID already exists, we don't want to reuse it.
	if _, ok := c.TokenOwners[tid]; ok {
		return ErrAlreadyExists
	}
	if c.TokenOwners == nil {
		if err := c.fetchTokenOwners(); err != nil {
			return err
		}
	}
	if c.OwnedTokens == nil {
		if err := c.fetchOwnedTokens(); err != nil {
			return err
		}
	}
	if c.OwnedTokenIndex == nil {
		if err := c.fetchOwnedTokenIndices(); err != nil {
			return err
		}
	}
	totalTokens, err := c.TotalSupply()
	if err != nil {
		return err
	}
	// add token to "to" address
	c.TokenOwners[tid] = to
	balance, err := c.BalanceOf(to)
	if err != nil && err != ErrNoExist {
		return err
	}
	c.OwnedTokens[to] = append(c.OwnedTokens[to], tid)
	c.OwnedTokenIndex[tid] = balance
	c.TotalTokens = totalTokens.Add(totalTokens, bigOne).String()
	return nil
}

func (c *DefaultContract) fetchOwnedTokens() error {
	resp, err := c.GetDragonObject("ownedTokens")
	if err != nil {
		return err
	}
	var m map[string][]string
	if err = json.Unmarshal(resp, &m); err != nil {
		return err
	}
	c.OwnedTokens = m
	return nil
}

func (c *DefaultContract) fetchTokenOwners() error {
	resp, err := c.GetDragonObject("tokenOwners")
	if err != nil {
		return err
	}
	var m map[string]string
	if err = json.Unmarshal(resp, &m); err != nil {
		return err
	}
	c.TokenOwners = m
	return nil
}

func (c *DefaultContract) fetchOwnedTokenIndices() error {
	resp, err := c.GetDragonObject("ownedTokenIndex")
	if err != nil {
		return err
	}
	var m map[string]uint64
	if err = json.Unmarshal(resp, &m); err != nil {
		return err
	}
	c.OwnedTokenIndex = m
	return nil
}

func (c *DefaultContract) fetchTotalSupply() (*big.Int, error) {
	resp, err := c.GetDragonObject("totalSupply")
	if err != nil {
		return nil, err
	}
	i, err := BigIntString(string(resp))
	if err == ErrInvalidBigIntString {
		return BigZero, nil
	}
	return i, err
}

// BigIntString is a convenience function for creating a big.Int from string. The string is assumed to be
// a base 10 number. If the big.Int could not be created from the provided string, the second boolean return
// argument will be false.
func BigIntString(s string) (*big.Int, error) {
	bi := &big.Int{}
	bi, ok := bi.SetString(s, 10)
	if !ok {
		return BigZero, ErrInvalidBigIntString
	}
	return bi, nil
}

// DefaultContractFactory creates a new DefaultContract from the heap.
type DefaultContractFactory struct{}

// CreateContract returns a new DefaultContract.
func (f *DefaultContractFactory) CreateContract(name, symbol string) (Contract, error) {
	dcClient, err := dragonClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create dragonchain client: %s", err)
	}
	return NewDefaultContract(name, symbol, dcClient), nil
}

func dragonClient() (*dragonchain.Client, error) {
	httpClient := &http.Client{}
	creds, err := dragonchain.NewCredentials("", "", "", dragonchain.HashSHA256)
	if err != nil {
		return nil, err
	}
	baseAPIURL := os.Getenv("DC_BASE_API_URL")
	client := dragonchain.NewClient(creds, baseAPIURL, httpClient)
	return client, nil
}
