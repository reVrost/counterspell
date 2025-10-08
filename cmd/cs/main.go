package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/revrost/counterspell/pkg/cagent"
)

func runCommand(_ context.Context, cmd *cli.Command) error {
	cfgPath := cmd.Args().First()
	if cfgPath == "" {
		return errors.New("usage: cs run <yaml_file>")
	}

	r := cagent.NewRunnerFromPath(cfgPath)
	if err := r.Run(); err != nil {
		return fmt.Errorf("runner error: %w", err)
	}
	return nil
}

func main() {
	app := &cli.Command{
		Name:  "cs",
		Usage: "Run cagent configurations",
		Commands: []*cli.Command{
			{
				Name:   "run",
				Usage:  "run a cagent from a yaml file",
				Action: runCommand,
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
