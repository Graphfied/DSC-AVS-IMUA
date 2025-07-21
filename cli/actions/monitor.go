package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	sdkutils "github.com/imua-xyz/imua-avs-sdk/utils"
	avs "github.com/imua-xyz/imua-avs/contracts/bindings/avs"
	"github.com/imua-xyz/imua-avs/core/config"
	"github.com/imua-xyz/imua-avs/types"

	"github.com/urfave/cli"
)

func Monitor(ctx *cli.Context) error {
	configPath := ctx.GlobalString(config.FileFlag.Name)
	nodeConfig := types.NodeConfig{}
	err := sdkutils.ReadYamlConfig(configPath, &nodeConfig)
	if err != nil {
		return err
	}
	// need to make sure we don't register the operator on startup
	// when using the cli commands to register the operator.
	nodeConfig.RegisterOperatorOnStartup = false
	configJson, err := json.MarshalIndent(nodeConfig, "", "  ")
	if err != nil {
		log.Fatalf(err.Error())
	}
	log.Println("Config:", string(configJson))

	client, err := ethclient.Dial(nodeConfig.EthWsUrl)
	if err != nil {
		log.Fatal("Connection failed:", err)
	}
	defer client.Close()

	contractAddress := common.HexToAddress(nodeConfig.AVSAddress)
	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddress},
	}

	logs := make(chan ethtypes.Log)
	sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		log.Fatal("Subscribe failed:", err)
	}
	defer sub.Unsubscribe()

	fmt.Println("Starting event monitoring...")

	for {
		select {
		case err := <-sub.Err():
			log.Fatal("Subscription error:", err)
		case vLog := <-logs:
			// Parse logs
			event, err := parseEvent(vLog)
			if err != nil {
				log.Printf("Parse error: %v", err)
				continue
			}

			// Handle events
			switch e := event.(type) {
			case *avs.ContracthelloWorldTaskCreated:
				fmt.Printf("New Task Created:\n"+
					"  TaskID: %v\n"+
					"  Issuer: %s\n"+
					"  Name: %s\n"+
					"  Number: %d\n"+
					"  Response Period: %d\n"+
					"  Challenge Period: %d\n"+
					"  Threshold: %d%%\n"+
					"  Statistical Period: %d\n",
					e.TaskId, e.Issuer.Hex(), e.Name, e.NumberToBeSquared,
					e.TaskResponsePeriod, e.TaskChallengePeriod,
					e.ThresholdPercentage, e.TaskStatisticalPeriod)

			case *avs.ContracthelloWorldTaskResolved:
				fmt.Printf("Task Resolved:\n"+
					"  TaskID: %d\n"+
					"  Address: %s\n",
					e.TaskId, e.TaskAddress.Hex())
			}
		}
	}
}

func parseEvent(vLog ethtypes.Log) (interface{}, error) {
	contractABI, _ := avs.ContracthelloWorldMetaData.GetAbi()

	switch vLog.Topics[0] {
	case contractABI.Events["TaskCreated"].ID:
		return parseTaskCreated(vLog)
	case contractABI.Events["TaskResolved"].ID:
		return parseTaskResolved(vLog)
	default:
		return nil, fmt.Errorf("unknown event type")
	}
}

func parseTaskCreated(vLog ethtypes.Log) (*avs.ContracthelloWorldTaskCreated, error) {
	var event avs.ContracthelloWorldTaskCreated
	contractABI, _ := avs.ContracthelloWorldMetaData.GetAbi()

	err := contractABI.UnpackIntoInterface(&event, "TaskCreated", vLog.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack TaskCreated: %v", err)
	}

	return &event, nil
}

func parseTaskResolved(vLog ethtypes.Log) (*avs.ContracthelloWorldTaskResolved, error) {
	var event avs.ContracthelloWorldTaskResolved
	contractABI, _ := avs.ContracthelloWorldMetaData.GetAbi()

	err := contractABI.UnpackIntoInterface(&event, "TaskResolved", vLog.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack TaskResolved: %v", err)
	}

	return &event, nil
}
