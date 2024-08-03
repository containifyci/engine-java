package maven

import (
	"log/slog"
	"os"

	"github.com/containifyci/engine-ci/pkg/build"
	"github.com/containifyci/engine-ci/pkg/container"
)

const (
	DEFAULT_MAVEN_VERSION = "v17"
)

func Steps(arg *container.Build) []build.Build {
	var from string
	if v, ok := arg.Custom["from"]; ok {
		slog.Info("Using custom build", "from", v[0])
		from = v[0]
	}

	if from == "" {
		slog.Info("Using default build", "from", DEFAULT_MAVEN_VERSION)
		from = DEFAULT_MAVEN_VERSION
	}

	var steps []build.Build

	switch from {
	case "v17", "v21":
		_from := from[1:]
		steps = append(steps, New(_from), NewProd(_from))
	default:
		slog.Error("Unsupported maven from image", "from", from)
		os.Exit(1)
	}
	return steps
}
