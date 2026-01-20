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
	PRODIMAGE             = "tomcat:latest"
	CacheLocation         = "/root/.m2/"
	DEFAULT_MAVEN_VERSION = "v17"
)

//go:embed Dockerfile.*
var f embed.FS

type MavenContainer struct {
	App       string
	File      u.SrcFile
	Folder    string
	Image     string
	ImageTag  string
	Platform  types.Platform
	ProdImage string

	Version string
	*container.Container
}

func New() build.BuildStepv3 {
	return build.Stepper{
		RunFnV3: func(build container.Build) (string, error) {
			container := new(&build)
			return container.Run()
		},
		MatchedFn: Matches,
		ImagesFn:  Images,
		Name_:     "gorelease",
		Async_:    false,
	}
}

func Matches(build container.Build) bool {
	if build.BuildType != container.Maven {
		return false
	}
	version := GetVersion(build)
	return version == "v17" || version == "v21"
}

func new(build *container.Build) *MavenContainer {
	return &MavenContainer{
		App:       build.App,
		Container: container.New(*build),
		Image:     build.Image,
		Folder:    build.Folder,
		File:      u.SrcFile(build.File),
		ImageTag:  build.ImageTag,
		Platform:  build.Platform,
		ProdImage: ProdImage(*build),
		Version:   GetVersion(*build),
	}
}

func (c *MavenContainer) IsAsync() bool {
	return false
}

func (c *MavenContainer) Name() string {
	return "maven"
}

func GetVersion(build container.Build) string {
	var from string
	if v, ok := build.Custom["from"]; ok {
		slog.Info("Using custom build", "from", v[0])
		from = v[0]
	}
	if from == "" {
		from = DEFAULT_MAVEN_VERSION
	}
	return from
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
	return c.Container.Pull(c.ProdImage)
}

func Images(build container.Build) []string {
	return []string{MavenImage(build), ProdImage(build)}
}

func ProdImage(build container.Build) string {
	prodImage := build.Custom.String("image")
	if prodImage == "" {
		prodImage = PRODIMAGE
	}
	return prodImage
}

