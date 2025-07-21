package main

import (
	"fmt"
	"github.com/imua-xyz/imua-avs/cmd/imua-key/generate"
	"os"

	"github.com/imua-xyz/imua-avs/cmd/imua-key/import"
	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.NewApp()
	app.Name = "imua-key"
	app.Description = "Imua key manager"
	app.Commands = []*cli.Command{
		_import.Command,
		generate.Command,
	}

	app.Usage = "Import keys for testing purpose"

	if err := app.Run(os.Args); err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}

}
