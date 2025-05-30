// Copyright 2021-2024 LiveKit, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/livekit/protocol/logger"
	lksdk "github.com/livekit/server-sdk-go/v2"

	livekitcli "github.com/livekit/livekit-cli/v2"
)

type appDidntRunForEnoughTimeError struct{}

func (appDidntRunForEnoughTimeError) Error() string {
	return "app didn't run for the requested amount of time"
}
func (appDidntRunForEnoughTimeError) Timeout() bool   { return true }
func (appDidntRunForEnoughTimeError) Temporary() bool { return true }

func main() {

	mustRunFor := parseMustRunFor(os.Args)

	ctx, cancel := context.WithCancel(context.Background())
	if mustRunFor > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), mustRunFor)
	}

	defer cancel()

	// Run the application with the provided arguments
	err := runApp(ctx, os.Args)

	if mustRunFor > 0 {
		if err == context.DeadlineExceeded { // we ran out of time and nothing happened = success
			err = nil
		} else {
			err = appDidntRunForEnoughTimeError{}
		}
	}

	if err != nil {
		logger.Errorw("test failed", err)
		os.Exit(1)
	}
}

func parseMustRunFor(args []string) time.Duration {
	var mustRunForFlag string

	// Iterate through command-line arguments to find --must-run-for
	for i, arg := range args {
		if strings.HasPrefix(arg, "--must-run-for=") {
			// Handle --must-run-for=<x>
			mustRunForFlag = strings.TrimPrefix(arg, "--must-run-for=")
			break
		} else if arg == "--must-run-for" && i+1 < len(args) {
			// Handle --must-run-for <x>
			mustRunForFlag = args[i+1]
			break
		}
	}

	if mustRunForFlag == "" {
		return 0 // Default to 0 if --must-run-for is not provided
	}

	// Parse the duration
	duration, err := time.ParseDuration(mustRunForFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid duration for --must-run-for: %v\n", err)
		os.Exit(1)
	}

	return duration
}

// runApp runs the application with the given arguments and handles its lifecycle
func runApp(ctx context.Context, args []string) error {
	// Create a channel to capture errors from the app
	errCh := make(chan error, 1)

	// Run the app in a separate goroutine
	go func() {
		app := &cli.Command{
			Name:                   "lk",
			Usage:                  "CLI client to LiveKit",
			Description:            "A suite of command line utilities allowing you to access LiveKit APIs services, interact with rooms in realtime, and perform load testing simulations.",
			Version:                livekitcli.Version,
			EnableShellCompletion:  true,
			Suggest:                true,
			HideHelpCommand:        true,
			UseShortOptionHandling: true,
			Flags:                  globalFlags,
			Commands: []*cli.Command{
				{
					Name:   "generate-fish-completion",
					Action: generateFishCompletion,
					Hidden: true,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:    "out",
							Aliases: []string{"o"},
						},
					},
				},
			},
			Before: initLogger,
		}

		app.Commands = append(app.Commands, AppCommands...)
		app.Commands = append(app.Commands, AgentCommands...)
		app.Commands = append(app.Commands, CloudCommands...)
		app.Commands = append(app.Commands, ProjectCommands...)
		app.Commands = append(app.Commands, RoomCommands...)
		app.Commands = append(app.Commands, TokenCommands...)
		app.Commands = append(app.Commands, JoinCommands...)
		app.Commands = append(app.Commands, DispatchCommands...)
		app.Commands = append(app.Commands, EgressCommands...)
		app.Commands = append(app.Commands, IngressCommands...)
		app.Commands = append(app.Commands, SIPCommands...)
		app.Commands = append(app.Commands, ReplayCommands...)
		app.Commands = append(app.Commands, LoadTestCommands...)
		app.Commands = append(app.Commands, AgentLoadTestCommands...)

		checkForLegacyName()

		// Run the app and send any errors to the channel
		errCh <- app.Run(ctx, args)
	}()

	// Wait for the app to complete or the context to timeout
	select {
	case <-ctx.Done():
		// If the context times out, return the timeout error
		return ctx.Err()
	case err := <-errCh:
		// Return any errors from the app
		return err
	}
}

func checkForLegacyName() {
	if !(strings.HasSuffix(os.Args[0], "lk") || strings.HasSuffix(os.Args[0], "lk.exe")) {
		fmt.Fprintf(
			os.Stderr,
			"\n~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~ DEPRECATION NOTICE ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~\n"+
				"The `livekit-cli` binary has been renamed to `lk`, and some of the options and\n"+
				"commands have changed. Though legacy commands my continue to work, they have\n"+
				"been hidden from the USAGE notes and may be removed in future releases."+
				"\n~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~\n\n",
		)
	}
}

func initLogger(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	logConfig := &logger.Config{
		Level: "info",
	}
	if cmd.Bool("verbose") {
		logConfig.Level = "debug"
	}
	logger.InitFromConfig(logConfig, "lk")
	lksdk.SetLogger(logger.GetLogger())

	return nil, nil
}

func generateFishCompletion(ctx context.Context, cmd *cli.Command) error {
	fishScript, err := cmd.ToFishCompletion()
	if err != nil {
		return err
	}

	outPath := cmd.String("out")
	if outPath != "" {
		if err := os.WriteFile(outPath, []byte(fishScript), 0o644); err != nil {
			return err
		}
	} else {
		fmt.Println(fishScript)
	}

	return nil
}