// TODO: provide a shorter checksum
func ComputeChecksum(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func MavenImage(build container.Build) string {
	fileName := fmt.Sprintf("Dockerfile.maven_%s-jdk-jammy", GetVersion(build))
	dockerFile, err := f.ReadFile(fileName)
	if err != nil {
		slog.Error("Failed to read Dockerfile.maven", "error", err)
		os.Exit(1)
	}
	tag := ComputeChecksum(dockerFile)
	image := fmt.Sprintf("maven-3-eclipse-temurin-%s-alpine", GetVersion(build))
	return utils.ImageURI(build.ContainifyRegistry, image, tag)
}

func (c *MavenContainer) BuildMavenImage() error {
	image := MavenImage(*c.GetBuild())
	fileName := fmt.Sprintf("Dockerfile.maven_%s-jdk-jammy", c.Version)
	slog.Debug("Building maven image", "image", image, "fileName", fileName)
	dockerFile, err := f.ReadFile(fileName)
	if err != nil {
		slog.Error("Failed to read Dockerfile.maven", "error", err)
		os.Exit(1)
	}

	platforms := types.GetPlatforms(c.GetBuild().Platform)
	slog.Info("Building intermediate image", "image", image, "platforms", platforms)

	err = c.BuildIntermidiateContainer(image, dockerFile, platforms...)
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
	imageTag := MavenImage(*c.GetBuild())

	ssh, err := network.SSHForward(*c.GetBuild())
	if err != nil {
		slog.Error("Failed to forward SSH", "error", err)
		os.Exit(1)
	}

	opts := types.ContainerConfig{}
	opts.Image = imageTag
	opts.Env = append(opts.Env, []string{
		"MAVEN_OPTS=-Xms512m -Xmx512m -XX:MaxDirectMemorySize=512m",
		fmt.Sprintf("CONTAINIFYCI_HOST=%s", getContainifyHost(c.GetBuild())),
	}...)

	// On MacOS, we need to set a special docker host so that the testcontainers can access the host
	if c.Platform.Host.OS == "darwin" {
		opts.Env = append(opts.Env, []string{
			fmt.Sprintf("TC_HOST=%s", c.Address().ForContainerDefault(c.GetBuild())),
			fmt.Sprintf("TESTCONTAINERS_HOST_OVERRIDE=%s", c.Address().ForContainerDefault(c.GetBuild())),
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
	opts = utils.ApplySocket(c.GetBuild().Runtime, &opts)

	if c.GetBuild().Runtime == utils.Podman {
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

	err = c.BuildingContainer(opts)
	if err != nil {
		slog.Error("Failed to build container", "error", err)
		os.Exit(1)
	}

	return err
}

// TODO should be moved to the engine-ci itself.
func getContainifyHost(build *container.Build) string {
	if v, ok := build.Custom["CONTAINIFYCI_HOST"]; ok {
		return v[0]
	}
	return ""
}

func (c *MavenContainer) BuildScript() string {
	// Create a temporary script in-memory
	return Script(NewBuildScript(c.Verbose, c.Folder, getContainifyHost(c.GetBuild())))
}

func NewProd() build.BuildStepv3 {
	return build.Stepper{
		RunFnV3: func(build container.Build) (string, error) {
			c := new(&build)
			if build.Image == "" {
				slog.Info("No image name skip prod image creation")
				return "", nil
			}
			return c.Prod()
		},
		ImagesFn: func(build container.Build) []string {
			return []string{ProdImage(build)}
		},
		MatchedFn: func(build container.Build) bool {
			return build.BuildType == container.Maven
		},
		Name_:  "maven-prod",
		Async_: false,
	}
}

func (c *MavenContainer) Prod() (string, error) {
	opts := types.ContainerConfig{}
	opts.Image = c.ProdImage
	opts.Platform = types.AutoPlatform
	opts.Cmd = []string{"sleep", "300"}

	err := c.Create(opts)
	if err != nil {
		slog.Error("Failed to create container: %s", "error", err)
		os.Exit(1)
	}

	err = c.Start()
	if err != nil {
		slog.Error("Failed to start container: %s", "error", err)
		os.Exit(1)
	}

	fileName := filepath.Base(c.File.Host())
	err = c.CopyFileTo(c.File.Host(), "/usr/local/tomcat/webapps/"+fileName)
	if err != nil {
		slog.Error("Failed to copy file to container", "error", err, "file", c.File)
		os.Exit(1)
	}

	imageId, err := c.Commit(fmt.Sprintf("%s:%s", c.Image, c.ImageTag), "Created from container", "CMD [\"catalina.sh\", \"run\"]")
	if err != nil {
		slog.Error("Failed to commit container: %s", "error", err)
		os.Exit(1)
	}

	err = c.Stop()
	if err != nil {
		slog.Error("Failed to stop container: %s", "error", err)
		os.Exit(1)
	}

	push := c.GetBuild().Custom.Bool("push", true)
	if !push {
		slog.Info("Skipping image push", "image", c.Image, "tag", c.ImageTag)
		return "", nil
	}
	imageUri := utils.ImageURI(c.GetBuild().Registry, c.Image, c.ImageTag)
	err = c.Push(imageId, imageUri)
	if err != nil {
		slog.Error("Failed to push image: %s", "error", err)
		os.Exit(1)
	}

	return c.ID, err
}

func (c *MavenContainer) Run() (string, error) {
	err := c.Pull()
	if err != nil {
		slog.Error("Failed to pull base images: %s", "error", err)
		return "", err
	}

	err = c.BuildMavenImage()
	if err != nil {
		slog.Error("Failed to build go image: %s", "error", err)
		return "", err
	}

	err = c.Build()
	slog.Info("Container created", "containerId", c.ID)
	if err != nil {
		slog.Error("Failed to create container: %s", "error", err)
		return "", err
	}
	return c.ID, nil
}
