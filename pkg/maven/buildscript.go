package maven

import "fmt"

type Image string

type BuildScript struct {
	Verbose bool
	Folder  string
	Host    string
}

func NewBuildScript(verbose bool, folder, host string) *BuildScript {
	return &BuildScript{
		Verbose: verbose,
		Folder:  folder,
		Host:    host,
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
cd %s
mvn --batch-mode package
`, bs.Folder)
}

func verboseScript(bs *BuildScript) string {
	return fmt.Sprintf(`#!/bin/sh
set -xe
cd %s
mvn --batch-mode package -X
`, bs.Folder)
}
