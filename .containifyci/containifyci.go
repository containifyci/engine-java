//go:generate bash -c "if [ ! -f go.mod ]; then echo 'Initializing go.mod...'; go mod init .containifyci; else echo 'go.mod already exists. Skipping initialization.'; fi"
//go:generate go get github.com/containifyci/engine-ci/protos2
//go:generate go get github.com/containifyci/engine-ci/client
//go:generate go mod tidy

package main

import (
	"os"

	"github.com/containifyci/engine-ci/client/pkg/build"
	"github.com/containifyci/engine-ci/protos2"
)

func registryAuth() map[string]*protos2.ContainerRegistry {
	return map[string]*protos2.ContainerRegistry{
		"docker.io": {
			Username: "env:DOCKER_USER",
			Password: "env:DOCKER_TOKEN",
		},
	}
}

func main() {
	os.Chdir("..")
	opts := build.NewGoServiceBuild("containifyci-java")
	opts.Verbose = false
	opts.File = "main.go"
	opts.Properties = map[string]*build.ListValue{
		"tags": build.NewList("containers_image_openpgp"),
	}
	opts.Registries = registryAuth()
	build.Serve(opts)
}
