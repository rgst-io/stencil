package dotnotation

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"gotest.tools/v3/assert"
)

func TestPrefixTree(t *testing.T) {
	pt := NewPrefixTree()
	pt.Add("hello.my.name.is.jared")
	pt.Add("hello.world")

	spew.Dump(pt)

	// shallow partial
	assert.Equal(t, pt.Has("hello"), true)

	// second path
	assert.Equal(t, pt.Has("hello.world"), true)

	// deep partial
	assert.Equal(t, pt.Has("hello.my.name"), true)

	// no match
	assert.Equal(t, pt.Has("hello.my.not"), false)
}
