package maven

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleScript(t *testing.T) {
	bs := NewBuildScript(false, ".", "localhost")
	script := Script(bs)

	assert.Equal(t, "#!/bin/sh\nset -xe\ncd .\nmvn --batch-mode package\n", script)
}

func TestVerboseScript(t *testing.T) {
	bs := NewBuildScript(true, "java", "localhost")
	script := Script(bs)

	assert.Equal(t, "#!/bin/sh\nset -xe\ncd java\nmvn --batch-mode package -X\n", script)
}
