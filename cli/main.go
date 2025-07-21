package main

import (
	"log"
	"os"

	"github.com/imua-xyz/imua-avs/cli/actions"
	"github.com/imua-xyz/imua-avs/core/config"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()

	app.Flags = []cli.Flag{config.FileFlag}
	app.Commands = []cli.Command{
		{
			Name:    "register-operator-with-chain",
			Aliases: []string{"rel"},
			Usage:   "registers operator with chain",
			Action:  actions.RegisterOperatorWithChain,
		},
		{
			Name:    "register-operator-with-avs",
			Aliases: []string{"r"},
			Usage:   "operator opt-in avs ",
			Action:  actions.RegisterOperatorWithAvs,
		},
		{
			Name:    "deregister-operator-with-avs",
			Aliases: []string{"d"},
			Action: func(ctx *cli.Context) error {
				log.Fatal("Command not implemented.")
				return nil
			},
		},
		{
			Name:    "print-operator-status",
			Aliases: []string{"s"},
			Usage:   "prints operator status as viewed from avs contracts",
			Action:  actions.PrintOperatorStatus,
		},
		{
			Name:    "monitor",
			Aliases: []string{"m"},
			Usage:   "Subscribe to events using websocket,Monitor create and challenge tasks",
			Action:  actions.Monitor,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
