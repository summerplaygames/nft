package nft

import (
	"errors"
	"math/big"
	"net/http"
	"testing"

	"github.com/dragonchain/dragonchain-sdk-go"
	"github.com/stretchr/testify/assert"
)

type dcResp struct {
	Response string
	Error    error
}

var (
	errFailed    = errors.New("failed")
	balanceTests = map[string]struct {
		DCResponse      *dcResp
		DefaultState    map[string][]string
		Input           string
		ExpectedBalance uint64
		ExpectedError   error
	}{
		"fetch": {
			DCResponse: &dcResp{
				Response: `{"owner": ["1"]}`,
			},
			Input:           "owner",
			ExpectedBalance: 1,
		},
		"fetch error": {
			DCResponse: &dcResp{
				Error: errFailed,
			},
			Input:           "owner",
			ExpectedBalance: 0,
			ExpectedError:   errFailed,
		},
		"balance exists": {
			DefaultState:    map[string][]string{"owner": {"1"}},
			Input:           "owner",
			ExpectedBalance: 1,
		},
		"owner not found": {
			DCResponse: &dcResp{
				Response: `{"owner": ["1"]}`,
			},
			Input:           "owner2",
			ExpectedBalance: 0,
			ExpectedError:   ErrNoExist,
		},
	}

	ownerOfTests = map[string]struct {
		DCResponse    *dcResp
		DefaultState  map[string]string
		Input         string
		ExpectedOwner string
		ExpectedError error
	}{
		"fetch": {
			DCResponse: &dcResp{
				Response: `{"tokenID": "owner"}`,
			},
			Input:         "tokenID",
			ExpectedOwner: "owner",
		},
		"fetch error": {
			DCResponse: &dcResp{
				Error: errFailed,
			},
			Input:         "tokenID",
			ExpectedOwner: "",
			ExpectedError: errFailed,
		},
		"owner exists": {
			DefaultState:  map[string]string{"tokenID": "owner"},
			Input:         "tokenID",
			ExpectedOwner: "owner",
		},
		"token not found": {
			DCResponse: &dcResp{
				Response: `{"tokenID": "owner"}`,
			},
			Input:         "tokenID2",
			ExpectedOwner: "",
			ExpectedError: ErrNoExist,
		},
	}

	mintTests = map[string]struct {
		TokenOwners   map[string]string
		OwnedTokens   map[string][]string
		TokenIndicies map[string]uint64
		TotalSupply   *big.Int
		To            string
		TokenID       string
		ExpectedError error
	}{
		"token minted": {
			TokenOwners:   map[string]string{},
			OwnedTokens:   map[string][]string{},
			TokenIndicies: map[string]uint64{},
			TotalSupply:   BigZero,
			To:            "owner2",
			TokenID:       "tokenID2",
		},
		"token already exists": {
			TokenOwners:   map[string]string{"tokenID": "owner"},
			OwnedTokens:   map[string][]string{"owner": {"tokenID"}},
			TokenIndicies: map[string]uint64{"tokenID": 0},
			TotalSupply:   bigOne,
			To:            "owner",
			TokenID:       "tokenID",
			ExpectedError: ErrAlreadyExists,
		},
	}

	burnTests = map[string]struct {
		TokenOwners   map[string]string
		OwnedTokens   map[string][]string
		TokenIndicies map[string]uint64
		TotalSupply   *big.Int
		From          string
		TokenID       string
		ExpectedError error
	}{
		"token burned": {
			TokenOwners:   map[string]string{"tokenID": "owner"},
			OwnedTokens:   map[string][]string{"owner": {"tokenID"}},
			TokenIndicies: map[string]uint64{"tokenID": 0},
			TotalSupply:   bigOne,
			From:          "owner",
			TokenID:       "tokenID",
		},
		"token no exist": {
			TokenOwners:   map[string]string{},
			OwnedTokens:   map[string][]string{},
			TokenIndicies: map[string]uint64{},
			TotalSupply:   BigZero,
			From:          "owner",
			TokenID:       "tokenID",
			ExpectedError: ErrNoExist,
		},
	}

	transferTests = map[string]struct {
		TokenOwners   map[string]string
		OwnedTokens   map[string][]string
		TokenIndicies map[string]uint64
		TotalSupply   *big.Int
		To            string
		From          string
		TokenID       string
		ExpectedError error
	}{
		"token transfered": {
			TokenOwners:   map[string]string{"tokenID": "owner"},
			OwnedTokens:   map[string][]string{"owner": {"tokenID"}},
			TokenIndicies: map[string]uint64{"tokenID": 0},
			TotalSupply:   bigOne,
			To:            "owner2",
			From:          "owner",
			TokenID:       "tokenID",
		},
		"token index no exist": {
			TokenOwners:   map[string]string{},
			OwnedTokens:   map[string][]string{},
			TokenIndicies: map[string]uint64{},
			TotalSupply:   BigZero,
			From:          "owner",
			TokenID:       "tokenID",
			ExpectedError: ErrNoExist,
		},
		"from address no exist": {
			TokenOwners:   map[string]string{},
			OwnedTokens:   map[string][]string{},
			TokenIndicies: map[string]uint64{"tokenID": 0},
			TotalSupply:   BigZero,
			From:          "owner",
			TokenID:       "tokenID",
			ExpectedError: ErrNoExist,
		},
	}

	totalSupplyTests = map[string]struct {
		DCResponse     *dcResp
		DefaultState   string
		ExpectedSupply string
		ExpectedError  error
	}{
		"fetch": {
			DCResponse: &dcResp{
				Response: "1",
			},
			ExpectedSupply: "1",
		},
		"fetch failed": {
			DCResponse: &dcResp{
				Error: errFailed,
			},
			ExpectedSupply: "0",
			ExpectedError:  errFailed,
		},
		"not exist in heap": {
			DCResponse:     &dcResp{},
			ExpectedSupply: "0",
		},
		"exist in memory": {
			DefaultState:   "5",
			ExpectedSupply: "5",
		},
	}

	tokensOwnedByTests = map[string]struct {
		DCResponse     *dcResp
		DefaultState   map[string][]string
		Owner          string
		ExpectedTokens []string
		ExpectedError  error
	}{
		"fetch tokens": {
			DCResponse: &dcResp{
				Response: `{"owner":["tokenID"]}`,
			},
			Owner:          "owner",
			ExpectedTokens: []string{"tokenID"},
		},
		"fetch failed": {
			DCResponse: &dcResp{
				Error: errFailed,
			},
			ExpectedError: errFailed,
		},
		"tokens exist in memory": {
			DefaultState:   map[string][]string{"owner": {"tokenID"}},
			Owner:          "owner",
			ExpectedTokens: []string{"tokenID"},
		},
	}
)

