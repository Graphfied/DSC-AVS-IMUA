package main

import (
	"context"
	"encoding/json"
	"fmt"
	sdkutils "github.com/imua-xyz/imua-avs-sdk/utils"
	"github.com/imua-xyz/imua-avs/types"
	"log"
	"os"

	"github.com/urfave/cli"

	"github.com/imua-xyz/imua-avs/avs"
	"github.com/imua-xyz/imua-avs/core/config"
)

var (
	Version   string
	GitCommit string
	GitDate   string
)

func main() {
	app := cli.NewApp()
	app.Flags = []cli.Flag{config.FileFlag}
	app.Version = fmt.Sprintf("%s-%s-%s", Version, GitCommit, GitDate)
	app.Name = "hello-avs-demo"
	app.Usage = "hello-avs-demo"
	app.Description = "Service that operator listens to AVS contract events, signs tasks, and submits results."

	app.Action = avsMain
	err := app.Run(os.Args)
	if err != nil {
		log.Fatalln("Application failed.", "Message:", err)
	}
}

func avsMain(ctx *cli.Context) error {
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

	log.Println("initializing avs")

	agg, err := avs.NewAvs(&nodeConfig)
	if err != nil {
		return err
	}

	err = agg.Start(context.Background())
	if err != nil {
		return err
	}

	return nil
}
