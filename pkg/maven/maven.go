package maven

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"os/user"
	"path/filepath"

	"github.com/containifyci/engine-ci/pkg/build"
	"github.com/containifyci/engine-ci/pkg/container"
	"github.com/containifyci/engine-ci/pkg/cri/types"
	"github.com/containifyci/engine-ci/pkg/cri/utils"

	"github.com/containifyci/engine-ci/pkg/filesystem"
	"github.com/containifyci/engine-ci/pkg/network"
	u "github.com/containifyci/engine-ci/pkg/utils"
)

const (
	ProdImage     = "registry.access.redhat.com/ubi8/openjdk-%s:latest"
	CacheLocation = "/root/.m2/"
)

//go:embed Dockerfile.*
var f embed.FS

type MavenContainer struct {
	App      string
	File     string
	Folder   string
	Image    string
	ImageTag string
	Platform types.Platform

	Version string
	*container.Container
}

func New(version string) *MavenContainer {
	return &MavenContainer{
		App:       container.GetBuild().App,
		Container: container.New(container.BuildEnv),
		Image:     container.GetBuild().Image,
		Folder:    container.GetBuild().Folder,
		ImageTag:  container.GetBuild().ImageTag,
		Platform:  container.GetBuild().Platform,
		Version:  version,
	}
}

func (c *MavenContainer) IsAsync() bool {
	return false
}

func (c *MavenContainer) Name() string {
	return "maven"
}

func CacheFolder() string {
	mvnHome := u.GetEnvs([]string{"MAVEN_HOME", "CONTAINIFYCI_CACHE"}, "build")
	if mvnHome == "" {
		usr, err := user.Current()
		if err != nil {
			slog.Error("Failed to get current user", "error", err)
			os.Exit(1)
		}
		mvnHome = fmt.Sprintf("%s%s%s", usr.HomeDir, string(os.PathSeparator), ".m2")
		slog.Info("MAVEN_HOME not set, using default", "mavenHome", mvnHome)
		err = filesystem.DirectoryExists(mvnHome)
		if err != nil {
			slog.Error("Failed to create cache folder", "error", err)
			os.Exit(1)
		}
	}
	return mvnHome
}

func (c *MavenContainer) Pull() error {
	return c.Container.Pull(fmt.Sprintf(ProdImage, c.Version))
}

func (c *MavenContainer) Images() []string {
	return []string{c.MavenImage(), fmt.Sprintf(ProdImage, c.Version)}
}

