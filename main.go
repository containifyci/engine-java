package main

import (
	"fmt"
	"os"

	"github.com/containifyci/engine-ci/cmd"
	"github.com/containifyci/engine-ci/pkg/build"

	"github.com/containifyci/engine-ci/pkg/github"
	"github.com/containifyci/engine-ci/pkg/sonarcloud"
	"github.com/containifyci/engine-ci/pkg/trivy"

	_ "github.com/containifyci/java/pkg/maven"
	maven "github.com/containifyci/java/pkg/maven/v21"
	_ "github.com/containifyci/java/pkg/maven/v17"
)

func main() {
	cmd.Init(cmd.GetBuild()...)

	bs := build.NewBuildSteps(
		maven.New(),
		maven.NewProd(),
		sonarcloud.New(),
		trivy.New(),
		github.New(),
	)

	cmd.InitBuildSteps(bs)
	err := cmd.Execute()
	if err != nil {
		fmt.Printf("Main Error: %v", err)
		os.Exit(1)
	}
}

