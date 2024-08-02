package maven

import "fmt"

type Image string

type BuildScript struct {
	Verbose bool
	host string
}

func NewBuildScript(verbose bool, host string) *BuildScript {
	return &BuildScript{
		Verbose: verbose,
		host: host,
	}
}

func Script(bs *BuildScript) string {
	if bs.Verbose {
		return verboseScript(bs)
	}
	return simpleScript(bs)
}

func simpleScript(bs *BuildScript) string {
	return fmt.Sprintf(`#!/bin/sh
set -xe
./mvnw --batch-mode package
`)
}

func verboseScript(bs *BuildScript) string {
	return fmt.Sprintf(`#!/bin/sh
set -xe
./mvnw --batch-mode package -X
`)
}