// TODO: provide a shorter checksum
func ComputeChecksum(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func (c *MavenContainer) MavenImage() string {
	fileName := fmt.Sprintf("Dockerfile.maven_%s-jdk-jammy", c.Version)
	dockerFile, err := f.ReadFile(fileName)
	if err != nil {
		slog.Error("Failed to read Dockerfile.maven", "error", err)
		os.Exit(1)
	}
	tag := ComputeChecksum(dockerFile)
	image := fmt.Sprintf("maven-3-eclipse-temurin-%s-alpine", c.Version)
	return utils.ImageURI(container.GetBuild().ContainifyRegistry, image, tag)
	// return fmt.Sprintf("%s/%s/%s:%s", container.GetBuild().Registry, "containifyci", "maven-3-eclipse-temurin-17-alpine", tag)
}

func (c *MavenContainer) BuildMavenImage() error {
	image := c.MavenImage()
	fileName := fmt.Sprintf("Dockerfile.maven_%s-jdk-jammy", c.Version)
	slog.Debug("Building maven image", "image", image, "fileName", fileName)
	dockerFile, err := f.ReadFile(fileName)
	if err != nil {
		slog.Error("Failed to read Dockerfile.maven", "error", err)
		os.Exit(1)
	}

	platforms := types.GetPlatforms(container.GetBuild().Platform)
	slog.Info("Building intermediate image", "image", image, "platforms", platforms)

	err = c.Container.BuildIntermidiateContainer(image, dockerFile, platforms...)
	if err != nil {
		slog.Error("Failed to build maven image", "error", err)
		os.Exit(1)
	}
	return nil
}

func (c *MavenContainer) Address() *network.Address {
	return &network.Address{Host: "localhost"}
}

func (c *MavenContainer) Build() error {
	imageTag := c.MavenImage()

	ssh, err := network.SSHForward()
	if err != nil {
		slog.Error("Failed to forward SSH", "error", err)
		os.Exit(1)
	}

	opts := types.ContainerConfig{}
	opts.Image = imageTag
	opts.Env = append(opts.Env, []string{
		"MAVEN_OPTS=-Xms512m -Xmx512m -XX:MaxDirectMemorySize=512m",
	}...)

	// On MacOS, we need to set a special docker host so that the testcontainers can access the host
	if c.Platform.Host.OS == "darwin" {
		opts.Env = append(opts.Env, []string{
			fmt.Sprintf("TC_HOST=%s", c.Address().ForContainerDefault()),
			fmt.Sprintf("TESTCONTAINERS_HOST_OVERRIDE=%s", c.Address().ForContainerDefault()),
			fmt.Sprintf("CONTAINIFYCI_HOST=%s", getContainifyHost()),
		}...)
	}

	opts.WorkingDir = "/src"

	dir, _ := filepath.Abs(".")

	opts.Volumes = []types.Volume{
		{
			Type:   "bind",
			Source: dir,
			Target: "/src",
		},
		{
			Type:   "bind",
			Source: CacheFolder(),
			Target: CacheLocation,
		},
	}
	opts.Memory = int64(4073741824)
	opts.CPU = uint64(2048)

	opts = ssh.Apply(&opts)
	opts = utils.ApplySocket(container.GetBuild().Runtime, &opts)

	if container.GetBuild().Runtime == utils.Podman {
		//https://stackoverflow.com/questions/71549856/testcontainers-with-podman-in-java-tests
		opts.Env = append(opts.Env, []string{
			"DOCKER_HOST=unix://var/run/podman.sock",
			"TESTCONTAINERS_RYUK_DISABLED=true",
			//TODO identify if we need privileged mode or not
		}...)
	}

	if privilged := u.GetEnv("CONTAINER_PRIVILGED", "build"); privilged == "false" {
		opts.Env = append(opts.Env,
			"TESTCONTAINERS_RYUK_DISABLED=true",
			"TESTCONTAINERS_RYUK_PRIVILEGED=false",
		)
	}

	opts.Script = c.BuildScript()

	err = c.Container.BuildingContainer(opts)
	if err != nil {
		slog.Error("Failed to build container", "error", err)
		os.Exit(1)
	}

	return err
}

//TODO should be moved to the engine-ci itself.
func getContainifyHost() string {
	if v, ok := container.GetBuild().Custom["CONTAINIFYCI_HOST"]; ok {
		return v[0]
	}
	return ""
}

func (c *MavenContainer) BuildScript() string {
	// Create a temporary script in-memory
	return Script(NewBuildScript(c.Container.Verbose, getContainifyHost()))
}

type MavenBuild struct {
	rf     build.RunFunc
	name   string
	images []string
	async  bool
}

func (g MavenBuild) Run() error {
	return g.rf()
}

func (g MavenBuild) Name() string {
	return g.name
}

func (g MavenBuild) Images() []string {
	return g.images
}

func (g MavenBuild) IsAsync() bool {
	return g.async
}

func NewProd(version string) build.Build {
	container := New(version)
	return MavenBuild{
		rf: func() error {
			return container.Prod()
		},
		name:   "maven-prod",
		images: []string{ProdImage},
		async:  false,
	}
}

func (c *MavenContainer) Prod() error {
	opts := types.ContainerConfig{}
	opts.Image = fmt.Sprintf(ProdImage, c.Version)
	opts.Env = []string{
		"JAVA_OPTS=-javaagent:/deployments/dd-java-agent.jar -Dquarkus.http.host=0.0.0.0 -Djava.util.logging.manager=org.jboss.logmanager.LogManager",
		"JAVA_APP_JAR=/deployments/quarkus-run.jar",
	}
	opts.Platform = types.AutoPlatform
	opts.Cmd = []string{"sleep", "300"}
	opts.User = "185"
	opts.WorkingDir = "/src"

	err := c.Container.Create(opts)
	if err != nil {
		slog.Error("Failed to create container: %s", "error", err)
		os.Exit(1)
	}

	err = c.Container.Start()
	if err != nil {
		slog.Error("Failed to start container: %s", "error", err)
		os.Exit(1)
	}

	err = c.Container.Exec("curl", "-Lo", "/deployments/dd-java-agent.jar", "https://dtdg.co/latest-java-tracer")
	if err != nil {
		slog.Error("Failed to execute command: %s", "error", err)
		os.Exit(1)
	}

	err = c.Container.CopyDirectoryTo(c.Folder, "/deployments")
	if err != nil {
		slog.Error("Failed to copy directory to container: %s", "error", err)
		os.Exit(1)
	}

	imageId, err := c.Container.Commit(fmt.Sprintf("%s:%s", c.Image, c.ImageTag), "Created from container", "CMD [\"/usr/local/s2i/run\"]", "USER 185")
	if err != nil {
		slog.Error("Failed to commit container: %s", "error", err)
		os.Exit(1)
	}

	err = c.Container.Stop()
	if err != nil {
		slog.Error("Failed to stop container: %s", "error", err)
		os.Exit(1)
	}

	imageUri := utils.ImageURI(container.GetBuild().Registry, c.Image, c.ImageTag)
	err = c.Container.Push(imageId, imageUri)
	if err != nil {
		slog.Error("Failed to push image: %s", "error", err)
		os.Exit(1)
	}

	return err
}

func (c *MavenContainer) Run() error {
	err := c.Pull()
	if err != nil {
		slog.Error("Failed to pull base images: %s", "error", err)
		return err
	}

	err = c.BuildMavenImage()
	if err != nil {
		slog.Error("Failed to build go image: %s", "error", err)
		return err
	}

	err = c.Build()
	slog.Info("Container created", "containerId", c.Container.ID)
	if err != nil {
		slog.Error("Failed to create container: %s", "error", err)
		return err
	}
	return nil
}
