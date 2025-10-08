package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/revrost/counterspell/pkg/cagent"
)

func runCommand(ctx context.Context, cmd *cli.Command) error {
	cfgPath := cmd.Args().First()
	if cfgPath == "" {
		return errors.New("usage: cs run <yaml_file> [message|-]")
	}

	// Optional first user message or stdin
	var firstMessage string
	if second := cmd.Args().Get(1); second != "" {
		if second == "-" {
			b, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read stdin: %w", err)
			}
			firstMessage = string(b)
		} else {
			firstMessage = second
		}
	}

	autoApprove := true
	if cmd != nil {
		if cmd.Bool("auto-approve") || cmd.Bool("y") {
			autoApprove = true
		}
	}
	r := cagent.NewRunnerFromPath(cfgPath, cagent.WithFirstMessage(firstMessage), cagent.WithAutoApproveTools(autoApprove))
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
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "y", Usage: "auto-approve tool calls", Value: true},
					&cli.BoolFlag{Name: "auto-approve", Usage: "auto-approve tool calls", Value: true},
				},
				Action: runCommand,
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}