package maven

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStepsv17(t *testing.T) {
	arg := InitTest(t)
	arg.Custom = map[string][]string{"from": {"v17"}}

	steps := Steps(arg)
	assert.Len(t, steps, 2)
	names := []string{steps[0].Name(), steps[1].Name()}
	images := append(steps[0].Images(), steps[1].Images()...)

	assert.Equal(t, []string{"maven", "maven-prod"}, names)
	assert.Equal(t, []string{"containifyci/maven-3-eclipse-temurin-17-alpine:214f702cd3dee211717ca17936aca88df6913e1db40f822ef9918e4369ea927d", "registry.access.redhat.com/ubi8/openjdk-17:latest", "registry.access.redhat.com/ubi8/openjdk-17:latest"}, images)
}

func TestStepsv21(t *testing.T) {
	arg := InitTest(t)
	arg.Custom = map[string][]string{"from": {"v21"}}

	steps := Steps(arg)
	assert.Len(t, steps, 2)
	names := []string{steps[0].Name(), steps[1].Name()}
	images := append(steps[0].Images(), steps[1].Images()...)

	assert.Equal(t, []string{"maven", "maven-prod"}, names)
	assert.Equal(t, []string{"containifyci/maven-3-eclipse-temurin-21-alpine:98108b284d14233f76bb9ce91e7932734b135f899654e48306ab07e4ca3e35d8", "registry.access.redhat.com/ubi8/openjdk-21:latest", "registry.access.redhat.com/ubi8/openjdk-21:latest"}, images)
}

func TestStepsDefault(t *testing.T) {
	arg := InitTest(t)

	steps := Steps(arg)
	assert.Len(t, steps, 2)
	names := []string{steps[0].Name(), steps[1].Name()}
	images := append(steps[0].Images(), steps[1].Images()...)

	assert.Equal(t, []string{"maven", "maven-prod"}, names)
	assert.Equal(t, []string{"containifyci/maven-3-eclipse-temurin-17-alpine:214f702cd3dee211717ca17936aca88df6913e1db40f822ef9918e4369ea927d", "registry.access.redhat.com/ubi8/openjdk-17:latest", "registry.access.redhat.com/ubi8/openjdk-17:latest"}, images)
}

func TestStepsUnknown(t *testing.T) {
	t.SkipNow()
	arg := InitTest(t)
	arg.Custom = map[string][]string{"from": {"v24"}}

	Steps(arg)
}
