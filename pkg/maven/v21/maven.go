package v21

import (
	"github.com/containifyci/engine-ci/pkg/build"
	"github.com/containifyci/java/pkg/maven"
)

func New() *maven.MavenContainer {
	return maven.New("21")
}

func NewProd() build.Build {
	return maven.NewProd("21")
}
