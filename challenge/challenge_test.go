package challenge_test

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/imua-xyz/imua-avs/core/chainio/eth"
	"log"
	"testing"
	"time"
)

// doc：https://docs.infura.io/api/networks/ethereum/json-rpc-methods/eth_getlogs
func TestEth_getlogs(t *testing.T) {
	ethRpcClient, err := eth.NewClient("http://localhost:8545")
	if err != nil {
		log.Fatal("Cannot create http ethclient", "err", err)
	}
	// Contract address and ABI
	_ = common.HexToAddress("0xaD6864A88b832100750Ff35881851c943e5BAc34")
	if err != nil {
		log.Fatal(err)
	}

	if err != nil {
		log.Fatal(err)
	}
	firstHeight, err := ethRpcClient.BlockNumber(context.Background())

	height := firstHeight
	fmt.Printf("Event firstHeight: %v\n", firstHeight)

	for {
		currentHeight, err := ethRpcClient.BlockNumber(context.Background())
		fmt.Printf("Event currentHeight: %v\n", currentHeight)

		if err != nil {
			log.Fatal(err)
		}
		if currentHeight == height+1 {

			height = currentHeight
		}
		time.Sleep(2 * time.Second)
	}
}
