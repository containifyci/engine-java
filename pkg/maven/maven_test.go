package maven

import (
	"errors"
	"fmt"
	"testing"

	"github.com/containifyci/engine-ci/pkg/container"
	"github.com/containifyci/engine-ci/pkg/cri"
	"github.com/containifyci/engine-ci/pkg/cri/critest"
	"github.com/containifyci/engine-ci/pkg/logger"
	"github.com/containifyci/engine-ci/pkg/network"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func InitTest(t *testing.T) *container.Build {
	network.RuntimeOS = "darwin"
	t.Setenv("SSH_AUTH_SOCK", "/tmp/ssh-auth.sock")
	t.Setenv("CONTAINER_RUNTIME", "test")

	logger.NewLogAggregator("")
	arg := &container.Build{
		App: "test",
		Custom: map[string][]string{
			"CONTAINIFYCI_HOST": {"localhost"},
			"from":              {"v17"},
		},
		Image:     "test-image",
		BuildType: container.Maven,
	}
	arg.Defaults()
	container.NewBuild(arg)

	cRuntime, err := cri.InitContainerRuntime()
	assert.NoError(t, err)
	if v, ok := cRuntime.(*critest.MockContainerManager); ok {
		v.Reset()
	}
	return arg
}

func TestNew(t *testing.T) {
	build := InitTest(t)

	mc := new(build)
	matches := Matches(*build)
	assert.True(t, matches)
	assert.Equal(t, "maven", mc.Name())
	assert.False(t, mc.IsAsync())
	assert.Equal(t, "test-image", mc.Image)
	assert.Equal(t, "v17", mc.Version)
	assert.Equal(t, []string{"containifyci/maven-3-eclipse-temurin-v17-alpine:cdbe73779492603b08a3e880bf25754e3a8e865811c51c0b45e2c5edfc5a8476", "tomcat:latest"}, Images(*build))
}

func TestNewProd(t *testing.T) {
	build := InitTest(t)

	mc := NewProd()
	assert.Equal(t, "maven-prod", mc.Name())
	assert.False(t, mc.IsAsync())
	assert.Equal(t, []string{"tomcat:latest"}, mc.Images(*build))
}

func TestPull(t *testing.T) {
	build := InitTest(t)

	mc := new(build)
	err := mc.Pull()
	assert.NoError(t, err)

	cRuntime, err := cri.InitContainerRuntime()
	assert.NoError(t, err)

	if v, ok := cRuntime.(*critest.MockContainerManager); ok {
		assert.Equal(t, "tomcat:latest pulled", v.ImagesLogEntries[0])
	}
}

func TestBuildLinuxPodman(t *testing.T) {
	expectedEnvs := []string{
		"MAVEN_OPTS=-Xms512m -Xmx512m -XX:MaxDirectMemorySize=512m",
		"SSH_AUTH_SOCK=/tmp/ssh-auth.sock",
		"CONTAINIFYCI_HOST=localhost",
		"DOCKER_HOST=unix://var/run/podman.sock",
		"TESTCONTAINERS_RYUK_DISABLED=true",
	}

	arg := InitTest(t)
	arg.Platform.Host.OS = "linux"
	arg.Runtime = "podman"

	mc := new(arg)
	matches := Matches(*arg)
	assert.True(t, matches)
	assert.Equal(t, "v17", mc.Version)
	err := mc.Build()
	require.NoError(t, err)

	cRuntime, err := cri.InitContainerRuntime()
	assert.NoError(t, err)

	if v, ok := cRuntime.(*critest.MockContainerManager); ok {
		img := "containifyci/maven-3-eclipse-temurin-v17-alpine:cdbe73779492603b08a3e880bf25754e3a8e865811c51c0b45e2c5edfc5a8476"
		require.Len(t, v.ContainerLogsEntries[img], 2)
		assert.Equal(t, []string{"container starting", "container running"}, v.ContainerLogsEntries[img])

		assert.Equal(t, "started", v.GetContainerByImage(img).State)
		assert.Equal(t, []string{"sh", "/tmp/script.sh"}, v.GetContainerByImage(img).Opts.Cmd)

		fmt.Printf("Envs %+v\n", v.GetContainerByImage(img).Opts.Env)
		for _, env := range expectedEnvs {
			assert.Contains(t, v.GetContainerByImage(img).Opts.Env, env)
		}
	}
}

func TestBuildDarwinPodman(t *testing.T) {
	expectedEnvs := []string{
		"MAVEN_OPTS=-Xms512m -Xmx512m -XX:MaxDirectMemorySize=512m",
		"TC_HOST=host.containers.internal",
		"TESTCONTAINERS_HOST_OVERRIDE=host.containers.internal",
		"CONTAINIFYCI_HOST=localhost",
		"DOCKER_HOST=unix://var/run/podman.sock",
		"TESTCONTAINERS_RYUK_DISABLED=true",
	}

	arg := InitTest(t)
	arg.Platform.Host.OS = "darwin"
	arg.Runtime = "podman"

	mc := new(arg)
	matches := Matches(*arg)
	assert.True(t, matches)
	assert.Equal(t, "v17", mc.Version)

	cRuntime, err := cri.InitContainerRuntime()
	assert.NoError(t, err)

	if v, ok := cRuntime.(*critest.MockContainerManager); ok {
		v.Errors["containifyci/maven-3-eclipse-temurin-v17-alpine:cdbe73779492603b08a3e880bf25754e3a8e865811c51c0b45e2c5edfc5a8476"] = errors.New("image not found")

		err := mc.Run()
		assert.NoError(t, err)

		img := "containifyci/maven-3-eclipse-temurin-v17-alpine:cdbe73779492603b08a3e880bf25754e3a8e865811c51c0b45e2c5edfc5a8476"
		assert.Len(t, v.ContainerLogsEntries[img], 2)
		assert.Equal(t, []string{"container starting", "container running"}, v.ContainerLogsEntries[img])

		assert.Equal(t, "started", v.GetContainerByImage(img).State)
		assert.Equal(t, []string{"sh", "/tmp/script.sh"}, v.GetContainerByImage(img).Opts.Cmd)
		assert.Equal(t, "#!/bin/sh\nset -xe\ncd .\nmvn --batch-mode package\n", v.GetContainerByImage(img).Opts.Script)
		assert.Equal(t, "/src", v.GetContainerByImage(img).Opts.WorkingDir)
		assert.Equal(t, "containifyci/maven-3-eclipse-temurin-v17-alpine:cdbe73779492603b08a3e880bf25754e3a8e865811c51c0b45e2c5edfc5a8476", v.GetContainerByImage(img).Opts.Image)
		assert.Equal(t, int64(4073741824), v.GetContainerByImage(img).Opts.Memory)
		assert.Equal(t, uint64(2048), v.GetContainerByImage(img).Opts.CPU)

		fmt.Printf("Envs %+v\n", v.GetContainerByImage(img).Opts.Env)
		for _, env := range expectedEnvs {
			assert.Contains(t, v.GetContainerByImage(img).Opts.Env, env)
		}

		//expect 3 images opnejdk image and maven-3-eclipse-temurin-17-alpine twice for both platforms amd64 and arm64
		assert.Len(t, v.Images, 3)
		assert.NotNil(t, v.Images["tomcat:latest"])
		assert.Equal(t, "linux/amd64", v.Images["containifyci/maven-3-eclipse-temurin-v17-alpine:cdbe73779492603b08a3e880bf25754e3a8e865811c51c0b45e2c5edfc5a8476-linux/amd64"].Opts.Platform.String())
		assert.Equal(t, "linux/arm64", v.Images["containifyci/maven-3-eclipse-temurin-v17-alpine:cdbe73779492603b08a3e880bf25754e3a8e865811c51c0b45e2c5edfc5a8476-linux/arm64"].Opts.Platform.String())
	}
}

func TestProd(t *testing.T) {
	arg := InitTest(t)
	arg.Platform.Host.OS = "darwin"
	arg.Runtime = "podman"

	mc := NewProd()

	cRuntime, err := cri.InitContainerRuntime()
	assert.NoError(t, err)

	if v, ok := cRuntime.(*critest.MockContainerManager); ok {
		err := mc.RunWithBuild(*arg)
		assert.NoError(t, err)

		img := "tomcat:latest"
		assert.Len(t, v.ContainerLogsEntries[img], 3)
		assert.Equal(t, []string{"container starting", "container running", "container stopped"}, v.ContainerLogsEntries[img])

		assert.Equal(t, "stopped", v.GetContainerByImage(img).State)
		assert.Equal(t, []string{"sleep", "300"}, v.GetContainerByImage(img).Opts.Cmd)
		assert.Equal(t, "tomcat:latest", v.GetContainerByImage(img).Opts.Image)

		fmt.Printf("Envs %+v\n", v.GetContainerByImage(img).Opts.Env)
	} else {
		t.Fatal("Container runtime is not a MockContainerManager")
	}
}
