package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/revrost/counterspell/pkg/agent"
)

func EnableDebug() {
	h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	slog.SetDefault(slog.New(h))
}

func runCommand(ctx context.Context, cmd *cli.Command) error {
	yamlFile := cmd.Args().First()
	if yamlFile == "" {
		return errors.New("usage: cspell run <yaml_file>")
	}
	if err := agent.RunAgentWithCagent(yamlFile); err != nil {
		return fmt.Errorf("run failed: %w", err)
	}
	return nil
}

func main() {
	EnableDebug()
	cmd := &cli.Command{
		Name:  "cspell",
		Usage: "A CLI for running agents",
		Commands: []*cli.Command{
			{
				Name:   "run",
				Usage:  "run an agent from a yaml file",
				Action: runCommand,
			},
		},
	}
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
