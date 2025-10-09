package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/revrost/counterspell/pkg/clearcast"
	"github.com/urfave/cli/v3"
)

// func runCagent(ctx context.Context, cmd *cli.Command) error {
// 	cfgPath := cmd.Args().First()
// 	if cfgPath == "" {
// 		return errors.New("usage: cs run <yaml_file> [message|-]")
// 	}
//
// 	// Optional first user message or stdin
// 	var firstMessage string
// 	if second := cmd.Args().Get(1); second != "" {
// 		if second == "-" {
// 			b, err := io.ReadAll(os.Stdin)
// 			if err != nil {
// 				return fmt.Errorf("failed to read stdin: %w", err)
// 			}
// 			firstMessage = string(b)
// 		} else {
// 			firstMessage = second
// 		}
// 	}
//
// 	autoApprove := true
// 	if cmd != nil {
// 		if cmd.Bool("auto-approve") || cmd.Bool("y") {
// 			autoApprove = true
// 		}
// 	}
// 	r := cagent.NewRunnerFromPath(cfgPath, cagent.WithFirstMessage(firstMessage), cagent.WithAutoApproveTools(autoApprove))
// 	if err := r.Run(); err != nil {
// 		return fmt.Errorf("runner error: %w", err)
// 	}
// 	return nil
// }

func runClearcast(ctx context.Context, cmd *cli.Command) error {
	yamlFile := cmd.Args().First()
	if yamlFile == "" {
		return errors.New("usage: cspell run <yaml_file>")
	}

	rt, err := clearcast.NewRunFromPath(yamlFile)
	if err != nil {
		panic(err)
	}

	eventsChan := rt.RunStream(ctx)

	for event := range eventsChan {
		switch e := event.(type) {
		case *clearcast.PlanResultEvent:
			fmt.Println("=== Plan Created ===")
			for i, plan := range e.Plans {
				fmt.Printf("Step %d: %s with params %v\n", i+1, plan.Tool, plan.Params)
			}
		case *clearcast.AgentChoiceEvent:
			fmt.Println("\n=== Agent Output ===")
			fmt.Println(e.Content)
		case *clearcast.ErrorEvent:
			return fmt.Errorf("runtime error: %s", e.Error)
		case *clearcast.FinalEvent:
			fmt.Println("\n=== Execution Finished ===")
			fmt.Println(e.Output)
		case *clearcast.ToolResultEvent:
			fmt.Println("\n=== Tool Output ===")
			fmt.Println(e.Result)
		case *clearcast.ToolErrorEvent:
			fmt.Println("\n=== Tool Error ===")
			fmt.Println(e.Error)

		default:
			// Optionally log unhandled event types
			slog.Warn("Received unhandled event type: %T\n", e)
		}
	}
	return nil
}

func main() {
	app := &cli.Command{
		Name:  "cs",
		Usage: "Run counterspell configurations",
		Commands: []*cli.Command{
			{
				Name:   "run",
				Usage:  "run a counterspell agent from a yaml file",
				Flags:  []cli.Flag{},
				Action: runClearcast,
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
