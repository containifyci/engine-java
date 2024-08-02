package v17

import (
	"github.com/containifyci/engine-ci/pkg/build"
	"github.com/containifyci/java/pkg/maven"
)

func New() *maven.MavenContainer {
	return maven.New("17")
}

func NewProd() build.Build {
	return maven.NewProd("17")
}
