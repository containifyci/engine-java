//go:generate bash -c "if [ ! -f go.mod ]; then echo 'Initializing go.mod...'; go mod init .containifyci; else echo 'go.mod already exists. Skipping initialization.'; fi"
//go:generate go get github.com/containifyci/engine-ci/protos2
//go:generate go get github.com/containifyci/engine-ci/client
//go:generate go mod tidy

package main

import (
	"os"

	"github.com/containifyci/engine-ci/client/pkg/build"
)

func main() {
	os.Chdir("../")
	opts := build.NewMavenServiceBuild("hello-world-servlet")
	opts.Verbose = false
	opts.File = "target/hello-world-servlet.war"
	//TODO: adjust the registry to your own container registry
	opts.Properties = map[string]*build.ListValue{
		"push": build.NewList("false"),
	}
	opts.Registry = "containifyci"
	build.Serve(opts)
}
