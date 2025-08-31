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
	assert.Equal(t, []string{"containifyci/maven-3-eclipse-temurin-17-alpine:cdbe73779492603b08a3e880bf25754e3a8e865811c51c0b45e2c5edfc5a8476", "tomcat:latest", "tomcat:latest"}, images)
}

func TestStepsv21(t *testing.T) {
	arg := InitTest(t)
	arg.Custom = map[string][]string{"from": {"v21"}}

	steps := Steps(arg)
	assert.Len(t, steps, 2)
	names := []string{steps[0].Name(), steps[1].Name()}
	images := append(steps[0].Images(), steps[1].Images()...)

	assert.Equal(t, []string{"maven", "maven-prod"}, names)
	assert.Equal(t, []string{"containifyci/maven-3-eclipse-temurin-21-alpine:889a67acb64eb60e00023503d6337c5455d2b4086d194103bd83a801dc39ec52", "tomcat:latest", "tomcat:latest"}, images)
}

func TestStepsDefault(t *testing.T) {
	arg := InitTest(t)

	steps := Steps(arg)
	assert.Len(t, steps, 2)
	names := []string{steps[0].Name(), steps[1].Name()}
	images := append(steps[0].Images(), steps[1].Images()...)

	assert.Equal(t, []string{"maven", "maven-prod"}, names)
	assert.Equal(t, []string{"containifyci/maven-3-eclipse-temurin-17-alpine:cdbe73779492603b08a3e880bf25754e3a8e865811c51c0b45e2c5edfc5a8476", "tomcat:latest", "tomcat:latest"}, images)
}

func TestStepsUnknown(t *testing.T) {
	t.SkipNow()
	arg := InitTest(t)
	arg.Custom = map[string][]string{"from": {"v24"}}

	Steps(arg)
}
