package operator_test

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	avs "github.com/imua-xyz/imua-avs/contracts/bindings/avs"
	"github.com/imua-xyz/imua-avs/core"
	"github.com/imua-xyz/imua-avs/core/chainio/eth"
	"github.com/prysmaticlabs/prysm/v5/crypto/bls/blst"
)

func TestDecodeRes(t *testing.T) {

	base64Str := "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABQg=="
	data, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		fmt.Println("Error decoding Base64:", err)
		return
	}

	hexStr := "0x" + hex.EncodeToString(data)

	fmt.Println("Hexadecimal format:", hexStr)
}

func TestBlsSig(t *testing.T) {
	// BLS12-381 Signed Message
	//ChainIDWithoutRevision: imuachainlocalnet_232
	//AccAddressBech32: im16tltge7d4yr0wtkr7ut6dwwnqgnwm2ge63djdp
	privateKey, _ := blst.RandKey()
	msg := fmt.Sprintf(core.BLSMessageToSign,
		core.ChainIDWithoutRevision("imuachainlocalnet_232-1"), "im16tltge7d4yr0wtkr7ut6dwwnqgnwm2ge63djdp")

	hashedMsg := crypto.Keccak256Hash([]byte(msg))
	sig := privateKey.Sign(hashedMsg.Bytes())
	fmt.Println(msg)

	fmt.Println(hexutil.Encode(privateKey.PublicKey().Marshal()))
	fmt.Println(hexutil.Encode(sig.Marshal()))

}
func TestAbi(t *testing.T) {

	task := core.TaskResponse{
		TaskID:        10,
		NumberSquared: 56169,
	}

	packed, err := core.Args.Pack(&task)
	if err != nil {
		t.Errorf("Error packing task: %v", err)
		return
	} else {
		t.Logf("ABI encoded: %s", hexutil.Encode(packed))
	}

	args := make(map[string]interface{})

	err = core.Args.UnpackIntoMap(args, packed)
	result, _ := core.Args.Unpack(packed)
	t.Logf("Unpacked: %v", result[0])
	hash := crypto.Keccak256Hash(packed)
	t.Logf("Hash: %s", hash.String())

	key := args["TaskResponse"]
	t.Logf("Key: %v", key)

}
func TestEvent(t *testing.T) {

	dataStr := "000000000000000000000000000000000000000000000000000000000000000100000000000000000000000010ed22d975453a5d4031440d51624552e4f204d50000000000000000000000004b99e597121c99ba5846c32bd49d8a4b95457f8c0000000000000000000000000000000000000000000000000000000000000064000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000e0000000000000000000000000000000000000000000000000000000000000012000000000000000000000000000000000000000000000000000000000000000010000000000000000000000003e108c058e8066da635321dc3018294ca82ddedf00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000"
	contractAbi, _ := avs.ContracthelloWorldMetaData.GetAbi()
	event := contractAbi.Events["TaskCreated"]

	eventInputs := event.Inputs
	data, _ := hex.DecodeString(dataStr)
	eventArgs, err := eventInputs.Unpack(data)

	if err != nil {
		log.Fatal(err)
	}

	taskId := eventArgs[0].(*big.Int)
	fmt.Printf("Task ID: %v\n", taskId)
	issuer := eventArgs[1].(common.Address)
	fmt.Printf("Issuer: %s\n", issuer.Hex())

	name := eventArgs[2].(string)
	fmt.Printf("Name: %s\n", name)
	num := eventArgs[3].(uint64)
	fmt.Printf("num: %d\n", num)

	taskResponsePeriod := eventArgs[4].(uint64)
	fmt.Printf("Task Response Period: %d\n", taskResponsePeriod)
	taskChallengePeriod := eventArgs[5].(uint64)
	fmt.Printf("Task Challenge Period: %d\n", taskChallengePeriod)
	thresholdPercentage := eventArgs[6].(uint8)
	fmt.Printf("Threshold Percentage: %d\n", thresholdPercentage)
	taskStatisticalPeriod := eventArgs[7].(uint64)
	fmt.Printf("Task Statistical Period: %d\n", taskStatisticalPeriod)
}

// doc：https://docs.infura.io/api/networks/ethereum/json-rpc-methods/eth_getlogs
func TestEth_GetLogs(t *testing.T) {
	ethRpcClient, err := eth.NewClient("http://localhost:8545")
	if err != nil {
		log.Fatal("Cannot create http ethclient", "err", err)
	}
	// Contract address and ABI
	contractAddress := common.HexToAddress("0x10Ed22D975453A5D4031440D51624552E4f204D5")
	if err != nil {
		log.Fatal(err)
	}

	if err != nil {
		log.Fatal(err)
	}
	firstHeight, err := ethRpcClient.BlockNumber(context.Background())

	GetLog(ethRpcClient, contractAddress, int64(29))

	height := firstHeight
	fmt.Printf("Event firstHeight: %v\n", firstHeight)

	for {
		currentHeight, err := ethRpcClient.BlockNumber(context.Background())
		fmt.Printf("Event currentHeight: %v\n", currentHeight)

		if err != nil {
			log.Fatal(err)
		}
		if currentHeight == height+1 {
			GetLog(ethRpcClient, contractAddress, int64(currentHeight))

			height = currentHeight
		}
		time.Sleep(2 * time.Second)
	}
}

func GetLog(client eth.EthClient, address common.Address, height int64) {
	query := ethereum.FilterQuery{
		Addresses: []common.Address{address},
		FromBlock: big.NewInt(height),
		ToBlock:   big.NewInt(height),
	}

	logs, err := client.FilterLogs(context.Background(), query)
	if err != nil {
		log.Fatal(err)
	}
	if logs != nil {
		contractAbi, _ := avs.ContracthelloWorldMetaData.GetAbi()
		event := contractAbi.Events["TaskCreated"]

		for _, vLog := range logs {
			fmt.Println(vLog.BlockHash.Hex()) // 0x3404b8c050aa0aacd0223e91b5c32fee6400f357764771d0684fa7b3f448f1a8
			fmt.Println(vLog.BlockNumber)     // 2394201
			fmt.Println(vLog.TxHash.Hex())    // 0x280201eda63c9ff6f305fcee51d5eb86167fab40ca3108ec784e8652a0e2b1a6
			data := vLog.Data
			fmt.Println(vLog.Topics)
			fmt.Println(event.ID)

			eventArgs, err := event.Inputs.Unpack(data)

			log.Println("parse logs",
				"data", data,
				"height", height,
				"event", event.Inputs)
			if err != nil {
				fmt.Println("Not as expected log, parse err:", err)
				return
			}

			taskId := eventArgs[0].(*big.Int)
			fmt.Println("taskId:", taskId)

			issuer := eventArgs[1].(common.Address)
			fmt.Println("Issuer:", issuer.Hex())

			name := eventArgs[2].(string)
			fmt.Println("name:", name)
			numberToBeSquared := eventArgs[3].(uint64)
			fmt.Println("numberToBeSquared:", numberToBeSquared)

			taskResponsePeriod := eventArgs[4].(uint64)
			fmt.Println("Task Response Period: ", taskResponsePeriod)

			taskChallengePeriod := eventArgs[5].(uint64)
			fmt.Println("Task Challenge Period:", taskChallengePeriod)

			thresholdPercentage := eventArgs[6].(uint8)
			fmt.Println("Threshold Percentage:", thresholdPercentage)

			taskStatisticalPeriod := eventArgs[7].(uint64)

			fmt.Println("Task Statistical Period:", taskStatisticalPeriod)
		}
	}

}
