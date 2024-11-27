package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
	"github.com/vkuznecovas/relemb/cli/commands"
)

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			commands.NewUpdateRelated(),
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
