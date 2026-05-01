package stdouttest_test

import (
	"fmt"
	"os"
	"testing"

	"go.rgst.io/stencil/v2/internal/testing/stdouttest"
	"gotest.tools/v3/assert"
)

func TestDefaultCapturesStdoutAndStderr(t *testing.T) {
	out := stdouttest.Run(t, func() {
		fmt.Print("stdout")
		fmt.Fprint(os.Stderr, "stderr")
	})
	assert.Equal(t, string(out), "stdoutstderr")
}

func TestDisableStderr(t *testing.T) {
	out := stdouttest.Run(t, func() {
		fmt.Print("stdout")
		fmt.Fprint(os.Stderr, "stderr")
	}, &stdouttest.RunOptions{Stdout: true})
	assert.Equal(t, string(out), "stdout")
}

func TestDisableStdout(t *testing.T) {
	out := stdouttest.Run(t, func() {
		fmt.Print("stdout")
		fmt.Fprint(os.Stderr, "stderr")
	}, &stdouttest.RunOptions{Stderr: true})
	assert.Equal(t, string(out), "stderr")
}