func TestDefaultContract_BalanceOf(t *testing.T) {
	for name, test := range balanceTests {
		t.Run(name, func(t *testing.T) {
			mockClient := &MockClient{}
			contract := NewDefaultContract("test", "TEST", mockClient)
			contract.OwnedTokens = test.DefaultState
			on := test.DCResponse != nil
			var ret *dragonchain.Response
			if on {
				if test.DCResponse.Response != "" {
					ret = &dragonchain.Response{
						OK:       true,
						Status:   http.StatusOK,
						Response: []byte(test.DCResponse.Response),
					}
				}
				mockClient.On("GetSmartContractObject", "ownedTokens", "").Once().Return(ret, test.DCResponse.Error)
			}
			balance, err := contract.BalanceOf(test.Input)
			assert.Equal(t, test.ExpectedError, err)
			assert.Equal(t, test.ExpectedBalance, balance)
			if !on {
				mockClient.AssertNotCalled(t, "GetSmartContractObject", "ownedTokens", "")
			}
		})
	}
}

func TestDefaultContract_OwnerOf(t *testing.T) {
	for name, test := range ownerOfTests {
		t.Run(name, func(t *testing.T) {
			mockClient := &MockClient{}
			contract := NewDefaultContract("test", "TEST", mockClient)
			contract.TokenOwners = test.DefaultState
			on := test.DCResponse != nil
			var ret *dragonchain.Response
			if on {
				if test.DCResponse.Response != "" {
					ret = &dragonchain.Response{
						OK:       true,
						Status:   http.StatusOK,
						Response: []byte(test.DCResponse.Response),
					}
				}
				mockClient.On("GetSmartContractObject", "tokenOwners", "").Once().Return(ret, test.DCResponse.Error)
			}
			owner, err := contract.OwnerOf(test.Input)
			assert.Equal(t, test.ExpectedError, err)
			assert.Equal(t, test.ExpectedOwner, owner)
			if !on {
				mockClient.AssertNotCalled(t, "GetSmartContractObject", "tokenOwners", "")
			}
		})
	}
}

func TestDefaultContract_Mint(t *testing.T) {
	for name, test := range mintTests {
		t.Run(name, func(t *testing.T) {
			mockClient := &MockClient{}
			contract := NewDefaultContract("test", "TEST", mockClient)
			contract.TokenOwners = test.TokenOwners
			contract.OwnedTokens = test.OwnedTokens
			contract.OwnedTokenIndex = test.TokenIndicies
			contract.TotalTokens = test.TotalSupply.String()
			err := contract.Mint(test.To, test.TokenID)
			assert.Equal(t, test.ExpectedError, err)
			assert.Len(t, contract.TokenOwners, 1)
			assert.Len(t, contract.OwnedTokens, 1)
			assert.Len(t, contract.OwnedTokenIndex, 1)
			assert.Equal(t, test.To, contract.TokenOwners[test.TokenID])
			assert.Contains(t, contract.OwnedTokens[test.To], test.TokenID)
			assert.Equal(t, uint64(0), contract.OwnedTokenIndex[test.TokenID])
			assert.Equal(t, "1", contract.TotalTokens)
		})
	}
}

