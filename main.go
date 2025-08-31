package main

import (
	"log/slog"
	"os"

	"github.com/containifyci/engine-ci/cmd"
	"github.com/containifyci/engine-ci/pkg/build"
	"github.com/containifyci/engine-ci/pkg/github"
	"github.com/containifyci/engine-ci/pkg/sonarcloud"
	"github.com/containifyci/engine-ci/pkg/trivy"
	"github.com/containifyci/java/pkg/maven"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	repo    = "github.com/containifyci/java"
)

func main() {
	v := cmd.SetVersionInfo(version, commit, date, repo)
	slog.Info("Version", "version", v)

	arg := cmd.GetBuild()

	bld := cmd.Init(arg[0].Builds[0])

	bs := build.NewBuildSteps(
		append(maven.Steps(bld),
			sonarcloud.New(*bld),
			trivy.New(*bld),
			github.New(*bld))...,
	)
	cmd.InitBuildSteps(bs)
	err := cmd.Execute()
	if err != nil {
		slog.Error("Main Error", "error", err)
		os.Exit(1)
	}
}
