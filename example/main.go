package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"nft"
	"os"

	"github.com/dragonchain/dragonchain-sdk-go"
)

var (
	dcID       string
	apiKey     string
	apiKeyID   string
	baseAPIURL string
	contractID string
)

func main() {
	contract := &nft.Contract{}
	b, err := heap()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get heap: %s", err)
		os.Exit(1)
	}
	if err := json.Unmarshal(b, contract); err != nil {
		fmt.Fprintf(os.Stderr, "failed to JSON unmarshal heap: %s", err)
		os.Exit(1)
	}
	rpc, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read STDIN: %s", err)
		os.Exit(1)
	}
	handleRPC(rpc, contract)
	if err := json.NewEncoder(os.Stdout).Encode(contract); err != nil {
		fmt.Fprintf(os.Stderr, "failed to JSON encode contract state to stdout: %s", err)
		os.Exit(1)
	}
}

func handleRPC(rpc []byte, contract *nft.Contract) {
	// interpret RPC and handle with contract.
}

func heap() ([]byte, error) {
	httpClient := &http.Client{}
	creds, err := dragonchain.NewCredentials(dcID, apiKey, apiKeyID, dragonchain.HashSHA256)
	if err != nil {
		return nil, err
	}
	client := dragonchain.NewClient(creds, baseAPIURL, httpClient)
	resp, err := client.GetSmartContractObject("", "")
	if err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New("Non OK response")
	}
	return resp.Response.([]byte), nil
}