func TestDefaultContract_Burn(t *testing.T) {
	for name, test := range burnTests {
		t.Run(name, func(t *testing.T) {
			mockClient := &MockClient{}
			contract := NewDefaultContract("test", "TEST", mockClient)
			contract.TokenOwners = test.TokenOwners
			contract.OwnedTokens = test.OwnedTokens
			contract.OwnedTokenIndex = test.TokenIndicies
			contract.TotalTokens = test.TotalSupply.String()
			expectedOwners := len(test.TokenOwners) - 1
			expectedTokens := len(test.OwnedTokens) - 1
			expectedIndeices := len(test.TokenIndicies) - 1
			err := contract.Burn(test.TokenID)
			if test.ExpectedError != nil {
				assert.Equal(t, test.ExpectedError, err)
				return
			}
			assert.NoError(t, err)
			assert.Len(t, contract.TokenOwners, expectedOwners)
			assert.Len(t, contract.OwnedTokens, expectedTokens)
			assert.Len(t, contract.OwnedTokenIndex, expectedIndeices)
			n := test.TotalSupply.Sub(test.TotalSupply, bigOne)
			assert.Equal(t, n.String(), contract.TotalTokens)
		})
	}
}

func TestDefaultContract_Transfer(t *testing.T) {
	for name, test := range transferTests {
		t.Run(name, func(t *testing.T) {
			mockClient := &MockClient{}
			contract := NewDefaultContract("test", "TEST", mockClient)
			contract.TokenOwners = test.TokenOwners
			contract.OwnedTokens = test.OwnedTokens
			contract.OwnedTokenIndex = test.TokenIndicies
			contract.TotalTokens = test.TotalSupply.String()
			expectedOwners := len(test.TokenOwners)
			expectedTokens := len(test.OwnedTokens)
			expectedIndeices := len(test.TokenIndicies)
			err := contract.Transfer(test.From, test.To, test.TokenID)
			if test.ExpectedError != nil {
				assert.Equal(t, test.ExpectedError, err)
				return
			}
			assert.NoError(t, err)
			assert.Len(t, contract.TokenOwners, expectedOwners)
			assert.Len(t, contract.OwnedTokens, expectedTokens)
			assert.Len(t, contract.OwnedTokenIndex, expectedIndeices)
			assert.Equal(t, test.TotalSupply.String(), contract.TotalTokens)
			assert.Equal(t, test.To, contract.TokenOwners[test.TokenID])
			assert.Contains(t, contract.OwnedTokens[test.To], test.TokenID)
			assert.Equal(t, uint64(0), contract.OwnedTokenIndex[test.TokenID])
		})
	}
}

func TestDefaultContract_TotalSupply(t *testing.T) {
	for name, test := range totalSupplyTests {
		t.Run(name, func(t *testing.T) {
			mockClient := &MockClient{}
			contract := NewDefaultContract("test", "TEST", mockClient)
			contract.TotalTokens = test.DefaultState
			on := test.DCResponse != nil
			var ret *dragonchain.Response
			if on {
				if test.DCResponse.Error == nil {
					ret = &dragonchain.Response{
						OK:       true,
						Status:   http.StatusOK,
						Response: []byte(test.DCResponse.Response),
					}
				}
				mockClient.On("GetSmartContractObject", "totalSupply", "").Once().Return(ret, test.DCResponse.Error)
			}
			supply, err := contract.TotalSupply()
			assert.Equal(t, test.ExpectedError, err)
			assert.Equal(t, test.ExpectedSupply, supply.String())
			if !on {
				mockClient.AssertNotCalled(t, "GetSmartContractObject", "totalSupply", "")
			}
		})
	}
}

func TestDefaultContract_TokensOwnedBy(t *testing.T) {
	for name, test := range tokensOwnedByTests {
		t.Run(name, func(t *testing.T) {
			mockClient := &MockClient{}
			contract := NewDefaultContract("test", "TEST", mockClient)
			contract.OwnedTokens = test.DefaultState
			on := test.DCResponse != nil
			var ret *dragonchain.Response
			if on {
				if test.DCResponse.Error == nil {
					ret = &dragonchain.Response{
						OK:       true,
						Status:   http.StatusOK,
						Response: []byte(test.DCResponse.Response),
					}
				}
				mockClient.On("GetSmartContractObject", "ownedTokens", "").Once().Return(ret, test.DCResponse.Error)
			}
			tokens, err := contract.TokensOwnedBy(test.Owner)
			assert.Equal(t, test.ExpectedError, err)
			assert.Equal(t, test.ExpectedTokens, tokens)
			if !on {
				mockClient.AssertNotCalled(t, "GetSmartContractObject", "ownedTokens", "")
			}
		})
	}
}
