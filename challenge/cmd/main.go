package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli"

	"github.com/imua-xyz/imua-avs/challenge"
	"github.com/imua-xyz/imua-avs/core/config"
	"github.com/imua-xyz/imua-avs/types"

	sdkutils "github.com/imua-xyz/imua-avs-sdk/utils"
)

func main() {
	app := cli.NewApp()
	app.Flags = []cli.Flag{config.FileFlag, config.TaskIDFlag, config.NumberToBeSquaredFlag, config.ExecTypeFlag}
	app.Name = "hello-world-demo-challenge"
	app.Usage = "hello-world-demo Challenge"
	app.Description = "Service that challenger listens to AVS contract events, Initiate challenges and validate the tasks already submitted by the operator."

	app.Action = challengeMain
	err := app.Run(os.Args)
	if err != nil {
		log.Fatalln("Application failed. Message:", err)
	}
}

func challengeMain(ctx *cli.Context) error {
	log.Println("Initializing challenge")
	configPath := ctx.GlobalString(config.FileFlag.Name)
	nodeConfig := types.NodeConfig{}
	err := sdkutils.ReadYamlConfig(configPath, &nodeConfig)
	if err != nil {
		return err
	}
	configJson, err := json.MarshalIndent(nodeConfig, "", "  ")
	if err != nil {
		log.Fatalf(err.Error())
	}
	log.Println("Config:", string(configJson))

	log.Println("initializing challenge")
	challenger, err := challenge.NewChallengeFromConfig(nodeConfig)
	if err != nil {
		return err
	}
	log.Println("initialized challenge")
	execType := ctx.Int(config.ExecTypeFlag.Name)
	if execType == 1 {
		err = challenger.Start(context.Background())
		if err != nil {
			return err
		}
		log.Println("challenger started")
	}
	if execType == 2 {
		taskID := ctx.Uint64(config.TaskIDFlag.Name)
		numBeSquared := ctx.Uint64(config.NumberToBeSquaredFlag.Name)
		if taskID == 0 || numBeSquared == 0 {
			return fmt.Errorf("task ID and Number to be squared must be provided")
		}
		log.Println("starting manual challenge")
		err = challenger.Exec(context.Background(), taskID, numBeSquared)
		if err != nil {
			return err
		}
		log.Println("challenger completed successfully")
	}
	return nil
}
