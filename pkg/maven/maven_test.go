package maven

import (
	"errors"
	"fmt"
	"testing"

	"github.com/containifyci/engine-ci/pkg/container"
	"github.com/containifyci/engine-ci/pkg/cri"
	"github.com/containifyci/engine-ci/pkg/cri/critest"
	"github.com/containifyci/engine-ci/pkg/network"
	"github.com/stretchr/testify/assert"
)

func InitTest(t *testing.T) *container.Build {
	network.RuntimeOS = "darwin"
	t.Setenv("SSH_AUTH_SOCK", "/tmp/ssh-auth.sock")
	t.Setenv("CONTAINER_RUNTIME", "test")
	arg := &container.Build{
		App:    "test",
		Custom: map[string][]string{"CONTAINIFYCI_HOST": {"localhost"}},
		Image: "test-image",
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
	InitTest(t)

	mc := New("17")
	assert.Equal(t, "maven", mc.Name())
	assert.False(t, mc.IsAsync())
	assert.Equal(t, "test-image", mc.Image)
	assert.Equal(t, "17", mc.Version)
	assert.Equal(t, []string{"containifyci/maven-3-eclipse-temurin-17-alpine:214f702cd3dee211717ca17936aca88df6913e1db40f822ef9918e4369ea927d", "registry.access.redhat.com/ubi8/openjdk-17:latest"}, mc.Images())
}

func TestNewProd(t *testing.T) {
	InitTest(t)

	mc := NewProd("17")
	assert.Equal(t, "maven-prod", mc.Name())
	assert.False(t, mc.IsAsync())
	assert.Equal(t, []string{"registry.access.redhat.com/ubi8/openjdk-17:latest"}, mc.Images())
}

func TestPull(t *testing.T) {
	InitTest(t)

	mc := New("17")
	mc.Pull()

	cRuntime, err := cri.InitContainerRuntime()
	assert.NoError(t, err)

	if v, ok := cRuntime.(*critest.MockContainerManager); ok {
		assert.Equal(t, "registry.access.redhat.com/ubi8/openjdk-17:latest pulled", v.ImagesLogEntries[0])
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

	mc := New("17")
	mc.Build()

	cRuntime, err := cri.InitContainerRuntime()
	assert.NoError(t, err)

	if v, ok := cRuntime.(*critest.MockContainerManager); ok {
		img := "containifyci/maven-3-eclipse-temurin-17-alpine:214f702cd3dee211717ca17936aca88df6913e1db40f822ef9918e4369ea927d"
		assert.Len(t, v.ContainerLogsEntries[img], 2)
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

	mc := New("17")

	cRuntime, err := cri.InitContainerRuntime()
	assert.NoError(t, err)

	if v, ok := cRuntime.(*critest.MockContainerManager); ok {
		v.Errors["containifyci/maven-3-eclipse-temurin-17-alpine:214f702cd3dee211717ca17936aca88df6913e1db40f822ef9918e4369ea927d"] = errors.New("image not found")

		mc.Run()

		img := "containifyci/maven-3-eclipse-temurin-17-alpine:214f702cd3dee211717ca17936aca88df6913e1db40f822ef9918e4369ea927d"
		assert.Len(t, v.ContainerLogsEntries[img], 2)
		assert.Equal(t, []string{"container starting", "container running"}, v.ContainerLogsEntries[img])

		assert.Equal(t, "started", v.GetContainerByImage(img).State)
		assert.Equal(t, []string{"sh", "/tmp/script.sh"}, v.GetContainerByImage(img).Opts.Cmd)
		assert.Equal(t, "#!/bin/sh\nset -xe\n./mvnw --batch-mode package\n", v.GetContainerByImage(img).Opts.Script)
		assert.Equal(t, "/src", v.GetContainerByImage(img).Opts.WorkingDir)
		assert.Equal(t, "containifyci/maven-3-eclipse-temurin-17-alpine:214f702cd3dee211717ca17936aca88df6913e1db40f822ef9918e4369ea927d", v.GetContainerByImage(img).Opts.Image)
		assert.Equal(t, int64(4073741824), v.GetContainerByImage(img).Opts.Memory)
		assert.Equal(t, uint64(2048), v.GetContainerByImage(img).Opts.CPU)

		fmt.Printf("Envs %+v\n", v.GetContainerByImage(img).Opts.Env)
		for _, env := range expectedEnvs {
			assert.Contains(t, v.GetContainerByImage(img).Opts.Env, env)
		}

		//expect 3 images opnejdk image and maven-3-eclipse-temurin-17-alpine twice for both platforms amd64 and arm64
		assert.Len(t, v.Images, 3)
		assert.NotNil(t, v.Images["registry.access.redhat.com/ubi8/openjdk-17:latest"])
		assert.Equal(t, "linux/amd64", v.Images["containifyci/maven-3-eclipse-temurin-17-alpine:214f702cd3dee211717ca17936aca88df6913e1db40f822ef9918e4369ea927d-linux/amd64"].Opts.Platform.String())
		assert.Equal(t, "linux/arm64", v.Images["containifyci/maven-3-eclipse-temurin-17-alpine:214f702cd3dee211717ca17936aca88df6913e1db40f822ef9918e4369ea927d-linux/arm64"].Opts.Platform.String())
	}
}

func TestProd(t *testing.T) {
	expectedEnvs := []string{
		"JAVA_OPTS=-javaagent:/deployments/dd-java-agent.jar -Dquarkus.http.host=0.0.0.0 -Djava.util.logging.manager=org.jboss.logmanager.LogManager",
		"JAVA_APP_JAR=/deployments/quarkus-run.jar",
	}

	arg := InitTest(t)
	arg.Platform.Host.OS = "darwin"
	arg.Runtime = "podman"

	mc := NewProd("17")

	cRuntime, err := cri.InitContainerRuntime()
	assert.NoError(t, err)

	if v, ok := cRuntime.(*critest.MockContainerManager); ok {
		mc.Run()

		img := "registry.access.redhat.com/ubi8/openjdk-17:latest"
		assert.Len(t, v.ContainerLogsEntries[img], 3)
		assert.Equal(t, []string{"container starting", "container running", "container stopped"}, v.ContainerLogsEntries[img])

		assert.Equal(t, "stopped", v.GetContainerByImage(img).State)
		assert.Equal(t, []string{"sleep", "300"}, v.GetContainerByImage(img).Opts.Cmd)
		assert.Equal(t, "185", v.GetContainerByImage(img).Opts.User)
		assert.Equal(t, "/src", v.GetContainerByImage(img).Opts.WorkingDir)
		assert.Equal(t, "registry.access.redhat.com/ubi8/openjdk-17:latest", v.GetContainerByImage(img).Opts.Image)

		fmt.Printf("Envs %+v\n", v.GetContainerByImage(img).Opts.Env)
		for _, env := range expectedEnvs {
			assert.Contains(t, v.GetContainerByImage(img).Opts.Env, env)
		}
	}
}
