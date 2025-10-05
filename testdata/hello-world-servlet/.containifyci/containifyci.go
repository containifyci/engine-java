//go:generate sh -c "if [ ! -f go.mod ]; then echo 'Initializing go.mod...'; go mod init .containifyci; else echo 'go.mod already exists. Skipping initialization.'; fi"
//go:generate go get github.com/containifyci/engine-ci/protos2
//go:generate go get github.com/containifyci/engine-ci/client
//go:generate go mod tidy

package main

import (
	"os"

	"github.com/containifyci/engine-ci/client/pkg/build"
	"github.com/containifyci/engine-ci/protos2"
)

func main() {
	os.Chdir("../")

	// Build Group 0
	tomcat7mavenplugin := build.NewMavenServiceBuild("tomcat7-maven-plugin")
	tomcat7mavenplugin.Folder = "."
	tomcat7mavenplugin.File = "target/hello-world-servlet.war"
	tomcat7mavenplugin.Properties = map[string]*build.ListValue{
		"push": build.NewList("false"),
	}

	//TODO: adjust the registries to your own container registry
	build.BuildGroups(
		&protos2.BuildArgsGroup{
			Args: []*protos2.BuildArgs{
				tomcat7mavenplugin,
			},
		},
	)
}
