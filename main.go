package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/containifyci/engine-ci/cmd"
	"github.com/containifyci/java/pkg/maven"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	repo    = "github.com/containifyci/engine-java"
)

func main() {
	v := cmd.SetVersionInfo(version, commit, date, repo)
	slog.Info("Version", "version", v)

	// somewhere after your commands have been created (e.g. in init() of a package)
	command, _, err := cmd.RootCmd().Find([]string{"run"})
	if err == nil && command != nil && command.Name() == "run" {
		slog.Info("Command", "command", command, "err", err)

		oldFnc := command.RunE
		// replace the handler, keep flags/persistent flags/subcommands
		command.RunE = func(command *cobra.Command, args []string) error {
			err := os.Setenv("CONTAINIFYCI_FILE", ".containifyci/containifyci.go")
			if err != nil {
				return err
			}

			arg := cmd.GetBuild(cmd.RootArgs.Auto)

			bld := cmd.Init(arg[0].Builds[0])
			fmt.Printf("Build Args: %+v\n", bld)

			// Get the default pipeline (preserves all existing steps in correct order)
			bs := cmd.GetDefaultBuildSteps(bld)

			// bs.AddToCategory(build.Build, maven.New(bld))
			// bs.AddToCategory(build.PostBuild, maven.NewProd(bld))

			// Replace default Maven steps with enhanced versions from engine-java
			err = bs.Replace("maven", maven.New())
			if err != nil {
				return err
			}
			err = bs.Replace("maven-prod", maven.NewProd())
			if err != nil {
				return err
			}

			// Configure the update command for your binary
			cmd.ConfigureUpdate(
				"engine-java",  // Your binary name
				"containifyci", // Your GitHub organization
				"engine-java",  // Your GitHub repository
			)

			// Set version info (optional but recommended)
			cmd.SetVersionInfo(version, commit, date, repo)
			return oldFnc(command, args)
		}
		command.Short = "engine-java (overridden)"
	}

	err = cmd.Execute()
	if err != nil {
		slog.Error("Main Error", "error", err)
		os.Exit(1)
	}

	slog.Info("Version", "version", v)
}
