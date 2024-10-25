package codegen

import (
	"context"
	"os"
	"testing"
	"time"

	"go.rgst.io/stencil/internal/modules/modulestest"
	"go.rgst.io/stencil/internal/testing/testmemfs"
	"go.rgst.io/stencil/pkg/slogext"
	"gotest.tools/v3/assert"
)

func fakeBlocksTemplate() *Template {
	return &Template{}
}

func TestParseBlocks(t *testing.T) {
	blocks, err := parseBlocks("testdata/blocks-test.txt", fakeBlocksTemplate())
	assert.NilError(t, err, "expected parseBlocks() not to fail")
	assert.Equal(t, blocks["helloWorld"].Contents, "Hello, world!", "expected parseBlocks() to parse basic block")
	assert.Equal(t, blocks["e2e"].Contents, "content", "expected parseBlocks() to parse e2e block")
}

func TestDanglingBlock(t *testing.T) {
	_, err := parseBlocks("testdata/danglingblock-test.txt", fakeBlocksTemplate())
	assert.Error(t, err, "found dangling Block (dangles) in testdata/danglingblock-test.txt", "expected parseBlocks() to fail")
}

func TestDanglingEndBlock(t *testing.T) {
	_, err := parseBlocks("testdata/danglingendblock-test.txt", fakeBlocksTemplate())
	assert.Error(t, err,
		"invalid EndBlock when not inside of a block, at testdata/danglingendblock-test.txt:8",
		"expected parseBlocks() to fail")
}

func TestBlockInsideBlock(t *testing.T) {
	_, err := parseBlocks("testdata/blockinsideblock-test.txt", fakeBlocksTemplate())
	assert.Error(t, err,
		"invalid Block when already inside of a block, at testdata/blockinsideblock-test.txt:3",
		"expected parseBlocks() to fail")
}

func TestWrongEndBlock(t *testing.T) {
	_, err := parseBlocks("testdata/wrongendblock-test.txt", fakeBlocksTemplate())
	assert.Error(t, err,
		"invalid EndBlock, found EndBlock with name \"wrongend\" while inside of block with name \"helloWorld\", at testdata/wrongendblock-test.txt:3", //nolint:lll
		"expected parseBlocks() to fail")
}

func TestParseV2Blocks(t *testing.T) {
	blocks, err := parseBlocks("testdata/v2blocks-test.txt", fakeBlocksTemplate())
	assert.NilError(t, err, "expected parseBlocks() not to fail")
	assert.Equal(t, blocks["helloWorld"].Contents, "Hello, world!", "expected parseBlocks() to parse basic block")
}

func TestV2BlocksErrors(t *testing.T) {
	_, err := parseBlocks("testdata/v2blocks-invalid.txt", fakeBlocksTemplate())
	if err == nil {
		t.Fatal("expected parseBlocks() to fail")
	}
}

func TestBasicAdopt(t *testing.T) {
	blocks := adoptTestHelper(t, "testdata/adopt/adopt1.tpl", "testdata/adopt/adopt1.yaml")
	expb := blockInfo{
		Name:      "version",
		StartLine: 1,
		EndLine:   3,
		Contents:  "  version: xyz",
	}
	assert.Equal(t, *blocks["version"], expb, "expected parseBlocks() to parse version block")
}

func TestAdoptWithMultiplePres(t *testing.T) {
	blocks := adoptTestHelper(t, "testdata/adopt/adopt2.tpl", "testdata/adopt/adopt2.yaml")
	expb := blockInfo{
		Name:      "version",
		StartLine: 1,
		EndLine:   3,
		Contents:  "  version: xyz",
	}
	assert.Equal(t, *blocks["version"], expb, "expected parseBlocks() to parse version block")
}

func TestAdoptWithMultiplePresUseNext(t *testing.T) {
	blocks := adoptTestHelper(t, "testdata/adopt/adopt3.tpl", "testdata/adopt/adopt3.yaml")
	expb := blockInfo{
		Name:      "version",
		StartLine: 5,
		EndLine:   7,
		Contents:  "  version: abc",
	}
	assert.Equal(t, *blocks["version"], expb, "expected parseBlocks() to parse version block")
}

func adoptTestHelper(t *testing.T, templateFile, targetFile string) map[string]*blockInfo {
	fs, err := testmemfs.WithManifest("name: testing\n")
	assert.NilError(t, err, "failed to testmemfs.WithManifest")
	m, err := modulestest.NewWithFS(context.Background(), "testing", fs)
	log := slogext.NewTestLogger(t)
	assert.NilError(t, err, "failed to NewWithFS")

	conts, err := os.ReadFile(templateFile)
	assert.NilError(t, err, "failed to read templateFile")
	tpl, err := NewTemplate(m, templateFile, 0o644, time.Now(), conts, log, true)
	assert.NilError(t, err, "failed to NewTemplate")

	blocks, err := parseBlocks(targetFile, tpl)
	assert.NilError(t, err, "expected parseBlocks() not to fail")
	return blocks
}
