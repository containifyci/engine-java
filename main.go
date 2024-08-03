package main

import (
	"log/slog"
	"os"

	"github.com/containifyci/engine-ci/cmd"
	"github.com/containifyci/engine-ci/pkg/build"
	"github.com/containifyci/engine-ci/pkg/container"

	"github.com/containifyci/engine-ci/pkg/github"
	"github.com/containifyci/engine-ci/pkg/sonarcloud"
	"github.com/containifyci/engine-ci/pkg/trivy"

	"github.com/containifyci/java/pkg/maven"
)

func main() {
	arg := cmd.GetBuild()
	cmd.Init(arg...)

	bs := build.NewBuildSteps(
		append(maven.Steps(container.GetBuild()),
		sonarcloud.New(),
		trivy.New(),
		github.New())...,
	)

	cmd.InitBuildSteps(bs)
	err := cmd.Execute()
	if err != nil {
		slog.Error("Main Error", "error", err)
		os.Exit(1)
	}
}

